// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
)

const TEST_FAILED_ERROR = "Test failed due to unexpected errors being logged. If the error above is expected, please add it to the list of allowed errors by invoking TestOutput.AllowErrorsMatching"
const POST_TERMINATED_ERROR = "*** Logged error after TestOutput.TestTerminated:"
const TEST_RUNNER_PANIC_ERROR = "*** Test runner panic while trying to fail test (try using TestOutput.TestTerminated):"

type TLog interface {
	Fatal(args ...interface{})
	Log(args ...interface{})
	Error(args ...interface{})
	Name() string
	Fail()
}

func NewTestOutput(tb TLog, formatter LogFormatter) *TestOutput {
	return &TestOutput{tb: tb, formatter: formatter}
}

type TestOutput struct {
	sync.RWMutex
	formatter            LogFormatter
	tb                   TLog
	loggingDisabled      uint32 // so that we can atomic.Store and atomic.Load it
	allowedErrors        []string
	allowedErrorPatterns []*regexp.Regexp
	hasErrors            bool
	testTerminated       bool
}

func (o *TestOutput) SetFilters(_ ...Filter) {
}

// assumes read lock (o.RLock())
func (o *TestOutput) allowed(message string, fields []*Field) bool {
	for _, allowedPattern := range o.allowedErrorPatterns {
		if allowedPattern.MatchString(message) {
			return true
		}
		for _, f := range fields {
			if f.Key == "error" {
				if allowedPattern.MatchString(f.String()) {
					return true
				}
			}
		}
	}

	return false
}

func (o *TestOutput) AllowErrorsMatching(pattern string) {
	o.Lock()
	defer o.Unlock()
	compiledPattern, _ := regexp.Compile(pattern)
	o.allowedErrors = append(o.allowedErrors, pattern)
	o.allowedErrorPatterns = append(o.allowedErrorPatterns, compiledPattern)
}

func (o *TestOutput) HasErrors() bool {
	o.RLock()
	defer o.RUnlock()

	return o.hasErrors
}

// the golang test runner throws a severe panic if trying to fail a test after it already passed
// this happens for example on t.Run where a goroutine logs an Error (which fails the test) after t.Run passed
// the solution is to add "defer testOutput.TestTerminated()" to execute as the t.Run body is returning
func (o *TestOutput) TestTerminated() {
	o.Lock()
	defer o.Unlock()

	o.testTerminated = true
}

// assumes write lock (o.Lock())
func (o *TestOutput) recordError(line string) {
	defer func() {
		if p := recover(); p != nil {
			// a known panic is when we try to fail from a goroutine a test that already passed
			fmt.Println(TEST_RUNNER_PANIC_ERROR, o.tb.Name(), ":", p, ":", line)
		}
	}()

	o.hasErrors = true
	if !o.testTerminated {

		o.tb.Error(line)
		o.tb.Error(TEST_FAILED_ERROR)

	} else {
		o.tb.Fail()
		// must use print because after test is terminated its t.Log does not output anything
		fmt.Println(POST_TERMINATED_ERROR, o.tb.Name(), ":", line)

	}
}

func (o *TestOutput) isLoggingDisabled() bool {
	return atomic.LoadUint32(&o.loggingDisabled) != 0
}

func (o *TestOutput) disableLogging() {
	atomic.StoreUint32(&o.loggingDisabled, 1)
}

// func (o *TestOutput) Append(level string, message string, fields ...*Field) moved to file t.go
