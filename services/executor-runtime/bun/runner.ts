/**
 * runeforge bun executor runner
 *
 * Persistent HTTP server that accepts POST /run requests, writes snippet code
 * to a temporary file, executes it as a Bun subprocess with the input payload
 * available as SNIPPET_INPUT, captures stdout/stderr, enforces a timeout, and
 * returns a structured result.
 *
 * Expected request body:
 *   { code: string, input: string, timeout_ms: number, max_memory_mb: number }
 *
 * Response body:
 *   { output: string, stderr: string, duration_ms: number, peak_memory_mb: number, exit_code: number, error: string }
 *
 * Snippet convention:
 *   The snippet file must export a default async function `handler` that
 *   accepts the parsed input and returns the output. The runner wraps execution
 *   and prints the JSON result to stdout.
 *
 * Example snippet:
 *   export default async function handler(input: { name: string }) {
 *     return { greeting: `Hello, ${input.name}!` };
 *   }
 */

import { randomBytes } from "node:crypto";
import { mkdtemp, writeFile, unlink, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

const PORT = 8080;

interface RunRequest {
  code: string;
  input: string;
  timeout_ms: number;
  max_memory_mb: number;
}

interface RunResult {
  output: string;
  stderr: string;
  duration_ms: number;
  peak_memory_mb: number;
  exit_code: number;
  error: string;
}

/**
 * Write the wrapper harness that imports the snippet and runs handler().
 * The snippet's output is printed as JSON to stdout.
 */
function buildHarnessCode(snippetPath: string, input: string): string {
  // Escape the input and path for safe embedding in JS source.
  const safeInput = JSON.stringify(input);
  const safePath = JSON.stringify(snippetPath);

  return `
import snippetModule from ${safePath};

const handler = typeof snippetModule === 'function'
  ? snippetModule
  : (snippetModule.default ?? snippetModule.handler);

if (typeof handler !== 'function') {
  console.error('Snippet must export a default function or a function named "handler"');
  process.exit(1);
}

const rawInput = ${safeInput};
let parsedInput;
try {
  parsedInput = JSON.parse(rawInput);
} catch (e) {
  parsedInput = rawInput;
}

(async () => {
  try {
    const result = await handler(parsedInput);
    const output = typeof result === 'string' ? result : JSON.stringify(result);
    process.stdout.write(output + '\\n');
    process.exit(0);
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    process.stderr.write('handler error: ' + msg + '\\n');
    process.exit(1);
  }
})();
`.trimStart();
}

async function runSnippet(req: RunRequest): Promise<RunResult> {
  const id = randomBytes(8).toString("hex");
  const workDir = await mkdtemp(join(tmpdir(), `rune_${id}_`));
  const snippetPath = join(workDir, "snippet.ts");
  const harnessPath = join(workDir, "harness.ts");

  try {
    await writeFile(snippetPath, req.code, "utf8");
    const harnessCode = buildHarnessCode(snippetPath, req.input);
    await writeFile(harnessPath, harnessCode, "utf8");

    const timeoutMs = req.timeout_ms > 0 ? req.timeout_ms : 30_000;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), timeoutMs);

    const start = Date.now();
    let proc: ReturnType<typeof Bun.spawn> | null = null;

    try {
      proc = Bun.spawn(["bun", "run", harnessPath], {
        env: {
          ...process.env,
          SNIPPET_INPUT: req.input,
        },
        stdout: "pipe",
        stderr: "pipe",
        signal: controller.signal,
      });

      const stdoutChunks: Uint8Array[] = [];
      const stderrChunks: Uint8Array[] = [];

      // Collect stdout.
      const collectStdout = async () => {
        if (!proc!.stdout) return;
        const reader = proc!.stdout.getReader();
        try {
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            if (value) stdoutChunks.push(value);
          }
        } catch {
          // Stream ended or process killed.
        }
      };

      // Collect stderr.
      const collectStderr = async () => {
        if (!proc!.stderr) return;
        const reader = proc!.stderr.getReader();
        try {
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            if (value) stderrChunks.push(value);
          }
        } catch {
          // Stream ended or process killed.
        }
      };

      await Promise.all([
        collectStdout(),
        collectStderr(),
        proc.exited,
      ]);

      clearTimeout(timer);
      const durationMs = Date.now() - start;
      const exitCode = proc.exitCode ?? -1;

      const decoder = new TextDecoder();
      const stdout = decoder.decode(
        Buffer.concat(stdoutChunks.map((c) => Buffer.from(c)))
      ).trim();
      const stderr = decoder.decode(
        Buffer.concat(stderrChunks.map((c) => Buffer.from(c)))
      ).trim();

      return {
        output: stdout,
        stderr,
        duration_ms: durationMs,
        peak_memory_mb: 0, // not available without cgroups
        exit_code: exitCode,
        error: exitCode !== 0 ? "non-zero exit" : "",
      };
    } catch (err: any) {
      clearTimeout(timer);
      const durationMs = Date.now() - start;

      if (err?.name === "AbortError" || controller.signal.aborted) {
        // Kill the process if it's still running.
        try {
          proc?.kill();
        } catch {}
        return {
          output: "",
          stderr: "execution timed out",
          duration_ms: durationMs,
          peak_memory_mb: 0,
          exit_code: -1,
          error: "timeout",
        };
      }

      return {
        output: "",
        stderr: String(err),
        duration_ms: durationMs,
        peak_memory_mb: 0,
        exit_code: -1,
        error: String(err),
      };
    }
  } finally {
    // Clean up temp directory.
    try {
      await rm(workDir, { recursive: true, force: true });
    } catch {}
  }
}

const server = Bun.serve({
  port: PORT,
  async fetch(req: Request) {
    const url = new URL(req.url);

    // Health check.
    if (req.method === "GET" && url.pathname === "/healthz") {
      return new Response(JSON.stringify({ status: "ok" }), {
        headers: { "Content-Type": "application/json" },
      });
    }

    if (req.method !== "POST" || url.pathname !== "/run") {
      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    }

    let body: RunRequest;
    try {
      body = (await req.json()) as RunRequest;
    } catch {
      return new Response(JSON.stringify({ error: "invalid JSON body" }), {
        status: 400,
        headers: { "Content-Type": "application/json" },
      });
    }

    if (!body.code) {
      return new Response(JSON.stringify({ error: "code is required" }), {
        status: 400,
        headers: { "Content-Type": "application/json" },
      });
    }

    const result = await runSnippet(body);
    return new Response(JSON.stringify(result), {
      headers: { "Content-Type": "application/json" },
    });
  },
});

console.log(`bun executor listening on :${PORT}`);
