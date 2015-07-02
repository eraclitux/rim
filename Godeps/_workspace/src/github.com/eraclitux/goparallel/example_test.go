// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package goparallel_test

import (
	"fmt"
	"math"
	"runtime"

	"github.com/eraclitux/goparallel"
)

type job struct {
	start   int
	stop    int
	results map[int]bool
}

func (j *job) Execute() {
	j.results = make(map[int]bool)
	for i := j.start; i <= j.stop; i++ {
		j.results[i] = isPrime(uint64(i))
	}
}

func isPrime(n uint64) bool {
	if n <= 2 {
		return true
	}
	var i uint64
	i = 2
	num := uint64(math.Sqrt(float64(n)))
	for i <= num {
		if n%i == 0 {
			return false
		}
		i++
	}
	return true
}

// Example shows example usage of the package.
func Example() {
	// Creates the slice of tasks that we want to execute in parallel.
	tasks := make([]goparallel.Tasker, 0, 1e3)
	prev := 1
	// Limit is the bigger number to check.
	var limit int = 1e5
	// Create as much tasks as number of cores.
	d := limit / runtime.NumCPU()
	for i := 1; i < limit; i++ {
		// This is not the best way to disbrubute load
		// as complexity is not the same in different
		// intervals (bigger numbers are more difficult to verify),
		// so some cores remains idle sooner.
		// We could increase efficency making different interval lenghts.
		if (i % d) == 0 {
			j := &job{start: prev, stop: i}
			prev = i + 1
			tasks = append(tasks, goparallel.Tasker(j))
		}
	}
	// Do not forget last interval.
	j := &job{start: prev, stop: limit}
	tasks = append(tasks, goparallel.Tasker(j))

	// Run tasks in parallel using all cores.
	err := goparallel.RunBlocking(tasks)
	if err == nil {
		fmt.Println("Example ok")
	}

	// Output:
	// Example ok
}
