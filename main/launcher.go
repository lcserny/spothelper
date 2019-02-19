package main

import (
	"flag"
	. "github.com/lcserny/spothelper"
	"strconv"

	//"github.com/lcserny/stringutil"
	"log"
	"os"
)

var commandFlag *string

func init() {
	commandFlag = flag.String("command", "unused", "Please provide a command like: UNUSED")
	flag.Parse()
}

func main() {
	startTime := MakeTimestamp()

	args := os.Args[1:]
	command := MewCommandFrom(*commandFlag)
	switch command {
	case UNUSED:
		if len(args) < 4 {
			log.Fatal("Please provide args: spotVersionsFile, globalConfigFile, inFolder and outFolder")
		}
		ProcessUnused(args[0], args[1], args[2], args[3])
		break
	case BACKUP:
		if len(args) < 4 {
			log.Fatal("Please provide args: backupCommandsFile, secondsBetween, startOffset and limit")
		}
		secondsBetween, err := strconv.ParseInt(args[1], 0, 32)
		CheckError(err)
		startOffset, err := strconv.ParseInt(args[2], 0, 32)
		CheckError(err)
		limit, err := strconv.ParseInt(args[3], 0, 32)
		CheckError(err)
		RunBackupCommands(args[0], int(secondsBetween), int(startOffset), int(limit))
		break
	case DELETE:
		if len(args) < 4 {
			log.Fatal("Please provide args: deleteCommandsFile, secondsBetween, startResource and limitResource")
		}
		secondsBetween, err := strconv.ParseInt(args[1], 0, 32)
		CheckError(err)
		RunDeleteCommands(args[0], int(secondsBetween), args[2], args[3])
		break
	case UNKNOWN:
		log.Fatal("Unknown command given")
	}

	endTime := MakeTimestamp() - startTime
	log.Printf("FINISHED: it took %d ms to execute program!", endTime)
}
