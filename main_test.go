package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

var pipFreezeOutput = ""
var pipFreezeExitCode = 0
var testExecutable string = os.Args[0] // we capture the test exec name so we can modify os.Args in our E2E tests

type File struct {
	name        string
	packageName string
	version     string
	mode        os.FileMode
}

type PipReq struct {
	name    string
	version string
}

func TestEndToEnd(t *testing.T) {
	oldStdout := os.Stdout
	os.Stdout = nil

	cases := []struct {
		testName string
		cliArgs  []string
		packages []PipReq
		files    []File
		exitCode int
	}{
		{
			"E2E Default Args, SUCCESS",
			[]string{"requirements_checker"},
			[]PipReq{{"foo", "1.2.3"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}},
			0,
		},
		{
			"Default E2E Args mismatch version, FAILURE",
			[]string{"requirements_checker"},
			[]PipReq{{"foo", "1.2.4"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}},
			1,
		},
		{
			"Default E2E Args missing env, FAILURE",
			[]string{"requirements_checker"},
			[]PipReq{},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}},
			1,
		},
		{
			"Default E2E Args missing file, FAILURE",
			[]string{"requirements_checker"},
			[]PipReq{{"foo", "1.2.3"}},
			[]File{{"requirements.txt", "bar", "3.4.5", 0644}},
			1,
		},
		{
			"E2E Single file, SUCCESS",
			[]string{"requirements_checker", "--files", "requirements.txt"},
			[]PipReq{{"foo", "1.2.3"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}},
			0,
		},
		{
			"E2E Single file mismatch version, FAILURE",
			[]string{"requirements_checker", "--files", "requirements.txt"},
			[]PipReq{{"foo", "1.2.4"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}},
			1,
		},
		{
			"E2E Single file missing env, FAILURE",
			[]string{"requirements_checker", "--files", "requirements.txt"},
			[]PipReq{},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}},
			1,
		},
		{
			"E2E Single file missing file, FAILURE",
			[]string{"requirements_checker", "--files", "requirements.txt"},
			[]PipReq{{"foo", "1.2.3"}},
			[]File{{"requirements.txt", "bar", "1.2.3", 0644}},
			1,
		},
		{
			"E2E Multiple files, SUCCESS",
			[]string{"requirements_checker", "--files", "requirements.txt,requirements-dev.txt"},
			[]PipReq{{"foo", "1.2.3"}, {"bar", "3.4.5"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}, {"requirements-dev.txt", "bar", "3.4.5", 0644}},
			0,
		},
		{
			"E2E Multiple files mismatch version, FAILURE",
			[]string{"requirements_checker", "--files", "requirements.txt,requirements-dev.txt"},
			[]PipReq{{"foo", "1.2.3"}, {"bar", "3.4.6"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}, {"requirements-dev.txt", "bar", "3.4.5", 0644}},
			1,
		},
		{
			"E2E Multiple files missing env, FAILURE",
			[]string{"requirements_checker", "--files", "requirements.txt,requirements-dev.txt"},
			[]PipReq{{"foo", "1.2.3"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}, {"requirements-dev.txt", "bar", "3.4.5", 0644}},
			1,
		},
		{
			"E2E Multiple files mismatch file, FAILURE",
			[]string{"requirements_checker", "--files", "requirements.txt,requirements-dev.txt"},
			[]PipReq{{"foo", "1.2.3"}, {"bar", "3.4.6"}},
			[]File{{"requirements.txt", "foo", "1.2.3", 0644}, {"requirements-dev.txt", "baz", "6.7.8", 0644}},
			1,
		},
	}

	for _, tc := range cases {
		AppFs = afero.NewMemMapFs()
		AppFs.MkdirAll("", 0755)
		_ = writeFiles(AppFs, tc.files, t)

		pipFreezeOutput = generatePipFreezeOutput(tc.packages)
		pipFreezeExitCode = 0
		execCommand = fakeExecCommand
		defer func() { execCommand = exec.Command }()
		os.Args = []string(tc.cliArgs)
		rc := mainWrapper()
		if rc != tc.exitCode {
			t.Errorf("\"%s\" - Invalid exit code returned, expected %d, found %d", tc.testName, tc.exitCode, rc)
		}

	}
	os.Stdout = oldStdout

}

func TestGetEnvironmentNonZeroExitCode(t *testing.T) {
	pipFreezeOutput = ""
	pipFreezeExitCode = 1
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	reqs := make(map[string]Requirement)
	out, err := getEnvironment(reqs)

	if err == nil {
		t.Errorf("Did not raise error! %s", out)
	}
}

func TestGetEnvironment(t *testing.T) {
	cases := []struct {
		testName string
		packages []PipReq
	}{
		{"Single req", []PipReq{{"foo", "1.2.3"}}},
		{"Multiple reqs", []PipReq{{"foo", "1.2.3"}, {"bar", "3.2.1"}, {"baz", "4.5.6-b"}}},
	}

	for _, tc := range cases {
		pipFreezeOutput = generatePipFreezeOutput(tc.packages)
		pipFreezeExitCode = 0

		execCommand = fakeExecCommand
		defer func() { execCommand = exec.Command }()

		reqs := make(map[string]Requirement)
		reqs, err := getEnvironment(reqs)
		if err != nil {
			t.Errorf("Error raised")
		}

		for _, p := range tc.packages {
			v, ok := reqs[p.name]
			if !ok {
				t.Errorf("\"%s\" - Key '%s' not found", tc.testName, p.name)
			}
			if v.Environment != p.version {
				t.Errorf("Version not correctly parsed for package '%s'. Expected '%s', found '%s'", p.name, p.version, v.Environment)
			}
			if len(reqs) != len(tc.packages) {
				t.Errorf("Incorrect number of keys found. Expected '%d', found '%d'", len(tc.packages), len(reqs))
			}
		}
	}
}

func TestParseFiles(t *testing.T) {

	cases := []struct {
		name  string
		files []File
	}{
		{
			"Test Single Files and Req",
			[]File{
				{"test/requirements.txt", "foo", "1.2.3", 0644},
			},
		},
		{
			"Test Single File and Multiple Reqs",
			[]File{
				{"test/requirements.txt", "foo", "1.2.3", 0644},
				{"test/requirements.txt", "bar", "3.4.5", 0644},
			},
		},
		{
			"Test Multiple Files and Single Reqs",
			[]File{
				{"test/requirements.txt", "foo", "1.2.3", 0644},
				{"test/requirements-dev.txt", "bar", "3.2.1", 0644},
			},
		},
		{
			"Test Multiple Files and Reqs",
			[]File{
				{"test/requirements.txt", "foo", "1.2.3", 0644},
				{"test/requirements.txt", "quux", "9.8.7", 0644},
				{"test/requirements-dev.txt", "bar", "3.2.1", 0644},
				{"test/requirements-dev.txt", "baz", "4.5.6-b", 0644},
			},
		},
	}

	for _, tc := range cases {
		AppFs = afero.NewMemMapFs()
		AppFs.MkdirAll("test", 0755)

		var filenames []string
		for _, s := range tc.files {
			filenames = append(filenames, s.name)
		}
		filenameString := strings.Join(filenames[:], ",")

		writeCount := writeFiles(AppFs, tc.files, t)

		reqs := make(map[string]Requirement)
		reqs = parseFiles(filenameString, reqs)
		if len(reqs) != writeCount {
			t.Errorf("Incorrect Number of requirements processed! Expected %d, found %d", writeCount, len(reqs))
		}

		for _, f := range tc.files {
			v, ok := reqs[f.packageName]
			if !ok {
				t.Errorf("Key '%s' not found", f.packageName)
			}
			if v.Defined != f.version {
				t.Errorf("Version not correctly parsed for package '%s'. Expected '%s', found '%s'", f.packageName, f.version, v.Defined)
			}
			if v.Found != f.name {
				t.Errorf("Filename not correctly parsed for package '%s', Expected '%s', found '%s'", f.packageName, f.name, v.Found)
			}
		}
	}
}

func writeFiles(fs afero.Fs, files []File, t *testing.T) int {
	writeCount := 0
	for _, f := range files {
		data := []byte(fmt.Sprintf("%s==%s", f.packageName, f.version))
		writeToFile(AppFs, f.name, data, f.mode, t)
		writeCount++
	}
	return writeCount

}
func writeToFile(fs afero.Fs, filename string, data []byte, mode os.FileMode, t *testing.T) {
	ok, err := afero.Exists(fs, filename)
	if err != nil {
		t.Errorf("Error appending to file %s", filename)
	}
	if ok {
		fileData, err := afero.ReadFile(fs, filename)
		if err != nil {
			t.Errorf("Error appending to file %s", filename)
		}
		fileData = append(fileData, byte('\n'))
		data = append(fileData, data...)

	}

	afero.WriteFile(AppFs, filename, data, mode)

}
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	rc, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(rc)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(testExecutable, cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "STDOUT=" + pipFreezeOutput, "EXIT_STATUS=" + strconv.Itoa(pipFreezeExitCode)}
	return cmd
}

func generatePipFreezeOutput(packages []PipReq) string {
	var tempPackages []string
	for _, p := range packages {
		tempPackages = append(tempPackages, fmt.Sprintf("%s==%s", p.name, p.version))
	}
	return strings.Join(tempPackages, "\n")
}
