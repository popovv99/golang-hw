package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	env := make(Environment)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && !strings.Contains(entry.Name(), "=") {
			f, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			if len(f) == 0 {
				env[entry.Name()] = EnvValue{
					Value:      "",
					NeedRemove: true,
				}
				continue
			}
			lines := bytes.SplitN(f, []byte{'\n'}, 2)
			line := bytes.ReplaceAll(lines[0], []byte{0x00}, []byte{'\n'})
			s := strings.TrimRight(string(line), " \t")
			env[entry.Name()] = EnvValue{
				Value:      s,
				NeedRemove: false,
			}
		}
	}
	return env, nil
}
