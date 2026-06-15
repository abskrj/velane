/**
 * Linux cgroup v2 helpers for per-snippet CPU limits. No-op on non-Linux.
 */

import { existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const CGROUP_ROOT = "/sys/fs/cgroup";

function linuxCgroupAvailable(): boolean {
  return process.platform === "linux" && existsSync(CGROUP_ROOT);
}

export function applyLimits(
  pid: number,
  maxMemoryMb: number,
  maxCpuPercent: number,
): string | null {
  if (!linuxCgroupAvailable()) return null;
  if (maxMemoryMb <= 0 && maxCpuPercent <= 0) return null;

  const cg = join(CGROUP_ROOT, `velane-${pid}`);
  try {
    mkdirSync(cg, { recursive: true });
    if (maxMemoryMb > 0) {
      writeFileSync(join(cg, "memory.max"), String(maxMemoryMb * 1024 * 1024));
    }
    if (maxCpuPercent > 0 && maxCpuPercent <= 100) {
      const period = 100000;
      const quota = Math.max(1, Math.floor((period * maxCpuPercent) / 100));
      writeFileSync(join(cg, "cpu.max"), `${quota} ${period}`);
    }
    writeFileSync(join(cg, "cgroup.procs"), String(pid));
    return cg;
  } catch {
    return null;
  }
}

export function applyCpuLimit(pid: number, maxCpuPercent: number): string | null {
  return applyLimits(pid, 0, maxCpuPercent);
}

export function applyMemoryLimitCgroup(pid: number, maxMemoryMb: number): string | null {
  return applyLimits(pid, maxMemoryMb, 0);
}

export function readPeakMemoryMb(cgroup: string | null): number {
  if (!cgroup) return 0;
  try {
    const raw = readFileSync(join(cgroup, "memory.peak"), "utf8").trim();
    if (!raw || raw === "max") return 0;
    return Math.floor(Number(raw) / (1024 * 1024));
  } catch {
    return 0;
  }
}

export function cleanupCgroup(cgroup: string | null): void {
  if (!cgroup) return;
  try {
    rmSync(cgroup, { recursive: true, force: true });
  } catch {
    // ignore
  }
}

export function isOomExit(exitCode: number): boolean {
  return exitCode === -9 || exitCode === 137;
}
