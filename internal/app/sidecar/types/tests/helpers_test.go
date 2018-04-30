package types_tests // nolint: golint

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/lawrencegripper/ion/sidecar/types"
)

func TestIsValidEvent(t *testing.T) {
	testCases := []struct {
		slice    []string
		target   string
		contains bool
	}{
		{
			slice: []string{
				"alice",
				"frank",
				"123",
			},
			target:   "frank",
			contains: true,
		},
		{
			slice: []string{
				"",
				"TEST",
				"T35t",
				"Test",
				"tes",
				"tester",
				"t est",
				"test ",
				" test",
			},
			target:   "test",
			contains: false,
		},
		{
			slice: []string{
				"a",
				"aa",
				"aaa",
				"aaa a",
				"aaaaa",
			},
			target:   "aaaa",
			contains: false,
		},
		{
			slice: []string{
				"123",
				"true",
				"false",
				"####",
				"",
				"\"",
			},
			target:   "\"",
			contains: true,
		},
	}
	for _, test := range testCases {
		contains := ContainsString(test.slice, test.target)
		if contains != test.contains {
			t.Errorf("expecting '%t' but got '%t' for event '%s'", test.contains, contains, test.target)
		}
	}
}

func TestJoinBlobPath(t *testing.T) {
	testCases := []struct {
		strs     []string
		expected string
	}{
		{
			strs: []string{
				"alice",
				"frank",
				"123",
			},
			expected: "alice-frank-123",
		},
		{
			strs: []string{
				"-",
				"-",
				"-",
			},
			expected: "-----",
		},
		{
			strs: []string{
				"",
				"testtest",
				"",
			},
			expected: "-testtest-",
		},
	}
	for _, test := range testCases {
		actual := JoinBlobPath(test.strs[0], test.strs[1], test.strs[2])
		if actual != test.expected {
			t.Errorf("expecting '%s' but got '%s'", test.expected, actual)
		}
	}
}

func TestRemoveFile(t *testing.T) {
	testCases := []struct {
		filename string
		removed  bool
	}{
		{
			filename: "testfile",
			removed:  true,
		},
		{
			filename: "",
			removed:  true,
		},
	}
	for _, test := range testCases {
		if test.filename != "" {
			f, err := os.Create(test.filename)
			f.Close()
			if err != nil {
				t.Errorf("error creating test file '%s'", test.filename)
				continue
			}
		}
		err := RemoveFile(test.filename)
		if test.removed == (err != nil) {
			t.Errorf("expecting file '%s' to be removed, but go error '%+v'", test.filename, err)
		}
	}
}

func TestClearDir(t *testing.T) {
	testCases := []struct {
		dirname string
		files   []string
	}{
		{
			dirname: "testdir",
			files: []string{
				"file1.txt",
				"file2.txt",
				"file3.txt",
			},
		},
	}
	for _, test := range testCases {
		_ = os.MkdirAll(test.dirname, 0777)
		for _, file := range test.files {
			path := path.Join(test.dirname, file)
			f, err := os.Create(path)
			f.Close()
			if err != nil {
				t.Errorf("error creating test file '%s'", file)
				continue
			}
		}
		err := ClearDir(test.dirname)
		if err != nil {
			t.Errorf("test failed with error '%+v'", err)
			continue
		}
		leftOverFiles, err := ioutil.ReadDir(test.dirname)
		if err != nil {
			t.Errorf("test failed with error '%+v'", err)
		}
		if len(leftOverFiles) > 0 {
			t.Errorf("expected all files from directory '%s' to be removed but were not", test.dirname)
		}
		_ = os.RemoveAll(test.dirname)
	}
}

func TestCompareHash(t *testing.T) {
	testCases := []struct {
		raw string
	}{
		{
			raw: "abcabc",
		},
	}
	for _, test := range testCases {
		hash := Hash(test.raw)
		err := CompareHash(test.raw, hash)
		if err != nil {
			t.Errorf("expected no errors to be returned but got '%+v'", err)
		}
	}
}
