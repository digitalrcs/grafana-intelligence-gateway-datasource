"""Standard-library tests for the local Grafana plugin archive builder."""

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
import zipfile


SPEC = importlib.util.spec_from_file_location("package_plugin", Path(__file__).with_name("package-plugin.py"))
packager = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(packager)


class PackagePluginTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.dist = self.root / "dist"
        self.dist.mkdir()
        self.plugin_id = "digitalrcs-intelligencegateway-datasource"
        self.binary = "gpx_intelligencegateway_linux_amd64"
        self.metadata = {
            "id": self.plugin_id,
            "type": "datasource",
            "backend": True,
            "executable": "gpx_intelligencegateway",
            "info": {
                "version": "2.3.4",
                "logos": {"small": "img/logo.svg", "large": "img/logo.svg"},
                "screenshots": [{"name": "Example", "path": "img/example.png"}],
            },
        }
        for name in packager.REQUIRED_FILES + (self.binary, "img/logo.svg", "img/example.png"):
            path = self.dist / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(b"fixture")
        self.write_metadata()

    def write_metadata(self):
        (self.dist / "plugin.json").write_text(json.dumps(self.metadata), encoding="utf-8")

    def assert_rejected(self, message):
        with self.assertRaisesRegex(ValueError, message):
            packager.package_plugin(self.dist, self.root)
        self.assertEqual(list(self.root.glob("*.zip")), [], "validation must precede archive creation")

    def test_version_and_id_come_from_built_manifest(self):
        self.metadata["id"] = "example-custom-datasource"
        self.metadata["info"]["version"] = "2.10.3-rc.2+build.04"
        self.write_metadata()
        output = packager.package_plugin(self.dist, self.root)
        self.assertEqual(output.name, "example-custom-datasource-2.10.3-rc.2+build.04.zip")
        with zipfile.ZipFile(output) as archive:
            self.assertTrue(all(name.startswith("example-custom-datasource/") for name in archive.namelist()))

    def test_archive_has_plugin_root_and_unix_executable_permissions(self):
        output = packager.package_plugin(self.dist, self.root)
        with zipfile.ZipFile(output) as archive:
            names = archive.namelist()
            self.assertTrue(all(name.startswith(self.plugin_id + "/") for name in names))
            self.assertTrue(all("\\" not in name for name in names))
            self.assertEqual(archive.read(f"{self.plugin_id}/README.md"), b"fixture")
            binary = archive.getinfo(f"{self.plugin_id}/{self.binary}")
            self.assertEqual(binary.create_system, 3)
            self.assertEqual(binary.external_attr >> 16, 0o100755)
            readme = archive.getinfo(f"{self.plugin_id}/README.md")
            self.assertEqual(readme.external_attr >> 16, 0o100644)

    def test_each_mandatory_file_is_required(self):
        for name in packager.REQUIRED_FILES:
            with self.subTest(name=name):
                path = self.dist / name
                original = path.read_bytes()
                path.unlink()
                self.assert_rejected("is missing")
                path.write_bytes(original)

    def test_each_referenced_asset_is_required(self):
        for name in ("img/logo.svg", "img/example.png"):
            with self.subTest(name=name):
                path = self.dist / name
                path.unlink()
                self.assert_rejected("referenced asset .* is missing")
                path.write_bytes(b"fixture")

    def test_backend_binary_is_required(self):
        (self.dist / self.binary).unlink()
        (self.dist / "gpx_other_linux_amd64").write_bytes(b"wrong executable")
        self.assert_rejected("missing a built .* backend executable")

    def test_backend_binary_must_be_at_plugin_root(self):
        (self.dist / self.binary).rename(self.dist / "img" / self.binary)
        self.assert_rejected("missing a built .* backend executable")

    def test_single_windows_binary_is_sufficient(self):
        windows = "gpx_intelligencegateway_windows_amd64.exe"
        (self.dist / self.binary).rename(self.dist / windows)
        output = packager.package_plugin(self.dist, self.root)
        with zipfile.ZipFile(output) as archive:
            self.assertEqual(archive.getinfo(f"{self.plugin_id}/{windows}").external_attr >> 16, 0o100755)

    def test_frontend_only_plugin_does_not_require_backend(self):
        self.metadata["backend"] = False
        self.metadata.pop("executable")
        self.write_metadata()
        (self.dist / self.binary).unlink()
        self.assertTrue(packager.package_plugin(self.dist, self.root).is_file())

    def test_invalid_plugin_ids_are_rejected(self):
        for plugin_id in (None, 123, "", "../plugin-datasource", "Bad-Name-datasource", "foo", "org-plugin-other"):
            with self.subTest(plugin_id=plugin_id):
                self.metadata["id"] = plugin_id
                self.write_metadata()
                self.assert_rejected("invalid plugin id")

    def test_plugin_type_must_match_identity(self):
        self.metadata["type"] = "panel"
        self.write_metadata()
        self.assert_rejected("type must match")

    def test_unbuilt_or_invalid_versions_are_rejected(self):
        for version in (None, 3, "", "%VERSION%", "v1.2.3", "1.2", "01.2.3", "1.2.3-01", "1.2.3/../../bad"):
            with self.subTest(version=version):
                self.metadata["info"]["version"] = version
                self.write_metadata()
                self.assert_rejected("built semantic version")

    def test_asset_paths_must_stay_inside_plugin(self):
        for path in ("../README.md", "/README.md", "C:/README.md", "img\\logo.svg", "", None):
            with self.subTest(path=path):
                self.metadata["info"]["logos"]["small"] = path
                self.write_metadata()
                self.assert_rejected("invalid packaged asset path")

    def test_backend_executable_name_is_validated(self):
        for name in (None, "", "../gpx_plugin", "gpx_plugin.exe", 12):
            with self.subTest(name=name):
                self.metadata["executable"] = name
                self.write_metadata()
                self.assert_rejected("valid executable name")

    def test_invalid_json_fails_before_archive_creation(self):
        (self.dist / "plugin.json").write_text("{", encoding="utf-8")
        self.assert_rejected("Expecting property name")


if __name__ == "__main__":
    unittest.main()
