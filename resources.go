package spothelper

import "strconv"

type GlobalConfig struct {
	Host       string   `json:"host"`
	Root       string   `json:"root"`
	BackupRoot string   `json:"backupRoot"`
	Sites      []string `json:"sites"`
}

type Resource struct {
	files []string
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

func NewLocaleResourceFrom(subgroups map[string]string) *LocaleResource {
	return &LocaleResource{
		NewSiteResourceFrom(subgroups),
		subgroups["locale"],
	}
}

func NewSiteResourceFrom(subgroups map[string]string) *SiteResource {
	return &SiteResource{
		NewGlobalResourceFrom(subgroups),
		subgroups["site"],
	}
}

func NewGlobalResourceFrom(subgroups map[string]string) *GlobalResource {
	version, err := strconv.ParseInt(subgroups["version"], 0, 32)
	CheckError(err)
	return &GlobalResource{
		NewResourceFrom(subgroups),
		subgroups["name"],
		int(version),
	}
}

func NewResourceFrom(subgroups map[string]string) *Resource {
	return &Resource{}
}
