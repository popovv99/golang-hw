package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCmd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "executor_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	progPath := filepath.Join(tmpDir, "testprog.go")
	prog := `package main
import "os"
func main() {
	if os.Getenv("FOO") == "bar" {
		os.Exit(0)
	}
	os.Exit(1)
}
`
	if err := os.WriteFile(progPath, []byte(prog), 0o644); err != nil {
		t.Fatal(err)
	}

	env := Environment{
		"FOO": EnvValue{Value: "bar", NeedRemove: false},
	}

	returnCode := RunCmd([]string{"go", "run", progPath}, env)
	if returnCode != 0 {
		t.Errorf("expected return code 0, got %d", returnCode)
	}
}

func TestRunCmdRemoveEnv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "executor_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Тест 1: проверяем что переменная есть
	progPath1 := filepath.Join(tmpDir, "testprog1.go")
	prog1 := `package main
import "os"
func main() {
	if os.Getenv("REMOVE_ME") == "value" {
		os.Exit(0)
	}
	os.Exit(1)
}
`
	if err := os.WriteFile(progPath1, []byte(prog1), 0o644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("REMOVE_ME", "value")
	env1 := Environment{}

	returnCode1 := RunCmd([]string{"go", "run", progPath1}, env1)
	if returnCode1 != 0 {
		t.Errorf("test 1: expected return code 0 (variable exists), got %d", returnCode1)
	}

	// Тест 2: проверяем что переменная удалена
	progPath2 := filepath.Join(tmpDir, "testprog2.go")
	prog2 := `package main
import "os"
func main() {
	_, exists := os.LookupEnv("REMOVE_ME")
	if !exists {
		os.Exit(0) // переменная действительно удалена
	}
	os.Exit(1) // переменная все еще существует
}
`
	if err := os.WriteFile(progPath2, []byte(prog2), 0o644); err != nil {
		t.Fatal(err)
	}

	env2 := Environment{
		"REMOVE_ME": EnvValue{Value: "", NeedRemove: true},
	}

	returnCode2 := RunCmd([]string{"go", "run", progPath2}, env2)
	if returnCode2 != 0 {
		t.Errorf("test 2: expected return code 0 (variable removed), got %d", returnCode2)
	}

	// Тест 3: проверяем что переменная с пустым значением установлена
	progPath3 := filepath.Join(tmpDir, "testprog3.go")
	prog3 := `package main
import "os"
func main() {
	_, exists := os.LookupEnv("EMPTY_VAR")
	if exists {
		os.Exit(0) // переменная установлена (даже с пустым значением)
	}
	os.Exit(1) // переменная не установлена
}
`
	if err := os.WriteFile(progPath3, []byte(prog3), 0o644); err != nil {
		t.Fatal(err)
	}

	env3 := Environment{
		"EMPTY_VAR": EnvValue{Value: "", NeedRemove: false},
	}

	returnCode3 := RunCmd([]string{"go", "run", progPath3}, env3)
	if returnCode3 != 0 {
		t.Errorf("test 3: expected return code 0 (variable set with empty value), got %d", returnCode3)
	}
}
