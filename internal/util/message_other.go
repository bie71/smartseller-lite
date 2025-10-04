//go:build !windows

package util

import (
	"bufio"
	"fmt"
	"os"
)

func showErrorDialog(title, message string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", title, message)
	fmt.Fprintln(os.Stderr, "Press Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
