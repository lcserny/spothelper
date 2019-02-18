package main

import (
	"flag"
	. "github.com/lcserny/spothelper"
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
	case UNKNOWN:
		log.Fatal("Unknown command given")
	}

	endTime := MakeTimestamp() - startTime
	log.Printf("FINISHED: it took %d ms to execute program!", endTime)
}

func oldMain() {
	//fmt.Printf(stringutil.Reverse("\n!oG ,olleH"))
	ReadFile("files/tmpFile")
}