package spothelper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	UNUSED_PREF        = "UNUSED"
	MISC_PREF          = "MISC_UNUSED"
	DEL_COMMANDS_PREF  = "DELETE_CMD"
	BACK_COMMANDS_PREF = "BACKUP_CMD"
)

type GlobalConfig struct {
	Host       string   `json:"host"`
	Root       string   `json:"root"`
	BackupRoot string   `json:"backupRoot"`
	Sites      []string `json:"sites"`
}

func ProcessUnused(spotVersionsFile string, globalConfigFile string, inFolder string, outFolder string) {
	log.Println("spotVersionsFile:", spotVersionsFile)
	log.Println("globalConfigFile:", globalConfigFile)
	log.Println("inFolder:", inFolder)
	log.Println("outFolder:", outFolder)

	spotVersionsMap := getVersions(spotVersionsFile)
	globalConfigMap := getGlobalConfigMap(globalConfigFile)
	outFolderPath := getOutFolder(outFolder)
	writeFiles(spotVersionsMap, globalConfigMap, inFolder, outFolderPath)
}

func getOutFolder(outFolder string) string {
	today := time.Now().Format("2006-01-02")
	count := 0
	newOutFolder := filepath.Join(outFolder, fmt.Sprintf("%s_%d", today, count))
	for {
		if _, err := os.Stat(newOutFolder); os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Join(newOutFolder, UNUSED_PREF), os.ModePerm)
			CheckError(err)
			err = os.MkdirAll(filepath.Join(newOutFolder, MISC_PREF), os.ModePerm)
			CheckError(err)
			err = os.MkdirAll(filepath.Join(newOutFolder, DEL_COMMANDS_PREF), os.ModePerm)
			CheckError(err)
			err = os.MkdirAll(filepath.Join(newOutFolder, BACK_COMMANDS_PREF), os.ModePerm)
			CheckError(err)
			log.Printf("Created output folders in: %s", newOutFolder)
			return newOutFolder
		} else {
			count++
			newOutFolder = filepath.Join(outFolder, fmt.Sprintf("%s_%d", today, count))
		}
	}
}

func writeFiles(versionsMap map[string]int, configs map[string]GlobalConfig, inFolder string, outFolder string) {
	err := filepath.Walk(inFolder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			allBytes, err := ioutil.ReadFile(path)
			CheckError(err)
			var jsonClusterResourcesMap map[string]string
			err = json.Unmarshal(allBytes, &jsonClusterResourcesMap)
			CheckError(err)

			unusedResources := getUnusedResources(jsonClusterResourcesMap, versionsMap, configs)
			writeLinesToFile(unusedResources, filepath.Join(outFolder, fmt.Sprintf("UNUSED_%s", info.Name())))

			// TODO: add other files to write
		}
		return nil
	})
	CheckError(err)
}

func writeLinesToFile(slice []string, fileName string) {
	file, err := os.Create(fileName)
	defer CloseFile(file)
	CheckError(err)

	for _, val := range slice {
		_, err := fmt.Fprintln(file, val)
		CheckError(err)
	}
	log.Printf("Done writing file: %s", fileName)
}

func getUnusedResources(resources map[string]string, versions map[string]int, configs map[string]GlobalConfig) []string {

}

func getGlobalConfigMap(globalConfigFile string) map[string]GlobalConfig {
	allBytes, err := ioutil.ReadFile(globalConfigFile)
	CheckError(err)

	resultMap := make(map[string]GlobalConfig)

	var jsonData []map[string]GlobalConfig
	err = json.Unmarshal(allBytes, &jsonData)
	CheckError(err)
	for _, element := range jsonData {
		for s, gc := range element {
			resultMap[s] = gc
		}
	}

	return resultMap
}

func getVersions(spotVersionsFile string) map[string]int {
	openedFile, err := os.Open(spotVersionsFile)
	defer CloseFile(openedFile)
	CheckError(err)

	versionsMap := make(map[string]int)

	exp, err := regexp.Compile("^\\s*(?P<name>[a-zA-Z]+)\\s+->\\s+[vV](?P<version>[0-9]{1,2})\\s*$")
	CheckError(err)
	scanner := bufio.NewScanner(openedFile)
	for scanner.Scan() {
		match := exp.FindStringSubmatch(scanner.Text())
		resultMap := make(map[string]string)
		for i, name := range exp.SubexpNames() {
			if i != 0 && name != "" {
				resultMap[name] = match[i]
			}
		}

		version, err := strconv.ParseInt(resultMap["version"], 0, 32)
		CheckError(err)
		versionsMap[resultMap["name"]] = int(version)
	}

	return versionsMap
}
