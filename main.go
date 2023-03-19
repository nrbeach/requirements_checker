package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/afero"
)

const VERSION string = "0.1.1"

var AppFs = afero.NewOsFs()

type Requirement struct {
	Environment string
	Defined     string
	Found       string
}

func (r Requirement) Cmp() bool {
	return r.Environment == r.Defined
}

var execCommand = exec.Command

const (
	SUCCESS  int = 0
	MISMATCH     = 1
)

var files string
var quiet bool
var version bool

func init() {
	flag.StringVar(&files, "files", "requirements.txt", "A comma separated list of files")
	flag.BoolVar(&quiet, "quiet", false, "Suppress output")
	flag.BoolVar(&version, "version", false, "Displays the version info and exits")
}

func mainWrapper() int {
	// Testing in Go sucks
	rc := SUCCESS
	flag.Parse()
	if version == true {
		fmt.Printf("requirements_checker: Version %s\n", VERSION)
		return rc
	}

	reqs := make(map[string]Requirement)
	reqs = parseFiles(files, reqs)
	reqs, err := getEnvironment(reqs)
	if err != nil {
		log.Fatal(err)
	}
	t, rc := generateTable(reqs)

	if rc != 0 && !quiet {
		t.Render()
	}

	return rc
}

func main() {
	os.Exit(mainWrapper())
}

func generateTable(m map[string]Requirement) (table.Writer, int) {
	rc := SUCCESS
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Module", "Environment", "Defined", "Found"})
	for k, v := range m {
		if !v.Cmp() {
			if len(v.Environment) == 0 {
				v.Environment = "Missing"
			}
			if len(v.Defined) == 0 {
				v.Defined = "Missing"
			}
			if len(v.Found) == 0 {
				v.Found = "Environment"
			}
			t.AppendRow([]interface{}{k, v.Environment, v.Defined, v.Found})
			rc = MISMATCH
		}
	}
	return t, rc
}
func getEnvironment(m map[string]Requirement) (map[string]Requirement, error) {
	out, err := execCommand("pip", "freeze").CombinedOutput()
	if err != nil {
		return m, err
	}

	reader := bytes.NewReader(out)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		module_ver := strings.Split(scanner.Text(), "==")
		module := module_ver[0]
		version := module_ver[1]
		_, ok := m[module]
		if !ok {
			m[module] = Requirement{}
		}
		val := m[module]
		val.Environment = version
		m[module] = val
	}
	return m, err
}

func readAndParseFile(f string, m map[string]Requirement) {
	file, err := AppFs.Open(f)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}
		module_ver := strings.Split(scanner.Text(), "==")
		module := module_ver[0]
		version := module_ver[1]
		_, ok := m[module]
		if !ok {
			m[module] = Requirement{}
		}
		val := m[module]
		val.Defined = version
		val.Found = f
		m[module] = val
	}
}

func parseFiles(files string, reqs map[string]Requirement) map[string]Requirement {

	for _, f := range strings.Split(files, ",") {
		readAndParseFile(f, reqs)
	}

	return reqs
}
