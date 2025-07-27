package adoptium

import (
	"jenvy/internal/utils"
	"sort"
)

type AdoptiumEntry struct {
	Version string
	OS      string
	Arch    string
	LTS     string
	Link    string
	Major   int
	Minor   int
	Patch   int
}

// parseVersion è stata sostituita da utils.ParseVersionNumber per coerenza
// Usa utils.ParseVersionNumber(v) invece di ParseVersion(v)
func ParseVersion(v string) (int, int, int) {
	return utils.ParseVersionNumber(v)
}

// getLatestAdoptium ordina le release per versione e restituisce solo la più recente
func GetLatestAdoptium(list []AdoptiumResponse, majorOnly bool) []AdoptiumEntry {
	var entries []AdoptiumEntry
	for _, j := range list {
		v := j.VersionData.OpenJDKVersion
		major, minor, patch := ParseVersion(v)

		if majorOnly && minor != 0 {
			continue
		}

		for _, b := range j.Binaries {
			isLTS := utils.IsLTSVersion(v)
			entries = append(entries, AdoptiumEntry{
				Version: v,
				OS:      b.OS,
				Arch:    b.Arch,
				LTS:     utils.IfBool(isLTS),
				Link:    b.Package.Link,
				Major:   major,
				Minor:   minor,
				Patch:   patch,
			})
		}
	}

	// ordina per Major, Minor, Patch decrescente
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Major != entries[j].Major {
			return entries[i].Major > entries[j].Major
		}
		if entries[i].Minor != entries[j].Minor {
			return entries[i].Minor > entries[j].Minor
		}
		return entries[i].Patch > entries[j].Patch
	})

	// restituisci la prima riga se disponibile
	if len(entries) > 0 {
		return []AdoptiumEntry{entries[0]}
	}
	return nil
}
