package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/mike-winberry/lulalib/src/pkg/common/oscal"
)

func OscalFromPath(t *testing.T, path string) *oscalTypes.OscalCompleteSchema {
	t.Helper()
	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}
	oscalModel, err := oscal.NewOscalModel(data)
	if err != nil {
		t.Fatalf("error creating oscal model from file: %v", err)
	}

	return oscalModel
}

func CreateTempFile(t *testing.T, ext string) *os.File {
	t.Helper()
	tempFile, err := os.CreateTemp("", fmt.Sprintf("tmp-*.%s", ext))
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return tempFile
}
