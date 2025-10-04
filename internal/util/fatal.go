package util

import "os"

func ShowFatalError(title, message string) {
	showErrorDialog(title, message)
	os.Exit(1)
}
