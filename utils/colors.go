package utils

import (
	"fmt"
	"runtime"
)

// ANSI color codes
const (
	// Reset
	Reset = "\033[0m"

	// Regular colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Styles
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"
)

// Color helper functions
func ColorText(text, color string) string {
	if !supportsColor() {
		return text
	}
	return color + text + Reset
}

func ErrorText(text string) string {
	return ColorText("[ERROR] "+text, BrightRed)
}

func SuccessText(text string) string {
	return ColorText("[SUCCESS] "+text, BrightGreen)
}

func InfoText(text string) string {
	return ColorText("[INFO] "+text, BrightBlue)
}

func WarningText(text string) string {
	return ColorText("[WARN] "+text, BrightYellow)
}

func FetchText(text string) string {
	return ColorText("[FETCH] "+text, BrightCyan)
}

func SearchText(text string) string {
	return ColorText("[SEARCH] "+text, BrightMagenta)
}

func DownloadText(text string) string {
	return ColorText("[DOWNLOAD] "+text, Cyan)
}

func ReadyText(text string) string {
	return ColorText("[READY] "+text, BrightGreen)
}

func UsageText(text string) string {
	return ColorText("[USAGE] "+text, BrightYellow)
}

func ExamplesText(text string) string {
	return ColorText("[EXAMPLES] "+text, BrightMagenta+Bold)
}

func SectionText(text string) string {
	return ColorText(text, Bold+BrightWhite)
}

// Check if terminal supports colors
func supportsColor() bool {
	// On Windows, check if we're in a modern terminal
	if runtime.GOOS == "windows" {
		// Modern Windows terminals support ANSI colors
		return true
	}
	return true
}

// Print colored text functions
func PrintError(text string) {
	fmt.Println(ErrorText(text))
}

func PrintSuccess(text string) {
	fmt.Println(SuccessText(text))
}

func PrintInfo(text string) {
	fmt.Println(InfoText(text))
}

func PrintWarning(text string) {
	fmt.Println(WarningText(text))
}

func PrintFetch(text string) {
	fmt.Println(FetchText(text))
}

func PrintSearch(text string) {
	fmt.Println(SearchText(text))
}

func PrintDownload(text string) {
	fmt.Println(DownloadText(text))
}

func PrintReady(text string) {
	fmt.Println(ReadyText(text))
}

func PrintUsage(text string) {
	fmt.Println(UsageText(text))
}

func PrintSection(text string) {
	fmt.Println(SectionText(text))
}
