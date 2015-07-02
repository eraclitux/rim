// Copyright (c) 2015 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package stracer_test

import (
	"fmt"

	"github.com/eraclitux/stracer"
)

func Example() {
	s := "my-value"
	fmt.Println("This is printed")
	stracer.Traceln("This string will be printed to stderr only if '-tags debug' is used when building/running. Value of s:", s)

	// Output:
	//This is printed

}
