// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"fmt"
	"io"
	"time"
)

type basicOutput struct {
	formatter LogFormatter
	writer    io.Writer
	filters   and
}

func (out *basicOutput) SetFilters(filters ...Filter) {
	out.filters = and{filters}
}

func (out *basicOutput) Append(onError func(err error), level string, message string, fields ...*Field) {
	if !out.filters.Allows(level, message, fields) {
		return
	}

	logLine := out.formatter.FormatRow(time.Now(), level, message, fields...)
	_, err := fmt.Fprintln(out.writer, logLine)
	if err != nil {
		onError(err)
	}
}

func NewFormattingOutput(writer io.Writer, formatter LogFormatter) Output {
	return &basicOutput{formatter: formatter, writer: writer}
}
