package liberica

import (
	"jvm/internal/utils"
	"sort"
)

type LibericaEntry struct {
	Version     string
	DownloadURL string
	OS          string
	Arch        string
	LTS         string
	Major       int
	Minor       int
	Patch       int
}

// parseLibericaVersion è stata sostituita da utils.ParseVersionNumber per coerenza
// Usa utils.ParseVersionNumber(v) invece di ParseLibericaVersion(v)
func ParseLibericaVersion(v string) (int, int, int) {
	return utils.ParseVersionNumber(v)
}

// GetLatestLiberica estrae solo la release più recente, eventualmente filtrando solo le major
func GetLatestLiberica(list []LibericaRelease, majorOnly bool) []LibericaEntry {
	var entries []LibericaEntry

	for _, j := range list {
		major, minor, patch := ParseLibericaVersion(j.Version)
		if majorOnly && minor != 0 {
			continue
		}
		isLTS := utils.IsLTSVersion(j.Version)

		entry := LibericaEntry{
			Version:     j.Version,
			OS:          j.OS,
			Arch:        j.Arch,
			DownloadURL: j.DownloadURL,
			LTS:         utils.IfBool(isLTS),
			Major:       major,
			Minor:       minor,
			Patch:       patch,
		}
		entries = append(entries, entry)
	}

	// Ordina in ordine decrescente
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Major != entries[j].Major {
			return entries[i].Major > entries[j].Major
		}
		if entries[i].Minor != entries[j].Minor {
			return entries[i].Minor > entries[j].Minor
		}
		return entries[i].Patch > entries[j].Patch
	})

	if len(entries) > 0 {
		return []LibericaEntry{entries[0]}
	}
	return nil
}
