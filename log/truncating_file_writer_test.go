// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

const string1 = "foo"
const string2 = "bar"
const string3 = "buz"

func closeSilently(f *os.File) {
	_ = f.Close()
}

func testFileContents(t *testing.T, filename string, expected string) {
	f, _ := os.Open(filename)
	defer closeSilently(f)

	contents, _ := ioutil.ReadAll(f)
	require.EqualValues(t, []byte(expected), contents)
}


func TestNewTruncatingFileWriterWithNoDefault(t *testing.T) {
	tmp, err := ioutil.TempFile("/tmp", "truncatingFileWriter")
	require.NoError(t, err)
	defer closeSilently(tmp)

	w := NewTruncatingFileWriter(tmp)
	_, _ = w.Write([]byte(string1))
	testFileContents(t, tmp.Name(), string1)

	_ = w.Truncate()

	_, _ = w.Write([]byte(string2))
	testFileContents(t, tmp.Name(), string2)
}

func TestNewTruncatingFileWriterWithAutoTruncate(t *testing.T) {
	tmp, err := ioutil.TempFile("/tmp", "truncatingFileWriter")
	require.NoError(t, err)
	defer closeSilently(tmp)

	w := NewTruncatingFileWriter(tmp, 1*time.Millisecond)
	_, _ = w.Write([]byte(string1))
	testFileContents(t, tmp.Name(), string1)

	time.Sleep(1 * time.Millisecond)

	_, _ = w.Write([]byte(string2))
	testFileContents(t, tmp.Name(), string2)

	time.Sleep(1 * time.Millisecond)

	_, _ = w.Write([]byte(string3))
	testFileContents(t, tmp.Name(), string3)
}

func TestNewTruncatingFileWriterDoesNotTruncateBeforeTimeoutElapsed(t *testing.T) {
	tmp, err := ioutil.TempFile("/tmp", "truncatingFileWriter")
	require.NoError(t, err)
	defer closeSilently(tmp)

	w := NewTruncatingFileWriter(tmp, 1*time.Second)
	_, _ = w.Write([]byte(string1))
	testFileContents(t, tmp.Name(), string1)

	time.Sleep(1 * time.Millisecond)

	_, _ = w.Write([]byte(string2))
	testFileContents(t, tmp.Name(), string1 + string2)

	time.Sleep(1 * time.Millisecond)

	_, _ = w.Write([]byte(string3))
	testFileContents(t, tmp.Name(), string1 + string2 + string3)
}
