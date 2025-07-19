package utils

import (
	"fmt"
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
