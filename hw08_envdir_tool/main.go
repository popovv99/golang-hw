package main

import (
	"fmt"
	"os"
)

func main() {
	// Place your code here.
	if len(os.Args) < 3 {
		fmt.Println("usage: go-envdir <env-dir> <command> [args...]")
		return
	}
	env, err := ReadDir(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	returnCode := RunCmd(os.Args[2:], env)
	os.Exit(returnCode)
}
