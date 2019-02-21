package spothelper

import (
	. "github.com/lcserny/goutils"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ExecuteBackup(backupCommandsFile string, secondsBetween int, startOffset int, limit int) {
	lines := ReadFileToLines(backupCommandsFile)
	for i := startOffset; i < MinInt(len(lines), limit); i++ {
		execute(secondsBetween, i, lines[i])
	}
}

func ExecuteDelete(deleteCommandsFile string, secondsBetween int, startResource string, limitResource string) {
	lines := ReadFileToLines(deleteCommandsFile)
	startIndex := 0
	limit := len(lines)

	for i := 0; i < limit; i++ {
		command := lines[i]
		if strings.Contains(command, startResource) {
			startIndex = i
			log.Printf("startIndex set to: %d\n", startIndex)
		}
		if strings.Contains(command, limitResource) {
			limit = i + 1
			log.Printf("limit set to: %d\n", limit)
		}
	}

	for i := startIndex; i < limit; i++ {
		execute(secondsBetween, i, lines[i])
	}
}

func execute(secondsBetween int, index int, command string) {
	log.Printf("Running command #%d: \"%s\"", index, command)

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	LogFatal(err)

	time.Sleep(time.Duration(secondsBetween) * time.Second)
}
