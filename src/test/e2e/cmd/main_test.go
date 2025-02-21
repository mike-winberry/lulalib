package cmd_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mike-winberry/lulalib/src/test/util"
	"github.com/spf13/cobra"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	m.Run()
}

func runCmdTest(t *testing.T, rootCmd *cobra.Command, cmdArgs ...string) error {
	_, _, err := util.ExecuteCommand(rootCmd, cmdArgs...)
	if err != nil {
		return err
	}

	return nil
}

func runCmdTestWithGolden(t *testing.T, goldenFilePath, goldenFileName string, rootCmd *cobra.Command, cmdArgs ...string) error {
	_, output, err := util.ExecuteCommand(rootCmd, cmdArgs...)
	if err != nil {
		return err
	}

	testGolden(t, goldenFilePath, goldenFileName, output)

	return nil
}

func runCmdTestWithOutputFile(t *testing.T, goldenFilePath, goldenFileName, outExt string, rootCmd *cobra.Command, cmdArgs ...string) error {
	tempFileName := fmt.Sprintf("output-%s.%s", goldenFileName, outExt)
	defer os.Remove(tempFileName)

	cmdArgs = append(cmdArgs, "-o", tempFileName)
	_, _, err := util.ExecuteCommand(rootCmd, cmdArgs...)
	if err != nil {
		return err
	}

	// Read the output file
	data, err := os.ReadFile(tempFileName)
	if err != nil {
		return err
	}

	// Scrub uniquely generated data
	data = scrubData(data)

	testGolden(t, goldenFilePath, goldenFileName, string(data))

	return nil
}

func testGolden(t *testing.T, filePath, filename, got string) {
	t.Helper()

	got = strings.ReplaceAll(got, "\r\n", "\n")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	goldenPath := filepath.Join(wd, "testdata", filePath, filename+".golden")

	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	wantBytes, _ := os.ReadFile(goldenPath)
	want := string(wantBytes)
	diff := cmp.Diff(want, got)

	if diff != "" {
		t.Fatalf("`%s` does not match.\n\nDiff:\n%s", goldenPath, diff)
	}
}

func scrubData(data []byte) []byte {
	timestamps := regexp.MustCompile(`(?i)(last-modified|published:\s)*(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:[-+]\d{2}:\d{2}|Z)?)`)
	uuids := regexp.MustCompile(`(?i)(uuid:\s*)([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})`)

	output := timestamps.ReplaceAllString(string(data), "${1}XXX")
	output = uuids.ReplaceAllString(string(output), "${1}XXX")
	return []byte(output)
}
