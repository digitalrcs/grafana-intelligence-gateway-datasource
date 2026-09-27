#!/usr/bin/env python3
"""Validate and package a built Grafana plugin with executable backend files."""

import json
from pathlib import Path
from pathlib import PurePosixPath
import re
import zipfile


ROOT = Path(__file__).resolve().parents[1]
REQUIRED_FILES = ("plugin.json", "README.md", "module.js", "LICENSE", "CHANGELOG.md")
PLUGIN_ID = re.compile(r"[a-z0-9]+(?:-[a-z0-9]+)?-(?:app|datasource|panel)")
NUMBER = r"(?:0|[1-9][0-9]*)"
PRERELEASE_ID = rf"(?:{NUMBER}|[0-9]*[A-Za-z-][0-9A-Za-z-]*)"
VERSION = re.compile(
    rf"{NUMBER}\.{NUMBER}\.{NUMBER}"
    rf"(?:-{PRERELEASE_ID}(?:\.{PRERELEASE_ID})*)?"
    r"(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?"
)


def asset_path(value: object) -> PurePosixPath:
    """Accept only portable paths inside the archive's plugin directory."""
    if not isinstance(value, str) or not value or "\\" in value or ":" in value:
        raise ValueError(f"invalid packaged asset path: {value!r}")
    path = PurePosixPath(value)
    if path.is_absolute() or ".." in path.parts or str(path) == ".":
        raise ValueError(f"invalid packaged asset path: {value!r}")
    return path


def package_plugin(dist: Path, output_dir: Path) -> Path:
    """Fail validation before opening the output ZIP, then package all dist files."""
    for name in REQUIRED_FILES:
        if not (dist / name).is_file():
            raise ValueError(f"dist/{name} is missing; build the complete plugin first")

    metadata = json.loads((dist / "plugin.json").read_text(encoding="utf-8"))
    if not isinstance(metadata, dict):
        raise ValueError("dist/plugin.json must contain an object")
    plugin_id = metadata.get("id")
    if not isinstance(plugin_id, str) or not PLUGIN_ID.fullmatch(plugin_id):
        raise ValueError("dist/plugin.json has an invalid plugin id")
    if metadata.get("type") != plugin_id.rsplit("-", 1)[-1]:
        raise ValueError("dist/plugin.json type must match the plugin id suffix")
    info = metadata.get("info")
    if not isinstance(info, dict):
        raise ValueError("dist/plugin.json info must contain an object")
    version = info.get("version")
    if not isinstance(version, str) or not VERSION.fullmatch(version):
        raise ValueError("dist/plugin.json info.version must be a built semantic version")

    logos = info.get("logos", {})
    screenshots = info.get("screenshots", [])
    if not isinstance(logos, dict) or not isinstance(screenshots, list):
        raise ValueError("dist/plugin.json logos or screenshots metadata is invalid")
    references = list(logos.values())
    for screenshot in screenshots:
        if not isinstance(screenshot, dict):
            raise ValueError("dist/plugin.json screenshot metadata must contain an object")
        references.append(screenshot.get("path"))
    for reference in references:
        relative = asset_path(reference)
        if not (dist / relative).is_file():
            raise ValueError(f"referenced asset dist/{relative} is missing")

    # Reject links rather than accidentally including files from outside dist.
    paths = sorted(dist.rglob("*"))
    if any(path.is_symlink() for path in paths):
        raise ValueError("dist must not contain symbolic links")
    files = [path for path in paths if path.is_file()]
    executables = set()
    if metadata.get("backend"):
        executable = metadata.get("executable")
        if not isinstance(executable, str) or not re.fullmatch(r"[A-Za-z0-9_-]+", executable):
            raise ValueError("backend plugin must declare a valid executable name")
        binary_name = re.compile(
            rf"{re.escape(executable)}_"
            r"(?:linux_(?:amd64|arm|arm64)|darwin_(?:amd64|arm64)|windows_(?:amd64|arm64)\.exe)"
        )
        executables = {path for path in files if path.parent == dist and binary_name.fullmatch(path.name)}
        if not executables:
            raise ValueError(f"dist is missing a built {executable}_<os>_<arch> backend executable")

    output = output_dir / f"{plugin_id}-{version}.zip"
    with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
        for source in files:
            relative = PurePosixPath(plugin_id) / source.relative_to(dist).as_posix()
            entry = zipfile.ZipInfo.from_file(source, relative.as_posix())
            entry.create_system = 3
            mode = 0o100755 if source in executables else 0o100644
            entry.external_attr = (mode << 16) | 0x20
            archive.writestr(entry, source.read_bytes(), compress_type=zipfile.ZIP_DEFLATED)
    return output


def main() -> None:
    try:
        print(package_plugin(ROOT / "dist", ROOT))
    except (OSError, ValueError) as error:
        raise SystemExit(str(error)) from error


if __name__ == "__main__":
    main()
