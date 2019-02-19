package spothelper

import (
	"log"
	"regexp"
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

func GetRegexSubgroups(exp *regexp.Regexp, text string) map[string]string {
	match := exp.FindStringSubmatch(text)
	resultMap := make(map[string]string)
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			resultMap[name] = match[i]
		}
	}
	return resultMap
}

func StringsContain(strings []string, match string) bool {
	for _, ele := range strings {
		if ele == match {
			return true
		}
	}
	return false
}
