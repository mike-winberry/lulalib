package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mike-winberry/lulalib/src/cmd/dev"
)

func TestDevValidateCommand(t *testing.T) {

	test := func(t *testing.T, args ...string) error {
		t.Helper()
		rootCmd := dev.DevValidateCommand()

		return runCmdTest(t, rootCmd, args...)
	}

	testAgainstGolden := func(t *testing.T, goldenFileName string, args ...string) error {
		rootCmd := dev.DevValidateCommand()

		return runCmdTestWithGolden(t, "dev/validate/", goldenFileName, rootCmd, args...)
	}

	t.Run("Valid validation file", func(t *testing.T) {

		args := []string{
			"--input-file", "./testdata/dev/get-resources/opa.validation.yaml",
		}

		err := test(t, args...)
		require.NoError(t, err)
	})

	t.Run("Valid validation file - template", func(t *testing.T) {
		args := []string{
			"--input-file", "./testdata/dev/get-resources/opa.validation.tpl.yaml",
		}

		err := test(t, args...)
		require.NoError(t, err)
	})

	t.Run("Valid validation file with passing tests", func(t *testing.T) {

		args := []string{
			"--input-file", "./testdata/dev/validate/opa.validation-passing-test.yaml",
			"--run-tests",
		}

		err := test(t, args...)
		require.NoError(t, err)
	})

	t.Run("Valid validation file with failing tests", func(t *testing.T) {

		args := []string{
			"--input-file", "./testdata/dev/validate/opa.validation-failing-test.yaml",
			"--run-tests",
		}

		err := test(t, args...)
		require.Error(t, err)
	})

	t.Run("Test help", func(t *testing.T) {
		err := testAgainstGolden(t, "help", "--help")
		require.NoError(t, err)
	})

}
