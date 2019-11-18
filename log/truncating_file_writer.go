// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"io"
	"os"
	"sync"
	"time"
)

type truncatingFileWriter struct {
	f        *os.File
	interval time.Duration

	mu            sync.RWMutex
	lastTruncated time.Time
}

type TruncatingFileWriter interface {
	io.Writer
	Truncate() error
}

func NewTruncatingFileWriter(f *os.File, intervals ...time.Duration) TruncatingFileWriter {
	interval := time.Duration(0)

	if len(intervals) > 0 {
		interval = intervals[0]
	}

	return &truncatingFileWriter{
		f:             f,
		interval:      interval,
		lastTruncated: time.Now(),
	}
}

func (w *truncatingFileWriter) Write(p []byte) (n int, err error) {
	now := time.Now()
	lastTruncated := w.getLastTruncatedUnixNano()
	if w.interval.Nanoseconds() > 0 && (shouldTruncate(now, lastTruncated, w.interval)) {
		w.autoTruncate(now)
	}

	return w.f.Write(p)
}

func (w *truncatingFileWriter) getLastTruncatedUnixNano() time.Time {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastTruncated
}

func (w *truncatingFileWriter) autoTruncate(now time.Time) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if shouldTruncate(now, w.lastTruncated, w.interval) { // check again under lock
		err := w.Truncate()
		if err != nil {
			return err
		}
		w.lastTruncated = now
	}

	return nil
}

func (w *truncatingFileWriter) Truncate() error {
	if err := w.f.Truncate(0); err != nil {
		return err
	} else {
		w.f.Seek(0, 0)
	}

	return nil
}

func shouldTruncate(now time.Time, lastTruncated time.Time, interval time.Duration) bool {
	return interval.Nanoseconds() > 0 && now.UnixNano()-lastTruncated.UnixNano() >= interval.Nanoseconds()
}
