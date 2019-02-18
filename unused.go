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
	"strings"
	"time"
)

const (
	UNUSED_PREF        = "UNUSED"
	MISC_PREF          = "MISC_UNUSED"
	DEL_COMMANDS_PREF  = "DELETE_CMD"
	BACK_COMMANDS_PREF = "BACKUP_CMD"

	GLOBAL_PATTERN = "^(?P<name>.*)/[vV](?P<version>[0-9]{1,2})/(?P<file>[^/]+)$"
	SITE_PATTERN   = "^(?P<site>.*)/(?P<name>.*)/[vV](?P<version>[0-9]{1,2})/(?P<file>[^/]+)$"
	LOCALE_PATTERN = "^(?P<site>.*)/(?P<locale>.*)/(?P<name>.*)/[vV](?P<version>[0-9]{1,2})/(?P<file>[^/]+)$"
)

var globalPattern = regexp.MustCompile(GLOBAL_PATTERN)
var sitePattern = regexp.MustCompile(SITE_PATTERN)
var localePattern = regexp.MustCompile(LOCALE_PATTERN)

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
	newOutFolder := generateNewOutFolderPath(outFolder, today, count)
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
			newOutFolder = generateNewOutFolderPath(outFolder, today, count)
		}
	}
}

func generateNewOutFolderPath(outFolder string, date string, count int) string {
	return filepath.Join(outFolder, fmt.Sprintf("%s_%d", date, count))
}

func writeFiles(versionsMap map[string]int, configs map[string]GlobalConfig, inFolder string, outFolder string) {
	err := filepath.Walk(inFolder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			allBytes, err := ioutil.ReadFile(path)
			CheckError(err)
			var jsonClusterResourcesMap map[string]string
			err = json.Unmarshal(allBytes, &jsonClusterResourcesMap)
			CheckError(err)
			clusterResourcesList := getClusterResourcesList(jsonClusterResourcesMap)

			spotInstance := getSpotInstanceName(info.Name())

			unusedResources := getUnusedResources(clusterResourcesList, versionsMap, configs[spotInstance])
			writeLinesToFile(unusedResources, filepath.Join(outFolder, UNUSED_PREF, spotInstance))

			// TODO: add other files to write
		}
		return nil
	})
	CheckError(err)
}

func getClusterResourcesList(clusterResourcesMap map[string]string) []string {
	list := make([]string, len(clusterResourcesMap))
	for k := range clusterResourcesMap {
		list = append(list, k)
	}
	return list
}

func getSpotInstanceName(fileName string) string {
	startIndex := strings.Index(fileName, "SPOT_") + 5
	endIndex := strings.LastIndex(fileName, ".json")
	return fileName[startIndex:endIndex]
}

func writeLinesToFile(slice []string, fullFilePath string) {
	file, err := os.Create(fullFilePath)
	defer CloseFile(file)
	CheckError(err)

	for _, val := range slice {
		_, err := fmt.Fprintln(file, val)
		CheckError(err)
	}
	log.Printf("Done writing file: %s", fullFilePath)
}

func getUnusedResources(resources []string, versions map[string]int, config GlobalConfig) []string {
	var globalResources []GlobalResource
	var siteResources []SiteResource
	var localeResources []LocaleResource

	var resultList []string

	for _, resource := range resources {
		if localePattern.MatchString(resource) {
			subgroups := GetRegexSubgroups(localePattern, resource)
			localeResources = append(localeResources, *NewLocaleResourceFrom(subgroups))
		} else if sitePattern.MatchString(resource) {
			subgroups := GetRegexSubgroups(sitePattern, resource)
			siteResources = append(siteResources, *NewSiteResourceFrom(subgroups))
		} else if globalPattern.MatchString(resource) {
			subgroups := GetRegexSubgroups(globalPattern, resource)
			globalResources = append(globalResources, *NewGlobalResourceFrom(subgroups))
		}
	}

	for _, e := range localeResources {
		resultList = append(resultList, fmt.Sprintf("LOCALE: %#v", e))
	}

	for _, e := range siteResources {
		resultList = append(resultList, fmt.Sprintf("SITE: %#v", e))
	}

	for _, e := range globalResources {
		resultList = append(resultList, fmt.Sprintf("GLOBAL: %#v", e))
	}

	return resultList
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

	exp, err := regexp.Compile("^\\s*(?P<name>[a-zA-Z_-]+)\\s+->\\s+[vV](?P<version>[0-9]{1,2})\\s*$")
	CheckError(err)
	scanner := bufio.NewScanner(openedFile)
	for scanner.Scan() {
		regexSubGroups := GetRegexSubgroups(exp, scanner.Text())
		version, err := strconv.ParseInt(regexSubGroups["version"], 0, 32)
		CheckError(err)
		versionsMap[regexSubGroups["name"]] = int(version)
	}

	return versionsMap
}
