package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	// Place your code here

	dir, err := os.MkdirTemp("", "envdir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFiles := map[string]string{
		"FOO":   "bar",
		"HELLO": `"hello"`,
		"UNSET": " ",
		"EMPTY": "",
	}

	for name, value := range testFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	env, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if env["FOO"].Value != "bar" {
		t.Errorf("expected FOO=bar, got %s", env["FOO"].Value)
	}
	if env["HELLO"].Value != `"hello"` {
		t.Errorf("expected HELLO=\"hello\", got %s", env["HELLO"].Value)
	}
	if env["UNSET"].Value != "" {
		t.Errorf("expected UNSET=, got %s", env["UNSET"].Value)
	}
	if env["UNSET"].NeedRemove {
		t.Errorf("expected UNSET to not be removed")
	}
	if !env["EMPTY"].NeedRemove {
		t.Error("expected EMPTY to have NeedRemove=true")
	}
	if env["EMPTY"].Value != "" {
		t.Errorf("expected EMPTY value to be empty, got %s", env["EMPTY"].Value)
	}
}
