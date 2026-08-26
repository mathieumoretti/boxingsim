package config_test

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLFilesAreValid(t *testing.T) {
	files := []string{
		"development.yaml",
		"test.yaml",
		"production.yaml",
	}

	fs := os.DirFS(".")
	for _, filename := range files {
		t.Run(filename, func(t *testing.T) {
			file, err := fs.Open(filename)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", filename, err)
			}
			defer file.Close()

			var result any
			if err := yaml.NewDecoder(file).Decode(&result); err != nil {
				t.Errorf("%s is not valid YAML: %v", filename, err)
			}
		})
	}
}
