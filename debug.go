// +build debug

package main

import "fmt"

func debugPrintln(args ...interface{}) {
	fmt.Println(args...)
}
