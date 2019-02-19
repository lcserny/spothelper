package spothelper

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func RunBackupCommands(backupCommandsFile string, secondsBetween int, startOffset int, limit int) {
	lines := ReadFileToLines(backupCommandsFile)
	for i := startOffset; i < MinInt(len(lines), limit); i++ {
		runProcess(secondsBetween, i, lines[i])
	}
}

func RunDeleteCommands(deleteCommandsFile string, secondsBetween int, startResource string, limitResource string) {
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
		runProcess(secondsBetween, i, lines[i])
	}
}

func runProcess(secondsBetween int, index int, command string) {
	log.Printf("Running command #%d: \"%s\"", index, command)

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	CheckError(err)

	time.Sleep(time.Duration(secondsBetween) * time.Second)
}
