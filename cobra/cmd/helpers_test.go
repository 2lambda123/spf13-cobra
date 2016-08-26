package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var _ = fmt.Println
var _ = os.Stderr

func checkGuess(t *testing.T, wd, input, expected string) {
	testWd = wd
	inputPath = input
	guessProjectPath()

	if projectPath != expected {
		t.Errorf("Unexpected Project Path. \n Got: %q\nExpected: %q\n", projectPath, expected)
	}

	reset()
}

func reset() {
	testWd = ""
	inputPath = ""
	projectPath = ""
}

func TestProjectPath(t *testing.T) {
	checkGuess(t, "", filepath.Join("github.com", "spf13", "hugo"), filepath.Join(getSrcPath(), "github.com", "spf13", "hugo"))
	checkGuess(t, "", filepath.Join("spf13", "hugo"), filepath.Join(getSrcPath(), "github.com", "spf13", "hugo"))
	checkGuess(t, "", filepath.Join("/", "bar", "foo"), filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo", "baz", filepath.Join("/", "bar", "foo", "baz"))
	checkGuess(t, "/bar/foo/cmd", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo/command", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "/bar/foo/commands", "", filepath.Join("/", "bar", "foo"))
	checkGuess(t, "github.com/spf13/hugo/../hugo", "", filepath.Join("github.com", "spf13", "hugo"))
}

type inPathTestCase struct {
	Src    string
	Prj    string
	InPath bool
}

func TestInPath(t *testing.T) {
	cases := []inPathTestCase{
		{"/bar/foo", "/bar/foo", false},
		{"/bar/foo", "/bar/foo/baz", true},
		{"/bar/foo/baz", "/bar/foo", false},
		{"/bar/foo", "/bar/foo/.wierd..dirname/", true},
		{"c:\\bar\\foo", "C:\\bar\\foo", false},
		{"c:\\bar\\..\\bar\\foo", "C:\\bar\\foo\\baz", true},
	}
	if runtime.GOOS == "windows" {
		cases = append(
			cases,
			inPathTestCase{"C:/Bar/foo", "c:/bar/foo/baz", true},
		)
	} else {
		cases = append(
			cases,
			inPathTestCase{"C:/Bar/foo", "c:/bar/foo/baz", false},
		)
	}

	for _, tc := range cases {
		ip := inPath(tc.Src, tc.Prj)
		if tc.InPath != ip {
			if tc.InPath {
				t.Errorf("Unexpected %s determined as inside %s", tc.Prj, tc.Src)
			} else {
				t.Errorf("Unexpected %s determined as not inside %s", tc.Prj, tc.Src)
			}
		}
	}
}
