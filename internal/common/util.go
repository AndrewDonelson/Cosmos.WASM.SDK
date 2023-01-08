package common

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// PrettyPrint is used to display any type nicely in the log output
func PrettyPrint(v interface{}) string {

	name := GetType(v)
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("Dump of [%s]:\n%s\n", name, string(b))
}

// GetType will return the name of the provided interface using reflection
func GetType(i interface{}) string {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}

	return t.Name()
}

func SplitStrBySize(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}

	if chunkSize <= 0 {
		chunkSize = DEFAULT_LINE_LENGTH
	}

	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, "\n", "", -1)

	//if chunkSize >= len(s) {
	if len(s) <= chunkSize {
		return []string{s}
	}

	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}
