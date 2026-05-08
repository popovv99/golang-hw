package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	// #nosec G204
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	currentEnv := os.Environ()
	finalEnv := make([]string, 0, len(currentEnv))

	for _, e := range currentEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			if envVal, exists := env[parts[0]]; exists && envVal.NeedRemove {
				continue
			}
		}
		finalEnv = append(finalEnv, e)
	}
	for k, v := range env {
		if !v.NeedRemove {
			finalEnv = append(finalEnv, k+"="+v.Value)
		}
	}
	c.Env = finalEnv

	err := c.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}
