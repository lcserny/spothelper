package spothelper

import (
	. "github.com/lcserny/goutils"
	"strconv"
)

type GlobalConfig struct {
	Host       string   `json:"host"`
	Root       string   `json:"root"`
	BackupRoot string   `json:"backupRoot"`
	Sites      []string `json:"sites"`
}

type Resource struct {
	file string
}

type GlobalResource struct {
	*Resource
	name    string
	version int
}

type SiteResource struct {
	*GlobalResource
	site string
}

type LocaleResource struct {
	*SiteResource
	locale string
}

func NewLocaleResourceFrom(resource string, subgroups map[string]string) *LocaleResource {
	return &LocaleResource{
		NewSiteResourceFrom(resource, subgroups),
		subgroups["locale"],
	}
}

func NewSiteResourceFrom(resource string, subgroups map[string]string) *SiteResource {
	return &SiteResource{
		NewGlobalResourceFrom(resource, subgroups),
		subgroups["site"],
	}
}

func NewGlobalResourceFrom(resource string, subgroups map[string]string) *GlobalResource {
	version, err := strconv.ParseInt(subgroups["version"], 0, 32)
	LogFatal(err)
	return &GlobalResource{
		NewResourceFrom(resource),
		subgroups["name"],
		int(version),
	}
}

func NewResourceFrom(resource string) *Resource {
	return &Resource{
		resource,
	}
}
