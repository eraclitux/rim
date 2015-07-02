// Copyright (c) 2015 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

// +build !debug

// Package stracer is possibly the simplest tracing package.
// The idea is that during early stages of code development one wants
// an easy way to inspect values of vars that can be quickly disabled.
//
// To enable stracer functions to print just build, run or test with ``debug`` tag otherwise they will do noop:
// 	go build -tags debug
// 	go test -tags debug
// 	go run -tags debug main.go
// stderr is used to not perturb Examples() functions.
//
// Credits
//
// Original idea is by Dave Cheney http://dave.cheney.net.
package stracer

// Traceln prints to stderr if '-tags debug'
// is used when building/running, noop otherwise.
// stderr is used to not perturb example tests.
func Traceln(args ...interface{}) {}

// Tracef prints to stderr using supplied format
// if '-tags debug' is used when building/running,
// noop otherwise.
// stderr is used to not perturb example tests.
func Tracef(format string, a ...interface{}) {}
