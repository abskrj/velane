"""
runeforge python executor runner

Persistent FastAPI HTTP server that accepts POST /run requests, writes snippet
code to a temporary file, executes it as a Python subprocess with the input
payload available as SNIPPET_INPUT, captures stdout/stderr, enforces a timeout
via asyncio, and returns a structured result.

Expected request body:
    { code: str, input: str, timeout_ms: int, max_memory_mb: int }

Response body:
    { output: str, stderr: str, duration_ms: int, peak_memory_mb: int,
      exit_code: int, error: str }

Snippet convention:
    The snippet module should define Pydantic models Input and Output, plus an
    async function handler(req: Input) -> Output. The runner wraps execution
    using the harness below and prints the serialised result to stdout.

Example snippet:
    from pydantic import BaseModel

    class Input(BaseModel):
        name: str

    class Output(BaseModel):
        greeting: str

    async def handler(req: Input) -> Output:
        return Output(greeting=f"Hello, {req.name}!")
"""

from __future__ import annotations

import asyncio
import os
import secrets
import sys
import tempfile
import time
from pathlib import Path
from typing import Any

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel

app = FastAPI(title="runeforge-python-executor")

# ---------------------------------------------------------------------------
# Request / response schemas
# ---------------------------------------------------------------------------


class RunRequest(BaseModel):
    code: str
    input: str = "{}"
    timeout_ms: int = 30_000
    max_memory_mb: int = 128


class RunResult(BaseModel):
    output: str = ""
    stderr: str = ""
    duration_ms: int = 0
    peak_memory_mb: int = 0
    exit_code: int = 0
    error: str = ""


# ---------------------------------------------------------------------------
# Harness template
# The harness is written alongside the snippet in the same temp directory.
# It imports the snippet module, calls handler(), and prints the result as JSON.
# ---------------------------------------------------------------------------

HARNESS_TEMPLATE = '''\
import asyncio
import json
import os
import sys

# Add the temp directory to the import path so we can import the snippet.
sys.path.insert(0, {snippet_dir!r})

import snippet as _snippet

async def _main():
    raw = os.environ.get("SNIPPET_INPUT", "{{}}")

    # Try to call handler with the parsed input. If the snippet defines
    # Pydantic Input/Output models, validate through them; otherwise pass the
    # raw parsed value directly.
    handler = getattr(_snippet, "handler", None)
    if handler is None:
        print(json.dumps({{"error": "snippet has no handler function"}}))
        sys.exit(1)

    # Determine if the snippet uses Pydantic models.
    Input = getattr(_snippet, "Input", None)
    if Input is not None:
        try:
            req = Input.model_validate_json(raw)
        except Exception as e:
            print(json.dumps({{"error": f"input validation: {{e}}"}}))
            sys.exit(1)
    else:
        import json as _json
        try:
            req = _json.loads(raw)
        except Exception:
            req = raw

    try:
        result = await handler(req)
    except Exception as e:
        print(json.dumps({{"error": f"handler raised: {{e}}"}}))
        sys.exit(1)

    # Serialise the result.
    Output = getattr(_snippet, "Output", None)
    if Output is not None and hasattr(result, "model_dump_json"):
        sys.stdout.write(result.model_dump_json() + "\\n")
    elif isinstance(result, (dict, list)):
        sys.stdout.write(json.dumps(result) + "\\n")
    else:
        sys.stdout.write(json.dumps(result) + "\\n")

asyncio.run(_main())
'''

# ---------------------------------------------------------------------------
# Execution logic
# ---------------------------------------------------------------------------


async def run_snippet(req: RunRequest) -> RunResult:
    """Write snippet code to a temp dir, run it with a timeout, collect output."""

    timeout_sec = max(req.timeout_ms / 1000.0, 1.0)

    # Use a temporary directory so both snippet.py and harness.py coexist.
    with tempfile.TemporaryDirectory(prefix="rune_") as work_dir:
        snippet_path = Path(work_dir) / "snippet.py"
        harness_path = Path(work_dir) / "harness.py"

        snippet_path.write_text(req.code, encoding="utf-8")
        harness_path.write_text(
            HARNESS_TEMPLATE.format(snippet_dir=work_dir),
            encoding="utf-8",
        )

        env = os.environ.copy()
        env["SNIPPET_INPUT"] = req.input

        start = time.monotonic()

        try:
            proc = await asyncio.create_subprocess_exec(
                sys.executable,
                str(harness_path),
                env=env,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=work_dir,
            )

            try:
                stdout_bytes, stderr_bytes = await asyncio.wait_for(
                    proc.communicate(),
                    timeout=timeout_sec,
                )
            except asyncio.TimeoutError:
                try:
                    proc.kill()
                    await proc.communicate()
                except Exception:
                    pass
                duration_ms = int((time.monotonic() - start) * 1000)
                return RunResult(
                    stderr="execution timed out",
                    duration_ms=duration_ms,
                    exit_code=-1,
                    error="timeout",
                )

            duration_ms = int((time.monotonic() - start) * 1000)
            exit_code = proc.returncode if proc.returncode is not None else -1

            stdout = stdout_bytes.decode("utf-8", errors="replace").strip()
            stderr = stderr_bytes.decode("utf-8", errors="replace").strip()

            error_str = ""
            if exit_code != 0:
                error_str = "non-zero exit"

            return RunResult(
                output=stdout,
                stderr=stderr,
                duration_ms=duration_ms,
                exit_code=exit_code,
                error=error_str,
            )

        except Exception as exc:
            duration_ms = int((time.monotonic() - start) * 1000)
            return RunResult(
                stderr=str(exc),
                duration_ms=duration_ms,
                exit_code=-1,
                error=str(exc),
            )


# ---------------------------------------------------------------------------
# HTTP endpoints
# ---------------------------------------------------------------------------


@app.get("/healthz")
async def healthz():
    return {"status": "ok"}


@app.post("/run", response_model=RunResult)
async def run(req: RunRequest):
    result = await run_snippet(req)
    return result


# ---------------------------------------------------------------------------
# Entry point (also launched via uvicorn in the Dockerfile)
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8080)
