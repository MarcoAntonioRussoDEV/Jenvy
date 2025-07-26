package azul

import (
	"jenvy/internal/utils"
	"sort"
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

func GetRecommendedJDKs(list []AzulPackage) []RecommendedEntry {
    group := make(map[int][]RecommendedEntry)
    for _, j := range list {
        if len(j.JavaVersion) < 2 || !endsWithZip(j.DownloadURL) {
            continue
        }
        major := j.JavaVersion[0]
        minor := j.JavaVersion[1]
        patch := 0
        if len(j.JavaVersion) > 2 {
            patch = j.JavaVersion[2]
        }

        isLTS := major == 11 || major == 17 || major == 21 || major == 24
        os, arch := utils.InferPlatform(j.Name)

        entry := RecommendedEntry{
            Version:     utils.FormatVersion(j.JavaVersion),
            DownloadURL: j.DownloadURL,
            OS:          os,
            Arch:        arch,
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


