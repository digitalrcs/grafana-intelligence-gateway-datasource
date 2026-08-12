#!/usr/bin/env python3
"""Create a Grafana plugin ZIP while preserving executable permissions."""

from pathlib import Path
import zipfile


PLUGIN_ID = "digitalrcs-intelligencegateway-datasource"
ROOT = Path(__file__).resolve().parents[1]
DIST = ROOT / "dist"
OUTPUT = ROOT / f"{PLUGIN_ID}-1.0.0.zip"


def main() -> None:
    if not (DIST / "plugin.json").is_file():
        raise SystemExit("dist/plugin.json is missing; run npm run build first")
    with zipfile.ZipFile(OUTPUT, "w", zipfile.ZIP_DEFLATED) as archive:
        for source in sorted(path for path in DIST.rglob("*") if path.is_file()):
            relative = Path(PLUGIN_ID) / source.relative_to(DIST)
            info = zipfile.ZipInfo.from_file(source, relative.as_posix())
            if source.name.startswith("gpx_"):
                info.create_system = 3
                info.external_attr = (0o100755 << 16) | 0x20
            archive.writestr(info, source.read_bytes(), compress_type=zipfile.ZIP_DEFLATED)
    print(OUTPUT)


if __name__ == "__main__":
    main()
