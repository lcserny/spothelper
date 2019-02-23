package main

import (
	. "../../src"
	"flag"
	. "github.com/lcserny/goutils"
	"strconv"
	"strings"

	//"github.com/lcserny/stringutil"
	"log"
	"os"
)

type Command int

const (
	UNKNOWN Command = -1
	UNUSED  Command = 0
	BACKUP  Command = 1
	DELETE  Command = 2
)

var commandFlag = flag.String("command", "unused", "Please provide a command like: UNUSED")

func init() {
	flag.Parse()
}

func main() {
	startTime := MakeTimestamp()

	args := os.Args[1:]
	argsLength := len(args)
	command := newCommandFrom(*commandFlag)
	switch command {
	case UNUSED:
		incr := 0
		if argsLength < 4 {
			log.Fatal("Please provide args: spotVersionsFile, globalConfigFile, inFolder and outFolder")
		} else if argsLength > 4 {
			incr = 1
		}
		ProcessUnused(args[0+incr], args[1+incr], args[2+incr], args[3+incr])
		break
	case BACKUP:
		if argsLength < 4 {
			log.Fatal("Please provide args: backupCommandsFile, secondsBetween, startOffset and limit")
		}
		secondsBetween, err := strconv.ParseInt(args[2], 0, 32)
		LogFatal(err)
		startOffset, err := strconv.ParseInt(args[3], 0, 32)
		LogFatal(err)
		limit, err := strconv.ParseInt(args[4], 0, 32)
		LogFatal(err)
		ExecuteBackup(args[1], int(secondsBetween), int(startOffset), int(limit))
		break
	case DELETE:
		if argsLength < 4 {
			log.Fatal("Please provide args: deleteCommandsFile, secondsBetween, startResource and limitResource")
		}
		secondsBetween, err := strconv.ParseInt(args[2], 0, 32)
		LogFatal(err)
		ExecuteDelete(args[1], int(secondsBetween), args[3], args[4])
		break
	case UNKNOWN:
		log.Fatal("Unknown command given")
	}

	endTime := MakeTimestamp() - startTime
	log.Printf("FINISHED: it took %d ms to execute program!", endTime)
}

func newCommandFrom(val string) Command {
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
