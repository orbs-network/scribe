// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"time"
)

// this file is part of test_output.go
// a file with short name (t.go) to make the testLogger prefix less annoying

func (o *TestOutput) Append(onError func(err error), level string, message string, fields ...*Field) {
	// take write lock because calling o.recordError()
	o.Lock()
	defer o.Unlock()

	// we use this mechanism to stop logging new log lines after the test failed from a different goroutine
	if o.isLoggingDisabled() {
		return
	}

	logLine := o.formatter.FormatRow(time.Now(), level, message, fields...)

	if level == "error" && !o.allowed(message, fields) {
		o.disableLogging()
		o.recordError(logLine)
	} else {
		o.tb.Log(logLine)
	}

}
