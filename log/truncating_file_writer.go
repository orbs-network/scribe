// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type truncatingFileWriter struct {
	f        *os.File
	interval time.Duration

	mu                    sync.Mutex
	lastTruncatedUnixNano int64
}

func NewTruncatingFileWriter(f *os.File, intervals ...time.Duration) *truncatingFileWriter {
	interval := time.Duration(0)

	if len(intervals) > 0 {
		interval = intervals[0]
	}

	return &truncatingFileWriter{
		f:                     f,
		interval:              interval,
		lastTruncatedUnixNano: time.Now().UnixNano(),
	}
}

func (w *truncatingFileWriter) Write(p []byte) (n int, err error) {
	now := time.Now()
	lastTruncated := atomic.LoadInt64(&w.lastTruncatedUnixNano)
	if shouldTruncate(now, lastTruncated, w.interval) {
		func () { // only for defer scope
			w.mu.Lock()
			defer w.mu.Unlock()

			lastTruncated = atomic.LoadInt64(&w.lastTruncatedUnixNano)
			if shouldTruncate(now, lastTruncated, w.interval) { // check again under lock
				err := w.Truncate()
				if err != nil {
					return // TODO log the failure to truncate?
				}
				atomic.StoreInt64(&w.lastTruncatedUnixNano, now.UnixNano())
			}
		}()
	}

	return w.f.Write(p)
}

func (w *truncatingFileWriter) Truncate() error {
	if err := w.f.Truncate(0); err != nil {
		return err
	} else {
		w.f.Seek(0, 0)
	}

	return nil
}

func shouldTruncate(now time.Time, lastTruncatedUnixNano int64, interval time.Duration) bool {
	return interval.Nanoseconds() > 0 && now.UnixNano() - lastTruncatedUnixNano >= interval.Nanoseconds()
}
