"""Linux cgroup v2 helpers for per-snippet CPU limits. No-op on non-Linux."""

from __future__ import annotations

import platform
from pathlib import Path

_CGROUP_ROOT = Path("/sys/fs/cgroup")


def _linux_cgroup_available() -> bool:
    return platform.system() == "Linux" and _CGROUP_ROOT.is_dir()


def apply_limits(pid: int, max_memory_mb: int, max_cpu_percent: int) -> Path | None:
    """Apply memory.max and cpu.max in a single ephemeral cgroup."""
    if not _linux_cgroup_available():
        return None
    if max_memory_mb <= 0 and max_cpu_percent <= 0:
        return None

    cg = _CGROUP_ROOT / f"velane-{pid}"
    try:
        cg.mkdir(exist_ok=True)
        if max_memory_mb > 0:
            (cg / "memory.max").write_text(str(max_memory_mb * 1024 * 1024))
        if max_cpu_percent > 0 and max_cpu_percent <= 100:
            period = 100000
            quota = max(1, int(period * max_cpu_percent / 100))
            (cg / "cpu.max").write_text(f"{quota} {period}")
        (cg / "cgroup.procs").write_text(str(pid))
        return cg
    except OSError:
        return None


def apply_cpu_limit(pid: int, max_cpu_percent: int) -> Path | None:
    return apply_limits(pid, 0, max_cpu_percent)


def apply_memory_limit_cgroup(pid: int, max_memory_mb: int) -> Path | None:
    return apply_limits(pid, max_memory_mb, 0)


def read_peak_memory_mb(cgroup: Path | None) -> int:
    if cgroup is None:
        return 0
    peak = cgroup / "memory.peak"
    try:
        raw = peak.read_text().strip()
        if not raw or raw == "max":
            return 0
        return int(int(raw) / (1024 * 1024))
    except OSError:
        return 0


def cleanup_cgroup(cgroup: Path | None) -> None:
    if cgroup is None:
        return
    try:
        cgroup.rmdir()
    except OSError:
        pass


def is_oom_exit(exit_code: int | None) -> bool:
    # SIGKILL (-9) often indicates cgroup OOM killer.
    return exit_code == -9 or exit_code == 137
