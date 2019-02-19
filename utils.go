package spothelper

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type Command int

const (
	UNKNOWN Command = -1
	UNUSED  Command = 0
	BACKUP  Command = 1
	DELETE  Command = 2
)

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MewCommandFrom(val string) Command {
	switch strings.ToUpper(val) {
	case "UNUSED":
		return UNUSED
	case "BACKUP":
		return BACKUP
	case "DELETE":
		return DELETE
	}
	return UNKNOWN
}

func CheckError(e error) {
	if e != nil {
		log.Fatalf("ERROR: %#v", e)
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

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ReadFileToLines(file string) []string {
	openFile, err := os.Open(file)
	defer CloseFile(openFile)
	CheckError(err)

	var lines []string
	scanner := bufio.NewScanner(openFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
