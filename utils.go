package spothelper

import (
	"log"
	"strings"
	"time"
)

type Command int

const (
	UNKNOWN Command = -1
	UNUSED  Command = 0
)

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MewCommandFrom(val string) Command {
	switch strings.ToUpper(val) {
	case "UNUSED":
		return UNUSED
	}
	return UNKNOWN
}

func CheckError(e error) {
	if e != nil {
		log.Fatalf("ERROR: %s", e)
	}
}
