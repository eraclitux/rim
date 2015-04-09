// +build debug

package main

import (
	"fmt"
	"os"
)

func debugPrintln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}
