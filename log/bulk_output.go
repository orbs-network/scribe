// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"sync"
	"time"
)

type row struct {
	level     string
	timestamp time.Time
	message   string
	fields    []*Field
}

type bulkOutput struct {
	formatter LogFormatter
	writer    io.Writer

	bulkSize int

	lock    *sync.Mutex
	logs    []*row
	filters and
}

func (out *bulkOutput) SetFilters(filters ...Filter) {
	out.filters = and{filters}
}

func (out *bulkOutput) Append(onError func(err error), level string, message string, fields ...*Field) {
	if !out.filters.Allows(level, message, fields) {
		return
	}
	row := &row{level, time.Now(), message, fields}

	out.lock.Lock()
	out.logs = append(out.logs, row)
	out.lock.Unlock()

	out.flushIfNeeded(onError)
}

func (out *bulkOutput) flushIfNeeded(onError func(err error)) {
	out.lock.Lock()
	defer out.lock.Unlock()

	if len(out.logs) >= out.bulkSize {
		b := new(bytes.Buffer)

		for _, row := range out.logs {
			b.Write([]byte(out.formatter.FormatRow(row.timestamp, row.level, row.message, row.fields...)))
			b.Write([]byte("\n"))
		}

		out.logs = nil

		go func() {
			if n, err := out.writer.Write(b.Bytes()); err != nil {
				onError(err)
				fmt.Println(fmt.Sprintf("%s failed to send logs via http, %d bytes lost: %s", time.Now().String(), n, err))
			}
		}()
	}
}

func NewBulkOutput(writer io.Writer, formatter LogFormatter, bulkSize int) Output {
	if bulkSize > 1000 {
		panic(fmt.Sprintf("bulk size can't be greater than 1000, please refer to this issue for explanation: https://github.com/orbs-network/orbs-network-go/issues/501"))
	}

	return &bulkOutput{
		formatter: formatter,
		writer:    writer,
		bulkSize:  bulkSize,
		lock:      &sync.Mutex{},
	}
}

type httpWriter struct {
	url        string
	httpClient *http.Client
}

func (w *httpWriter) Write(p []byte) (n int, err error) {
	reader := bytes.NewReader(p)
	size := reader.Len()
	resp, err := w.httpClient.Post(w.url, "application/json", reader)

	if err != nil {
		return 0, errors.Errorf("Failed to send logs: %s", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, errors.Errorf("Failed to send logs: %d, %s", resp.StatusCode, err)
	}

	return size, err
}

func NewHttpWriter(url string) *httpWriter {
	return NewHttpWriterWithTimeout(url, time.Second*60)
}

func NewHttpWriterWithTimeout(url string, timeout time.Duration) *httpWriter {
	return &httpWriter{url, &http.Client{Timeout: timeout}}
}
