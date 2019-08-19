package log

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/orbs-network/gojay"
)

func TestMarhsallingALogLineCorrectly(t *testing.T) {
	v := logData{}

	v.level = "info"
	v.timestamp = &logTimeStamp{value: "1920-30-23", key: "timestamp"}
	v.message = "An idiomatic logline object"

	v.params = append(v.params, Int("intValue", 35))
	v.params = append(v.params, Int32("int32Value", 37))
	v.params = append(v.params, Int32("int64Value", 40))
	v.params = append(v.params, Uint("uintValue", 30))
	v.params = append(v.params, Uint32("uint32Value", 33))
	v.params = append(v.params, Uint64("uint64Value", 65))

	v.params = append(v.params, Float32("float32Value", float32(32.64)))
	v.params = append(v.params, Float64("float64Value", 64.32))

	v.params = append(v.params, StringableSlice("arrayOfStrings", []stringable{{"stranger"}, {"strings"}}))

	sb := strings.Builder{}
	enc := gojay.NewEncoder(&sb)
	defer enc.Release()

	if err := enc.Encode(&v); err != nil {
		log.Fatal("Failed encoding the logline", err)
	}

	var data map[string]interface{}

	jsonAsString := sb.String()
	fmt.Println(jsonAsString)

	_ = json.Unmarshal([]byte(jsonAsString), &data)

	require.Equal(t, "An idiomatic logline object", data["message"])
	require.Equal(t, float64(35), data["intValue"])
	require.Equal(t, float64(37), data["int32Value"])
	require.Equal(t, float64(40), data["int64Value"])
	require.Equal(t, float64(30), data["uintValue"])
	require.Equal(t, float64(33), data["uint32Value"])
	require.Equal(t, float64(65), data["uint64Value"])
	require.Equal(t, float64(32.63999938964844), data["float32Value"])
	require.Equal(t, float64(64.32), data["float64Value"])
	require.Equal(t, []interface{}([]interface{}{"stranger", "strings"}), data["arrayOfStrings"])
}

func TestFormatHumanReadableRow(t *testing.T) {
	f := NewHumanReadableFormatter()
	tm, err := time.Parse(TIMESTAMP_FORMAT, "2006-01-02T01:23:45.123456789Z")
	require.NoError(t, err)
	row := f.FormatRow(tm, "info", "foobar")
	require.Equal(t, "i 01:23:45.123456 foobar ", row)
}

