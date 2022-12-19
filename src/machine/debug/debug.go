package debug

import (
	"strconv"
	"strings"
)

type DebugRecord struct {
	Key   string
	Value interface{}
}

var (
	log      chan DebugRecord
	DebugLog <-chan DebugRecord
)

func init() {
	log = make(chan DebugRecord, 128)
	DebugLog = log
}

func Hex(b []byte) string {
	var builder strings.Builder
	for _, v := range b {
		s := "0" + strconv.FormatInt(int64(v), 16)
		builder.WriteString(s[len(s)-2 : len(s)])
	}
	return builder.String()
}

func Debug(key string, v ...interface{}) {
	log <- DebugRecord{key, v}
}
