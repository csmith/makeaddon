package makeaddon

import (
	"fmt"
	"github.com/sebdah/goldie/v2"
	"os"
	"path"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []string{"example1", "example2", "example3", "example4"}
	gold := goldie.New(t)

	for i := range tests {
		t.Run(tests[i], func(t *testing.T) {
			f, _ := os.Open(path.Join("testdata", fmt.Sprintf("%s.yml", tests[i])))
			defer f.Close()

			actual, err := ReadMetaData(f, "unused")
			if err != nil {
				t.Fatalf("unable to read pkgmeta: %v", err)
			}

			gold.AssertJson(t, tests[i], actual)
		})
	}
}
