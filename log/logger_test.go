// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	TransactionFlow     = "TransactionFlow"
	TransactionAccepted = "Transaction accepted"
)

func parseOutput(input string) map[string]interface{} {
	jsonMap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(input), &jsonMap)
	return jsonMap
}

func TestBasicLogger_WithTags_ClonesLoggerFully(t *testing.T) {
	v1 := String("k1", "v1")
	v2 := String("c1", "v2")
	v3 := String("c2", "v3")

	parent := GetLogger(v1)
	child1 := parent.WithTags(v2)
	child2 := parent.WithTags(v3)

	require.ElementsMatch(t, []*Field{v1}, parent.Tags())
	require.ElementsMatch(t, []*Field{v1, v2}, child1.Tags())
	require.ElementsMatch(t, []*Field{v1, v3}, child2.Tags())
}

func TestSimpleLogger(t *testing.T) {
	b := new(bytes.Buffer)
	GetLogger(Node("node1"), Int("vcid", 999), Service("public-api")).WithOutput(NewFormattingOutput(b, NewJsonFormatter())).Info("Service initialized")

	jsonMap := parseOutput(b.String())

	require.Equal(t, "info", jsonMap["level"])
	require.Equal(t, "node1", jsonMap["node"])
	require.Equal(t, 999.0, jsonMap["vcid"]) // because golang JSON parser decodes ints as float64
	require.Equal(t, "public-api", jsonMap["service"])
	require.Equal(t, "log.TestSimpleLogger", jsonMap["function"])
	require.Equal(t, "Service initialized", jsonMap["message"])
	require.Regexp(t, "log/logger_test.go", jsonMap["source"])
	require.NotNil(t, jsonMap["timestamp"])
}

func TestSimpleLogger_AggregateField(t *testing.T) {
	aggregatedFields := Aggregate("music", String("Artist", "Iggy Pop"), String("Song", "Passenger"))
	b := new(bytes.Buffer)
	GetLogger().
		WithOutput(NewFormattingOutput(b, NewJsonFormatter())).
		Info("bar", aggregatedFields)

	jsonMap := parseOutput(b.String())

	require.Equal(t, "Iggy Pop", jsonMap["Artist"])
	require.Equal(t, "Passenger", jsonMap["Song"])
}

func TestSimpleLogger_AggregateField_NestedLogger(t *testing.T) {
	aggregatedFields := Aggregate("music", String("Artist", "Iggy Pop"), String("Song", "Passenger"))
	b := new(bytes.Buffer)
	GetLogger(String("k1", "v1")).
		WithTags(aggregatedFields).
		WithOutput(NewFormattingOutput(b, NewJsonFormatter())).
		Info("bar")

	jsonMap := parseOutput(b.String())

	require.Equal(t, "Iggy Pop", jsonMap["Artist"])
	require.Equal(t, "Passenger", jsonMap["Song"])
}

func TestBasicLogger_WithFilter(t *testing.T) {
	b := new(bytes.Buffer)
	GetLogger().WithOutput(NewFormattingOutput(b, NewJsonFormatter())).
		WithFilters(OnlyErrors()).
		Info("foo")
	require.Empty(t, b.String(), "output was not empty")
}

func TestNestedLogger(t *testing.T) {
	b := new(bytes.Buffer)

	txId := String("txId", "1234567")
	txFlowLogger := GetLogger().WithOutput(NewFormattingOutput(b, NewJsonFormatter())).WithTags(String("flow", TransactionFlow))
	txFlowLogger.Info(TransactionAccepted, txId, Bytes("payload", []byte{1, 2, 3, 99, 250}))

	jsonMap := parseOutput(b.String())

	require.Equal(t, TransactionAccepted, jsonMap["message"])
	require.Equal(t, "1234567", jsonMap["txId"])
	require.Equal(t, TransactionFlow, jsonMap["flow"])
	require.Equal(t, "01020363fa", jsonMap["payload"])

}

func TestStringableSlice(t *testing.T) {
	b := new(bytes.Buffer)
	var receipts = []stringable{{"hello"}, {"darkness"}}

	GetLogger().WithOutput(NewFormattingOutput(b, NewJsonFormatter())).Info("StringableSlice test", StringableSlice("a-collection", receipts))

	jsonMap := parseOutput(b.String())

	require.Equal(t, []interface{}{"hello", "darkness"}, jsonMap["a-collection"])
}

func TestCustomLogFormatter(t *testing.T) {
	b := new(bytes.Buffer)
	serviceLogger := GetLogger(Node("node1"), Service("public-api")).
		WithOutput(NewFormattingOutput(b, NewHumanReadableFormatter())).WithSourcePrefix("scribe/")
	serviceLogger.Info("Service initialized",
		Int("some-int-value", 12),
		Int("block-height", 9999),
		Bytes("bytes", []byte{2, 3, 99}),
		Stringable("vchainId", stringable{"123"}),
		String("_test-id", "hello"),
		String("_underscore", "wow"))

	out := b.String()

	require.Regexp(t, "^info", out)
	require.Regexp(t, "Service initialized", out)
	require.Regexp(t, "node=node1", out)
	require.Regexp(t, "service=public-api", out)
	require.Regexp(t, "block-height=9999", out)
	require.Regexp(t, "vchainId=123", out)
	require.Regexp(t, "bytes=020363", out)
	require.Regexp(t, "some-int-value=12", out)
	require.Regexp(t, "function=log.TestCustomLogFormatter", out)
	require.Regexp(t, "source=.*log/logger_test.go", out)
	require.Regexp(t, "_test-id=hello", out)
	require.Regexp(t, "_underscore=wow", out)
}

func TestHumanReadable_AggregateField(t *testing.T) {
	aggregatedFields := Aggregate("music", String("Artist", "Iggy Pop"), String("Song", "Passenger"))
	b := new(bytes.Buffer)
	GetLogger().
		WithOutput(NewFormattingOutput(b, NewHumanReadableFormatter())).
		Info("bar", aggregatedFields)

	out := b.String()
	require.Regexp(t, "Artist=Iggy Pop", out)
	require.Regexp(t, "Song=Passenger", out)

}

func TestHumanReadableFormatterFormatWithStringableSlice(t *testing.T) {
	b := new(bytes.Buffer)
	var receipts = []stringable{{"hello"}, {"darkness"}}

	GetLogger(Node("node1"), Service("public-api")).WithOutput(NewFormattingOutput(b, NewHumanReadableFormatter())).
		Info("StringableSlice HR test", StringableSlice("a-collection", receipts))

	out := b.String()

	require.Regexp(t, "a-collection=", out)
	require.Regexp(t, `"hello"`, out)
	require.Regexp(t, `"darkness"`, out)
}

func TestMultipleOutputs(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "logger_test_multiple_outputs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name()) // clean up

	fileOutput, _ := os.Create(tempFile.Name())

	b := new(bytes.Buffer)

	GetLogger(Node("node1"), Service("public-api")).WithOutput(NewFormattingOutput(b, NewJsonFormatter()), NewFormattingOutput(fileOutput, NewJsonFormatter())).
		Info("Service initialized")

	rawFile, _ := ioutil.ReadFile(tempFile.Name())
	fileContents := string(rawFile)

	checkOutput := func(output string) {
		jsonMap := parseOutput(output)

		require.Equal(t, "info", jsonMap["level"])
		require.Equal(t, "node1", jsonMap["node"])
		require.Equal(t, "public-api", jsonMap["service"])
		require.Equal(t, "log.TestMultipleOutputs", jsonMap["function"])
		require.Equal(t, "Service initialized", jsonMap["message"])
		require.NotEmpty(t, jsonMap["source"])
		require.NotNil(t, jsonMap["timestamp"])
	}

	checkOutput(b.String())
	checkOutput(fileContents)
}

func TestMultipleOutputsForMemoryViolationByHumanReadable(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "logger_test_multiple_outputs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name()) // clean up

	b := new(bytes.Buffer)

	fileOutput, _ := os.Create(tempFile.Name())

	require.NotPanics(t, func() {
		GetLogger(Node("node1"), Service("public-api")).WithOutput(NewFormattingOutput(b, NewHumanReadableFormatter()), NewFormattingOutput(fileOutput, NewJsonFormatter())).
			Info("Service initialized")
	})
}

func TestJsonFormatterWithCustomTimestampColumn(t *testing.T) {
	f := NewJsonFormatter().WithTimestampColumn("@timestamp")
	row := f.FormatRow(time.Now(), "info", "hello")

	require.Regexp(t, "@timestamp", row)
}

func Test_getCaller(t *testing.T) {
	l := &basicLogger{
		sourceRootPrefixIndex: getSourceRootPrefixIndex("scribe/"),
	}

	function, source := l.getCaller(2)

	require.Equal(t, "log.Test_getCaller", function)
	require.Regexp(t, "^log/logger_test.go:", source) // skipping line number because it will shift when this file is edited
}

func BenchmarkBasicLoggerInfoFormatters(b *testing.B) {
	collection := []string{"David Bowie", "Diamond Dogs"}
	formatters := []LogFormatter{NewHumanReadableFormatter(), NewJsonFormatter()}

	for _, formatter := range formatters {
		b.Run(reflect.TypeOf(formatter).String(), func(b *testing.B) {
			serviceLogger := GetLogger(Node("node1"), Service("public-api")).
				WithOutput(NewFormattingOutput(ioutil.Discard, NewJsonFormatter()))

			b.StartTimer()
			for i := 0; i < b.N; i++ {
				serviceLogger.Info("Benchmark test", StringableSlice("a-collection", collection))
			}
			b.StopTimer()
		})
	}

}

func BenchmarkBasicLoggerInfoWithDevNull(b *testing.B) {
	collection := []string{"David Bowie", "Diamond Dogs"}
	outputs := []io.Writer{os.Stdout, ioutil.Discard}

	for _, output := range outputs {
		b.Run(reflect.TypeOf(output).String(), func(b *testing.B) {

			serviceLogger := GetLogger(Node("node1"), Service("public-api")).
				WithOutput(NewFormattingOutput(output, NewHumanReadableFormatter()))

			b.StartTimer()
			for i := 0; i < b.N; i++ {
				serviceLogger.Info("Benchmark test", StringableSlice("a-collection", collection))
			}
			b.StopTimer()
		})
	}
}

type stringable struct {
	value string
}

func (s stringable) String() string {
	return s.value
}
