package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	. "github.com/lcserny/goutils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	mainProcessThreadCount = 10

	UNUSED_PREF        = "UNUSED"
	MISC_PREF          = "MISC_UNUSED"
	DEL_COMMANDS_PREF  = "DELETE_CMD"
	BACK_COMMANDS_PREF = "BACKUP_CMD"
)

var (
	globalPattern  = regexp.MustCompile("^(?P<name>.*)/[vV](?P<version>[0-9]{1,2})/(?P<file>[^/]+)$")
	sitePattern    = regexp.MustCompile("^(?P<site>.*)/(?P<name>.*)/[vV](?P<version>[0-9]{1,2})/(?P<file>[^/]+)$")
	localePattern  = regexp.MustCompile("^(?P<site>.*)/(?P<locale>.*)/(?P<name>.*)/[vV](?P<version>[0-9]{1,2})/(?P<file>[^/]+)$")
	nameExcPattern = regexp.MustCompile("orderedcount-[a-zA-Z]+")
)

func ProcessUnused(spotVersionsFile, globalConfigFile, inFolder, outFolder string) {
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
			LogFatal(err)
			err = os.MkdirAll(filepath.Join(newOutFolder, MISC_PREF), os.ModePerm)
			LogFatal(err)
			err = os.MkdirAll(filepath.Join(newOutFolder, DEL_COMMANDS_PREF), os.ModePerm)
			LogFatal(err)
			err = os.MkdirAll(filepath.Join(newOutFolder, BACK_COMMANDS_PREF), os.ModePerm)
			LogFatal(err)
			log.Printf("Created output folders in: %s", newOutFolder)
			return newOutFolder
		}

		count++
		newOutFolder = generateNewOutFolderPath(outFolder, today, count)
	}
}

func generateNewOutFolderPath(outFolder, date string, count int) string {
	return filepath.Join(outFolder, fmt.Sprintf("%s_%d", date, count))
}

func writeFiles(versionsMap map[string]int, configs map[string]GlobalConfig, inFolder, outFolder string) {
	err := filepath.Walk(inFolder, func(path string, info os.FileInfo, err error) error {
		LogFatal(err)
		if !info.IsDir() {
			allBytes, err := ioutil.ReadFile(path)
			LogFatal(err)
			var jsonClusterResourcesMap map[string]string
			err = json.Unmarshal(allBytes, &jsonClusterResourcesMap)
			LogFatal(err)
			clusterResourcesList := getClusterResourcesList(jsonClusterResourcesMap)

			spotInstance := getSpotInstanceName(info.Name())

			unusedResources, unmatchedResources, deleteCommands, backupCommands := process(clusterResourcesList, versionsMap, configs[spotInstance])
			writeLinesToFile(unusedResources, filepath.Join(outFolder, UNUSED_PREF, spotInstance))
			writeLinesToFile(unmatchedResources, filepath.Join(outFolder, MISC_PREF, spotInstance))
			writeLinesToFile(deleteCommands, filepath.Join(outFolder, DEL_COMMANDS_PREF, spotInstance))
			writeLinesToFile(backupCommands, filepath.Join(outFolder, BACK_COMMANDS_PREF, spotInstance))
		}
		return nil
	})
	LogFatal(err)
}

func getClusterResourcesList(clusterResourcesMap map[string]string) []string {
	var list []string
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
	LogFatal(err)
	defer CloseFile(file)

	for _, val := range slice {
		_, err := fmt.Fprintln(file, val)
		LogFatal(err)
	}
	log.Printf("Done writing file: %s", fullFilePath)
}

func process(resources []string, versions map[string]int, config GlobalConfig) ([]string, []string, []string, []string) {
	resCount := len(resources)
	threadsCount := mainProcessThreadCount
	if resCount < threadsCount {
		threadsCount = 1
	}
	resourcesChunkSize := resCount / threadsCount

	unusedResourcesHolder := make(map[int][]string, threadsCount)
	unmatchedResourcesHolder := make(map[int][]string, threadsCount)

	wg := &sync.WaitGroup{}
	wg.Add(threadsCount)

	for i := 1; i <= threadsCount; i++ {
		startIdx := 0
		if i > 1 {
			startIdx = (i - 1) * resourcesChunkSize
		}

		endIdx := i * resourcesChunkSize
		if i == threadsCount {
			endIdx = resCount
		}

		go func(iter, start, end int) {
			unusedResources, unmatchedResources := processChunk(resources[start:end], versions, config)
			unusedResourcesHolder[iter] = unusedResources
			unmatchedResourcesHolder[iter] = unmatchedResources
			wg.Done()
		}(i, startIdx, endIdx)
	}

	wg.Wait()

	var unusedResources []string
	for _, slice := range unusedResourcesHolder {
		unusedResources = append(unusedResources, slice...)
	}

	var unmatchedResources []string
	for _, slice := range unmatchedResourcesHolder {
		unmatchedResources = append(unmatchedResources, slice...)
	}

	sort.Strings(unusedResources)
	sort.Strings(unmatchedResources)
	deleteCommands, backupCommands := produceCommands(append(unusedResources, unmatchedResources...), &config)

	return unusedResources, unmatchedResources, deleteCommands, backupCommands
}

func processChunk(resourcesChunk []string, versions map[string]int, config GlobalConfig) (unusedResources, unmatchedResources []string) {
	for _, resource := range resourcesChunk {
		if localePattern.MatchString(resource) {
			localeResource := NewLocaleResourceFrom(resource, GetRegexSubgroups(localePattern, resource))
			if unused, found := findUnusedResource(versions, config, *localeResource.SiteResource); found {
				unusedResources = append(unusedResources, unused)
			}
		} else if sitePattern.MatchString(resource) {
			siteResource := NewSiteResourceFrom(resource, GetRegexSubgroups(sitePattern, resource))
			if unused, found := findUnusedResource(versions, config, *siteResource); found {
				unusedResources = append(unusedResources, unused)
			}
		} else if globalPattern.MatchString(resource) {
			globalResource := NewGlobalResourceFrom(resource, GetRegexSubgroups(globalPattern, resource))
			if unused, found := findSpecificUnusedResource(versions, globalResource); found {
				unusedResources = append(unusedResources, unused)
			}
		} else {
			unmatchedResources = append(unmatchedResources, resource)
		}
	}
	return unusedResources, unmatchedResources
}

func findUnusedResource(versions map[string]int, config GlobalConfig, siteResource SiteResource) (string, bool) {
	if !StringsContain(config.Sites, siteResource.site) {
		return siteResource.file, true
	}
	return findSpecificUnusedResource(versions, siteResource.GlobalResource)
}

func findSpecificUnusedResource(versions map[string]int, resource *GlobalResource) (string, bool) {
	noConsumerMatched := true
	for rName, rVer := range versions {
		if rName == resource.name {
			noConsumerMatched = false
			if resource.version < rVer {
				return resource.file, true
			}
		}
	}

	if noConsumerMatched && noNameExceptionApplies(resource.name) {
		return resource.file, true
	}

	return "", false
}

func produceCommands(concatenatedList []string, config *GlobalConfig) ([]string, []string) {
	var deleteCommands []string
	var backupCommands []string
	for _, ele := range concatenatedList {
		deleteCommands = addDeleteCommand(deleteCommands, ele, config)
		backupCommands = addBackupCommand(backupCommands, ele, config)
	}
	return deleteCommands, backupCommands
}

func addBackupCommand(commands []string, element string, config *GlobalConfig) []string {
	parent := filepath.Dir(element)

	mkdir := fmt.Sprintf("mkdir -p %s/%s", config.BackupRoot, parent)
	if !StringsContain(commands, mkdir) {
		commands = append(commands, mkdir)
	}

	cp := fmt.Sprintf("cp -f %s/%s* %s/%s", config.Root, element, config.BackupRoot, parent)
	if !StringsContain(commands, cp) {
		commands = append(commands, cp)
	}

	return commands
}

func addDeleteCommand(commands []string, element string, config *GlobalConfig) []string {
	curl := fmt.Sprintf("curl -X \"DELETE\" %s/spot/resource/%s", config.Host, element)
	if !StringsContain(commands, curl) {
		commands = append(commands, curl)
	}
	return commands
}

func noNameExceptionApplies(resourceName string) bool {
	if nameExcPattern.MatchString(resourceName) {
		return false
	}
	// TODO: add more exceptions?
	return true
}

func getGlobalConfigMap(globalConfigFile string) map[string]GlobalConfig {
	allBytes, err := ioutil.ReadFile(globalConfigFile)
	LogFatal(err)

	resultMap := make(map[string]GlobalConfig)

	var jsonData []map[string]GlobalConfig
	err = json.Unmarshal(allBytes, &jsonData)
	LogFatal(err)
	for _, element := range jsonData {
		for s, gc := range element {
			resultMap[s] = gc
		}
	}

	return resultMap
}

func getVersions(spotVersionsFile string) map[string]int {
	openedFile, err := os.Open(spotVersionsFile)
	LogFatal(err)
	defer CloseFile(openedFile)

	versionsMap := make(map[string]int)

	exp := regexp.MustCompile("^\\s*(?P<name>[a-zA-Z_-]+)\\s+->\\s+[vV](?P<version>[0-9]{1,2})\\s*$")
	scanner := bufio.NewScanner(openedFile)
	for scanner.Scan() {
		regexSubGroups := GetRegexSubgroups(exp, scanner.Text())
		version, err := strconv.ParseInt(regexSubGroups["version"], 0, 32)
		LogFatal(err)
		versionsMap[regexSubGroups["name"]] = int(version)
	}

	return versionsMap
}
