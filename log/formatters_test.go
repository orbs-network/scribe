package log

import (
	"encoding/json"
	"fmt"
	"github.com/francoispqt/gojay"
	"github.com/stretchr/testify/require"
	"log"
	"strings"
	"testing"
)

func TestMarhsallingALogLineCorrectly(t *testing.T) {
	v := make(logLine)
	v["stringValue"] = "An idiomatic logline object"
	v["intValue"] = 35
	v["int32Value"] = int32(37)
	v["int64Value"] = int64(40)

	v["uintValue"] = uint(30)
	v["uint32Value"] = uint32(33)
	v["uint64Value"] = uint64(65)

	v["float32Value"] = float32(32.64)
	v["float64Value"] = float32(64.32)

	v["arrayOfStrings"] = []string{"stranger", "strings"}

	sb := strings.Builder{}
	enc := gojay.NewEncoder(&sb)
	defer enc.Release()

	if err := enc.Encode(v); err != nil {
		log.Fatal("Failed encoding the logline")
	}

	var data map[string]interface{}

	jsonAsString := sb.String()
	fmt.Println(jsonAsString)

	_ = json.Unmarshal([]byte(jsonAsString), &data)

	require.Equal(t, "An idiomatic logline object", data["stringValue"])
	require.Equal(t, float64(35), data["intValue"])
	require.Equal(t, float64(37), data["int32Value"])
	require.Equal(t, float64(40), data["int64Value"])
	require.Equal(t, float64(30), data["uintValue"])
	require.Equal(t, float64(33), data["uint32Value"])
	require.Equal(t, float64(65), data["uint64Value"])
	require.Equal(t, float64(32.64), data["float32Value"])
	require.Equal(t, float64(64.32), data["float64Value"])
	require.Equal(t, []interface {}([]interface {}{"stranger", "strings"}), data["arrayOfStrings"])
}
