"""Exercise the installer offline with fixture releases and real SHA-256 checks."""
from pathlib import Path
import hashlib
import io
import os
import subprocess
import tarfile
import tempfile
import unittest

INSTALLER = Path(__file__).resolve().parents[1] / "install.sh"


class InstallerTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="typeit-installer-test-")
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.fakebin = self.root / "tools"
        self.fakebin.mkdir()
        self.home = self.root / "home"
        self.home.mkdir()
        self.release = self.root / "release"
        self.release.mkdir()
        self.env = {**os.environ, "HOME": str(self.home), "PATH": str(self.fakebin) + os.pathsep + os.environ["PATH"],
                    "FIXTURE_RELEASE": str(self.release), "FAKE_OS": "Linux", "FAKE_ARCH": "x86_64"}
        for name in ["TYPEIT_INSTALL_DIR", "XDG_DATA_HOME"]:
            self.env.pop(name, None)
        self.tool("uname", '#!/bin/sh\ncase "$1" in -s) echo "$FAKE_OS";; -m) echo "$FAKE_ARCH";; esac\n')
        self.tool("curl", '''#!/usr/bin/env python3
from pathlib import Path
import os,shutil,sys
args=sys.argv[1:]
url=next(x for x in args if x.startswith("https://"))
if not url.startswith("https://github.com/michaelmjhhhh/typeit/releases/"):sys.exit(99)
root=Path(os.environ["FIXTURE_RELEASE"])
with (root/"requests").open("a") as log:log.write(url+"\\n")
if os.environ.get("FAIL_DOWNLOAD"):sys.exit(22)
shutil.copyfile(root/url.rsplit("/",1)[1],args[args.index("-o")+1])
''')
        self.prepare("linux", "amd64")

    def tool(self, name, text):
        path = self.fakebin / name
        path.write_text(text)
        path.chmod(0o755)

    def prepare(self, system, arch):
        asset = self.release / f"typeit_{system}_{arch}.tar.gz"
        with tarfile.open(asset, "w:gz") as archive:
            for name, content in {"typeit": b"#!/bin/sh\necho 'typeit fixture'\n", "LICENSE": b"license", "NOTICE": b"notice", "THIRD_PARTY_NOTICES.md": b"dependencies"}.items():
                info = tarfile.TarInfo(name)
                info.size = len(content)
                info.mode = 0o755 if name == "typeit" else 0o644
                archive.addfile(info, io.BytesIO(content))
        (self.release / "SHA256SUMS").write_text(f"{hashlib.sha256(asset.read_bytes()).hexdigest()}  {asset.name}\n")

    def run_installer(self, *args):
        return subprocess.run(["sh", str(INSTALLER), *args], env=self.env, text=True, capture_output=True)

    def test_default_install_and_licenses(self):
        result = self.run_installer()
        self.assertEqual(result.returncode, 0, result.stderr)
        binary = self.home / ".local/bin/typeit"
        self.assertEqual(subprocess.check_output([str(binary)], text=True).strip(), "typeit fixture")
        self.assertTrue((self.home / ".local/share/typeit/THIRD_PARTY_NOTICES.md").is_file())
        self.assertIn("PATH", result.stdout)

    def test_mac_arm64_custom_path_and_pinned_version(self):
        self.env.update(FAKE_OS="Darwin", FAKE_ARCH="arm64", TYPEIT_INSTALL_DIR=str(self.root / "path with spaces"))
        self.prepare("darwin", "arm64")
        result = self.run_installer("--version", "v1.2.3")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue((self.root / "path with spaces/typeit").is_file())
        self.assertIn("/download/v1.2.3/typeit_darwin_arm64.tar.gz", (self.release / "requests").read_text())

    def test_checksum_failure_preserves_existing_binary(self):
        binary = self.home / ".local/bin/typeit"
        binary.parent.mkdir(parents=True)
        binary.write_text("existing")
        (self.release / "SHA256SUMS").write_text("0" * 64 + "  typeit_linux_amd64.tar.gz\n")
        result = self.run_installer()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Checksum mismatch", result.stderr)
        self.assertEqual(binary.read_text(), "existing")

    def test_unsupported_architecture(self):
        self.env["FAKE_ARCH"] = "riscv64"
        self.assertNotEqual(self.run_installer().returncode, 0)
        self.assertFalse((self.release / "requests").exists())

    def test_missing_checksum(self):
        (self.release / "SHA256SUMS").write_text("")
        self.assertNotEqual(self.run_installer().returncode, 0)
        self.assertFalse((self.home / ".local/bin/typeit").exists())

    def test_download_failure(self):
        self.env["FAIL_DOWNLOAD"] = "1"
        result = self.run_installer()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Download failed", result.stderr)
        self.assertFalse((self.home / ".local/bin/typeit").exists())

    def test_invalid_version_and_help(self):
        self.assertNotEqual(self.run_installer("--version", "../main").returncode, 0)
        self.assertEqual(self.run_installer("--help").returncode, 0)
        self.assertFalse((self.release / "requests").exists())


if __name__ == "__main__":
    unittest.main()
