package adoptium

import (
	"jvm/utils"
	"sort"
	"strconv"
	"strings"
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

// parseVersion tenta di estrarre major, minor, patch da versioni come:
// "21.0.2+13", "17.0.0+35", "1.8.0_452-b09"
func ParseVersion(v string) (int, int, int) {
    v = strings.TrimSpace(v)
    // Legacy Java 8 es: "1.8.0_452-b09"
    if strings.HasPrefix(v, "1.8.0_") {
        parts := strings.Split(v, "_")
        num := strings.Trim(parts[1], "b-")
        if u, err := strconv.Atoi(num); err == nil {
            return 8, 0, u
        }
        return 8, 0, 0
    }

    // Versioni moderne es: "21.0.2+13"
    parts := strings.Split(v, "+")
    left := parts[0]
    subs := strings.Split(left, ".")
    major, minor, patch := 0, 0, 0

    if len(subs) > 0 {
        major, _ = strconv.Atoi(subs[0])
    }
    if len(subs) > 1 {
        minor, _ = strconv.Atoi(subs[1])
    }
    if len(subs) > 2 {
        patch, _ = strconv.Atoi(subs[2])
    }
    return major, minor, patch
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
            isLTS := strings.Contains(strings.ToLower(v), "lts")
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
