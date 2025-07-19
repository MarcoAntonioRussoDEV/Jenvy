package liberica

import (
	"jvm/utils"
	"sort"
	"strconv"
	"strings"
)

type LibericaEntry struct {
    Version string
    DownloadURL string
    OS      string
    Arch    string
    LTS     string
    Major   int
    Minor   int
    Patch   int
}

// parseLibericaVersion estrae major/minor/patch da versioni tipo "17.0.10+9", "21+35", "8u352", "14+36"
func ParseLibericaVersion(v string) (int, int, int) {
    // Rimuovi eventuali "+build" o "u" con patch
    v = strings.Split(v, "+")[0]
    v = strings.TrimSpace(v)

    // Java 8 es: "8u352"
    if strings.HasPrefix(v, "8u") {
        patchStr := strings.TrimPrefix(v, "8u")
        patch, _ := strconv.Atoi(patchStr)
        return 8, 0, patch
    }

    // Es. "21+35", "17.0.10"
    if strings.Contains(v, ".") {
        parts := strings.Split(v, ".")
        major, minor, patch := 0, 0, 0
        if len(parts) > 0 {
            major, _ = strconv.Atoi(parts[0])
        }
        if len(parts) > 1 {
            minor, _ = strconv.Atoi(parts[1])
        }
        if len(parts) > 2 {
            patch, _ = strconv.Atoi(parts[2])
        }
        return major, minor, patch
    }

    // Es. "21"
    major, err := strconv.Atoi(v)
    if err != nil {
        return 0, 0, 0
    }
    return major, 0, 0
}

// GetLatestLiberica estrae solo la release più recente, eventualmente filtrando solo le major
func GetLatestLiberica(list []LibericaRelease, majorOnly bool) []LibericaEntry {
    var entries []LibericaEntry

    for _, j := range list {
        major, minor, patch := ParseLibericaVersion(j.Version)
        if majorOnly && minor != 0 {
            continue
        }
        isLTS := strings.HasPrefix(j.Version, "17.") || strings.HasPrefix(j.Version, "21.") || strings.Contains(strings.ToLower(j.Version), "lts")

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
