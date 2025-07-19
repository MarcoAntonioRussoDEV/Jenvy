package liberica

import (
	"jvm/utils"
	"sort"
	"strings"
)

type RecommendedEntry struct {
    Version     string
    DownloadURL string
    OS          string
    Arch        string
    LTS         string
    Major       int
    Minor       int
    Patch       int
}


// GetRecommendedJDKs restituisce una sola release per ciascun major
func GetRecommendedJDKs(list []LibericaRelease) []RecommendedEntry {
    group := make(map[int][]RecommendedEntry)
    for _, j := range list {
        major, minor, patch := ParseLibericaVersion(j.Version)
        isLTS := strings.HasPrefix(j.Version, "17.") ||
            strings.HasPrefix(j.Version, "21.") ||
            strings.Contains(strings.ToLower(j.Version), "lts")

        entry := RecommendedEntry{
            Version:     j.Version,
            DownloadURL: j.DownloadURL,
            OS:          j.OS,
            Arch:        j.Arch,
            LTS:         utils.IfBool(isLTS),
            Major:       major,
            Minor:       minor,
            Patch:       patch,
        }
        group[major] = append(group[major], entry)
    }

    var result []RecommendedEntry
    for _, entries := range group {
        sort.SliceStable(entries, func(i, j int) bool {
            // Priorità: LTS > Patch più alta
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
