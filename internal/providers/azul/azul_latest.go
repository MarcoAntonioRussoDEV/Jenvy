package azul

import (
	"jenvy/internal/utils"
	"sort"
)

type AzulEntry struct {
    Version     string
    DownloadURL string
    OS          string
    Arch        string
    LTS         string
    Major       int
    Minor       int
    Patch       int
}

// GetLatestAzul restituisce la release Azul più recente (solo .zip standard)
func GetLatestAzul(list []AzulPackage, majorOnly bool) []AzulEntry {
    var entries []AzulEntry

    for _, j := range list {
        if !j.Latest || !endsWithZip(j.DownloadURL) {
            continue
        }
        if len(j.JavaVersion) < 2 {
            continue
        }
        major := j.JavaVersion[0]
        minor := j.JavaVersion[1]
        patch := 0
        if len(j.JavaVersion) > 2 {
            patch = j.JavaVersion[2]
        }
        if majorOnly && minor != 0 {
            continue
        }

        os, arch := utils.InferPlatform(j.Name)
        isLTS := major == 11 || major == 17 || major == 21 || major == 24

        entry := AzulEntry{
            Version:     utils.FormatVersion(j.JavaVersion),
            DownloadURL: j.DownloadURL,
            OS:          os,
            Arch:        arch,
            LTS:         utils.IfBool(isLTS),
            Major:       major,
            Minor:       minor,
            Patch:       patch,
        }
        entries = append(entries, entry)
    }

    // Ordina discendente
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
        return []AzulEntry{entries[0]}
    }
    return nil
}

func endsWithZip(url string) bool {
    return len(url) > 4 && url[len(url)-4:] == ".zip"
}
