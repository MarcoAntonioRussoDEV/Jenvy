package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// Stampa intestazione e tabella con evidenziazione delle righe LTS
func PrintTable(data [][]string, headers []string) {
	// Intestazione
	color.New(color.FgHiYellow, color.Bold).Printf(
		"%-18s %-10s %-8s %-6s %s\n",
		headers[0], headers[1], headers[2], headers[3], headers[4])

	color.New(color.FgHiWhite).Printf(
		"%-18s %-10s %-8s %-6s %s\n",
		"──────────────────", "──────────", "────────", "──────", "─────────────────────────────────────────────────────────────")

	// Righe dati
	for _, row := range data {
		if row[3] == "✅" {
			// Riga evidenziata (versione LTS)
			color.New(color.FgHiGreen).Printf(
				"%-18s %-10s %-8s %-6s %s\n", row[0], row[1], row[2], row[3], row[4])
		} else {
			// Riga normale
			fmt.Printf("%-18s %-10s %-8s %-6s %s\n", row[0], row[1], row[2], row[3], row[4])
		}
	}
}

// Stampa una singola riga LTS colorata (usata in versioni precedenti)
func PrintColoredRow(row []string) {
	color.New(color.FgHiGreen).Printf(
		"%-18s %-10s %-8s %-6s %s\n", row[0], row[1], row[2], row[3], row[4])
}

// Converte boolean in icona ✓ o trattino
func IfBool(b bool) string {
	if b {
		return "✅"
	}
	return "–"
}

// Converte slice di interi in versione stringa, es. [21 0 0] → "21.0.0"
func FormatVersion(v []int) string {
	var parts []string
	for _, n := range v {
		parts = append(parts, fmt.Sprintf("%d", n))
	}
	return strings.Join(parts, ".")
}

// Deduce OS e Arch dal nome del file Azul
func InferPlatform(name string) (string, string) {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "win_x64"):
		return "windows", "x64"
	case strings.Contains(name, "linux_x64"):
		return "linux", "x64"
	case strings.Contains(name, "macos_x64"):
		return "macos", "x64"
	default:
		return "?", "?"
	}
}

// ParseGenericVersion estrae major, minor e patch da una stringa tipo "17.0.7"
func ParseGenericVersion(v string) (int, int, int) {
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

// SortRecommended ordina una slice con priorità: LTS ✅ > patch più alta > minor più alto
func SortRecommended[T VersionComparable](list []T) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].LtsValue() != list[j].LtsValue() {
			return list[i].LtsValue()
		}
		if list[i].PatchValue() != list[j].PatchValue() {
			return list[i].PatchValue() > list[j].PatchValue()
		}
		return list[i].MinorValue() > list[j].MinorValue()
	})
}

type Entry interface {
	MajorValue() int
	LtsValue() bool
	PatchValue() int
	MinorValue() int
}

func LatestForEachMajor(list []Entry, majorOnly bool) []Entry {
	group := make(map[int][]Entry)
	for _, e := range list {
		group[e.MajorValue()] = append(group[e.MajorValue()], e)
	}

	var result []Entry
	for _, entries := range group {
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].LtsValue() != entries[j].LtsValue() {
				return entries[i].LtsValue() // Priorità LTS
			}
			if entries[i].PatchValue() != entries[j].PatchValue() {
				return entries[i].PatchValue() > entries[j].PatchValue()
			}
			return entries[i].MinorValue() > entries[j].MinorValue()
		})
		result = append(result, entries[0])
	}

	// Ordina i risultati
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].MajorValue() < result[j].MajorValue()
	})

	return result
}

type VersionComparable interface {
	LtsValue() bool
	PatchValue() int
	MinorValue() int
}
