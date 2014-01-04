package custerr

import (
	"os"
)

func Fatal(message string) {
	os.Stderr.WriteString(message + "\n")
	os.Stdout.WriteString("null")
	os.Exit(1)
}
