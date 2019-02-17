package spothelper

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
)

type GlobalConfig struct {
	Host       string   `json:"host"`
	Root       string   `json:"root"`
	BackupRoot string   `json:"backupRoot"`
	Sites      []string `json:"sites"`
}

func ProcessUnused(spotVersionsFile string, globalConfigFile string, secondsBetween int, clusterResourcesFile string) {
	log.Println(spotVersionsFile)
	spotVersionsMap := getVersions(spotVersionsFile)
	for k, v := range spotVersionsMap {
		log.Printf("key=%s val=%d\n", k, v)
	}

	log.Println(globalConfigFile)
	globalConfigMap := getGlobalConfigMap(globalConfigFile)
	for k, v := range globalConfigMap {
		log.Printf("key=%s val=%#v\n", k, v)
	}

	log.Println(secondsBetween)

	log.Println(clusterResourcesFile)
	unusedResources := getUnusedResources(clusterResourcesFile, spotVersionsMap, globalConfigMap, secondsBetween)
	for i, ele := range unusedResources {
		log.Printf("unusedResource#%d: %s", i, ele)
	}
}

func getUnusedResources(clusterResourcesFile string, versionsMap map[string]int,
	configs map[string]GlobalConfig, secondsBetween int) []string {
	allBytes, err := ioutil.ReadFile(clusterResourcesFile)
	CheckError(err)

	var unusedResources []string

	var jsonClusterResourcesMap map[string]string
	err = json.Unmarshal(allBytes, &jsonClusterResourcesMap)
	CheckError(err)
	for k := range jsonClusterResourcesMap {
		unusedResources = append(unusedResources, k)
	}

	// TODO: leave ONLY unusedResources

	return unusedResources
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
