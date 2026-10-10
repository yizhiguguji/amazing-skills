package main

import (
	amazing "amazing-skills"
	"os"
)

func main() { os.Exit(amazing.CLI(os.Args[1:], os.Stdout)) }
