package custerr

import (
	"os"
)

func Fatal(message string) {
	os.Stderr.WriteString(message)
	os.Stdout.WriteString("null")
	os.Exit(1)
}
