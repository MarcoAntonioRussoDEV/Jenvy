package adoptium

import (
	"jvm/internal/utils"
	"sort"
)

type RecommendedEntry struct {
	Version string
	Link    string
	OS      string
	Arch    string
	LTS     string
	Major   int
	Minor   int
	Patch   int
}

// GetRecommendedJDKs restituisce una sola release per ciascun major JDK
func GetRecommendedJDKs(list []AdoptiumResponse) []RecommendedEntry {
	group := make(map[int][]RecommendedEntry)

	for _, j := range list {
		version := j.VersionData.OpenJDKVersion
		major, minor, patch := ParseVersion(version)
		isLTS := utils.IsLTSVersion(version)

		for _, b := range j.Binaries {
			entry := RecommendedEntry{
				Version: version,
				Link:    b.Package.Link,
				OS:      b.OS,
				Arch:    b.Arch,
				LTS:     utils.IfBool(isLTS),
				Major:   major,
				Minor:   minor,
				Patch:   patch,
			}
			group[major] = append(group[major], entry)
		}
	}

	var result []RecommendedEntry
	for _, entries := range group {
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].LTS != entries[j].LTS {
				return entries[i].LTS == utils.IfBool(true)
			}
			if entries[i].Patch != entries[j].Patch {
				return entries[i].Patch > entries[j].Patch
			}
			return entries[i].Minor > entries[j].Minor
		})
		result = append(result, entries[0])
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Major < result[j].Major
	})
	return result
}
