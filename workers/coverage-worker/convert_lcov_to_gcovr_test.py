import importlib.util
import tempfile
import textwrap
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("convert_lcov_to_gcovr.py")
SPEC = importlib.util.spec_from_file_location("convert_lcov_to_gcovr", MODULE_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(MODULE)


class ConvertLcovToGcovrTest(unittest.TestCase):
    def test_normalize_filename(self):
        self.assertEqual(
            MODULE.normalize_filename("/tmp/bitcoin/src/node/foo.cpp"),
            "src/node/foo.cpp",
        )
        self.assertEqual(
            MODULE.normalize_filename("/tmp/bitcoin/build/src/node/foo.cpp"),
            "src/node/foo.cpp",
        )
        self.assertEqual(
            MODULE.normalize_filename("build/src/node/foo.cpp"),
            "src/node/foo.cpp",
        )

    def test_should_include_filename(self):
        self.assertTrue(MODULE.should_include_filename("src/node/foo.cpp"))
        self.assertTrue(MODULE.should_include_filename("src/node/foo.h"))
        self.assertFalse(MODULE.should_include_filename("build/src/generated.h"))
        self.assertFalse(MODULE.should_include_filename("src/test/foo.cpp"))
        self.assertFalse(MODULE.should_include_filename("src/node/foo.txt"))

    def test_lcov_to_gcovr_json_rewrites_build_paths_to_src(self):
        lcov = textwrap.dedent(
            """\
            SF:/tmp/bitcoin/src/node/foo.cpp
            DA:10,2
            end_of_record
            SF:/tmp/bitcoin/build/src/node/foo.cpp
            DA:11,3
            end_of_record
            """
        )

        with tempfile.NamedTemporaryFile("w+", encoding="utf-8") as handle:
            handle.write(lcov)
            handle.flush()
            data = MODULE.lcov_to_gcovr_json(handle.name)

        self.assertEqual(len(data["files"]), 1)
        self.assertEqual(data["files"][0]["file"], "src/node/foo.cpp")
        self.assertEqual(
            [line["line_number"] for line in data["files"][0]["lines"]],
            [10, 11],
        )


if __name__ == "__main__":
    unittest.main()
