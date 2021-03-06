// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"github.com/orbs-network/go-mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func onErrorStub(err error) {}

func TestTestOutput_Append_NoRacesInMultipleConcurrentCalls(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})
	m.When("Log", "foobarbaz").Times(10)

	wait := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			o.Append(onErrorStub, "error", "foobarbaz")
			wait <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-wait
	}
}

func TestTestOutputLogsToTLog(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})
	m.When("Log", "foo").Times(1)

	o.Append(onErrorStub, "info", "foo")

	_, err := m.Verify()
	require.NoError(t, err)
}

func TestOutputLogsUnAllowedErrorToTLogAsErrorAndStopsLogging(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})
	m.When("Error", "foo").Times(1)
	m.When("Error", TEST_FAILED_ERROR).Times(1)
	m.Never("Log", "bar")

	o.Append(onErrorStub, "error", "foo")
	o.Append(onErrorStub, "info", "bar")

	_, err := m.Verify()
	require.NoError(t, err)
	require.True(t, o.HasErrors())
}

func TestOutputLogsAllowedErrorToTLogAsInfo(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})

	o.AllowErrorsMatching("foo")
	m.When("Log", "foo").Times(1)

	o.Append(onErrorStub, "error", "foo")

	_, err := m.Verify()
	require.NoError(t, err)
	require.False(t, o.HasErrors())
}

func TestOutputStopsRecordingErrorsAfterTestTerminated(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})

	m.When("Error", "foo").Times(0)

	o.TestTerminated()
	o.Append(onErrorStub, "error", "foo")

	_, err := m.Verify()
	require.NoError(t, err)
}

func TestOutputRecoversFromTestRunnerPanicsDuringRecordError(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})

	m.When("Error", "foo").Call(func(string) {
		panic("test runner panic")
	})

	require.NotPanics(t, func() {
		o.Append(onErrorStub, "error", "foo")
	})
}

func TestOutputFailsTestEvenAfterTermination(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})

	m.When("Error", "foo").Times(0)
	m.When("Fail").Times(1)

	o.TestTerminated()
	o.Append(onErrorStub, "error", "foo")

	_, err := m.Verify()
	require.NoError(t, err)
}

// designed for race detector
func TestOutputSynchronizesHasErrorsAndAppendError(t *testing.T) {
	m := &fakeTLog{}
	o := NewTestOutput(m, nopFormatter{})
	ch := make(chan struct{})
	go func() {
		for i := 1; i <= 100; i++ {
			o.Append(onErrorStub, "error", "foo")
		}
		ch <- struct{}{}
	}()
	go func() {
		for i := 1; i <= 100; i++ {
			o.HasErrors()
		}
		ch <- struct{}{}
	}()
	<-ch
	<-ch
}

type fakeTLog struct {
	mock.Mock
}

func (t *fakeTLog) Error(args ...interface{}) {
	t.Called(args...)
}

func (t *fakeTLog) Fatal(args ...interface{}) {
	t.Called(args...)
}

func (t *fakeTLog) Log(args ...interface{}) {
	t.Called(args...)
}

func (t *fakeTLog) Fail() {
	t.Called()
}

func (t *fakeTLog) Name() string {
	return "FakeTestName"
}

type nopFormatter struct {
}

func (nopFormatter) FormatRow(timestamp time.Time, level string, message string, params ...*Field) (formattedRow string) {
	return message
}
