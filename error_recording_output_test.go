package log

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestErrorRecordingOutput_IgnoresNonError(t *testing.T) {
	o := NewErrorRecordingOutput([]string{})
	o.Append("info", "foo")

	require.False(t, o.HasErrors(), "info was recorded")
}

func TestErrorRecordingOutput_IgnoresAllowedError(t *testing.T) {
	o := NewErrorRecordingOutput([]string{"foo"})
	o.Append("error", "foo")

	require.False(t, o.HasErrors(), "allowed error was recorded")
}

func TestErrorRecordingOutput_RecordsDisallowedError(t *testing.T) {
	o := NewErrorRecordingOutput([]string{"foo"})
	e := errors.Errorf("foo error")
	o.Append("error", "bar", Error(e))

	require.True(t, o.HasErrors(), "disallowed error was not recorded")
	require.Contains(t, o.GetUnexpectedErrors(), "bar (passed Error object: foo error)")
}

