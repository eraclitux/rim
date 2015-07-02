// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package goparallel

import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"testing"
	"time"
)

type dummy struct {
	done bool
}

type dummyNop struct {
	done bool
}

// isPrime returns true if a given int is prime.
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

func (d *dummy) Execute() {
	for i := 0; i < 1e4; i++ {
		isPrime(uint64(i))
	}
	d.done = true
}

func (d dummyNop) Execute() {
	for i := 0; i < 1e4; i++ {
		isPrime(uint64(i))
	}
	d.done = true
}

func TestRunBlocking(t *testing.T) {
	tasks := make([]Tasker, 1e2)
	// []*dummy does not convert []Tasker.
	// We need to iterate on []Tasker making an explicit cast.
	// http://golang.org/doc/faq#convert_slice_of_interface
	for i := range tasks {
		tasks[i] = Tasker(&dummy{false})
	}

	err := RunBlocking(tasks)
	if err != nil {
		t.Fatal("Test has failed", err)
	}
	for _, e := range tasks {
		if !e.(*dummy).done {
			t.Fatal("Error executig task")
		}
	}
}

// TestRunBlocking_nopointer shows that Execute() method
// must be implemented on a pointer receiver or computed values
// will be lost.
func TestRunBlocking_nopointer(t *testing.T) {
	tasks := make([]Tasker, 1e1)
	for i := range tasks {
		tasks[i] = Tasker(&dummyNop{false})
	}

	err := RunBlocking(tasks)
	if err != nil {
		t.Fatal("Test has failed", err)
	}
	for _, e := range tasks {
		if e.(*dummyNop).done {
			t.Fatal("Error, receiver modified!")
		}
	}
}

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

// Example_performance verify that using goparallel is faster than a serial execution.
func Example_performance() {
	// Creates the slice of tasks that we want to execute in parallel.
	tasks := make([]Tasker, 0, 1e3)
	prev := 1
	// Limit is the bigger number to check.
	var limit int = 1e6
	pre := time.Now()
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
			tasks = append(tasks, Tasker(j))
		}
	}
	// Do not forget last interval.
	j := &job{start: prev, stop: limit}
	tasks = append(tasks, Tasker(j))

	// Run tasks in parallel using all cores.
	err := RunBlocking(tasks)
	if err != nil {
		log.Fatal(err)
	}

	after := time.Now()
	Δt1 := after.Sub(pre)

	// Lets compare execution time using single core.
	pre = time.Now()
	results := make(map[int]bool)
	for i := 1; i <= limit; i++ {
		results[i] = isPrime(uint64(i))
	}
	after = time.Now()
	Δt2 := after.Sub(pre)
	if Δt2 > Δt1 {
		fmt.Println("Using goparallel takes less time.")
	} else {
		fmt.Println("Using goparallel takes more time.")
	}
	// We use stderr as stdout is checked to pass test.
	fmt.Fprintf(os.Stderr, "%30s %9dns\n", "Time with goworker:", Δt1)
	fmt.Fprintf(os.Stderr, "%30s %9dns\n", "Time without goworker:", Δt2)

	// Output:
	// Using goparallel takes less time.
}
