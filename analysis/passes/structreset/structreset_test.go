package structreset

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestStructReset(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer)
}