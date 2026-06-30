// code-from-spec: SPEC/golang/test/utils/chdir@cKY1laDjF37NxxYeajTgtPDgdI0
package testutils

import (
	"os"
	"testing"
)

func Chdir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("Chdir cleanup: %v", err)
		}
	})
	return dir
}
