package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"jvm/providers/adoptium"
	"jvm/providers/azul"
	"jvm/providers/liberica"
	"jvm/providers/private"
	"jvm/ui"
)

// RuntimeInfo holds OS and architecture information
type RuntimeInfo struct {
	OS   string
	Arch string
}

// getRuntimeInfo returns the current OS and architecture in JDK format
func getRuntimeInfo() RuntimeInfo {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Convert Go runtime names to JDK format
	switch osName {
	case "windows":
		osName = "windows"
	case "darwin":
		osName = "mac"
	case "linux":
		osName = "linux"
	}

	switch archName {
	case "amd64":
		archName = "x64"
	case "386":
		archName = "x32"
	case "arm64":
		archName = "aarch64"
	}

	return RuntimeInfo{OS: osName, Arch: archName}
}

// parseVersionNumber parses a version string like "17", "17.0", "17.0.5" into components
func parseVersionNumber(version string) (major, minor, patch int) {
	parts := strings.Split(version, ".")
	major = -1
	minor = -1
	patch = -1

	if len(parts) >= 1 {
		if m, err := strconv.Atoi(parts[0]); err == nil {
			major = m
		}
	}
	if len(parts) >= 2 {
		if m, err := strconv.Atoi(parts[1]); err == nil {
			minor = m
		}
	}
	if len(parts) >= 3 {
		if p, err := strconv.Atoi(parts[2]); err == nil {
			patch = p
		}
	}

	return major, minor, patch
}

// shouldPreferVersion returns true if version1 should be preferred over version2
// Prefers newer versions and LTS versions
func shouldPreferVersion(version1, version2 string) bool {
	v1Major, v1Minor, v1Patch := adoptium.ParseVersion(version1)
	v2Major, v2Minor, v2Patch := adoptium.ParseVersion(version2)

	// Prefer higher major version
	if v1Major != v2Major {
		return v1Major > v2Major
	}

	// Prefer higher minor version
	if v1Minor != v2Minor {
		return v1Minor > v2Minor
	}

	// Prefer higher patch version
	return v1Patch > v2Patch
}

// DownloadJDK downloads a specific JDK version
func DownloadJDK(defaultProvider string) {
	ui.ShowBanner()

	// Parse command line arguments
	args := os.Args[2:] // Skip "download"
	if len(args) == 0 {
		fmt.Println("❌ No JDK version specified")
		fmt.Println("💡 Usage: jvm download <version> [options]")
		fmt.Println("💡 Examples:")
		fmt.Println("  jvm download 17          # Download JDK 17")
		fmt.Println("  jvm download 21.0.5      # Download specific version")
		fmt.Println("  jvm download 17 --provider=azul")
		return
	}

	version := args[0]
	provider := defaultProvider

	// Get default download directory: ~/.jvm/versions
	outputDir, dirErr := getDefaultDownloadDir()
	if dirErr != nil {
		fmt.Printf("❌ Failed to determine download directory: %v\n", dirErr)
		outputDir = "./downloads" // fallback
	}

	// Parse optional flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--provider=") {
			provider = strings.TrimPrefix(arg, "--provider=")
		} else if strings.HasPrefix(arg, "--output=") {
			outputDir = strings.TrimPrefix(arg, "--output=")
		}
	}

	fmt.Printf("🔍 Searching for JDK version %s from provider: %s\n", version, provider)
	fmt.Printf("📁 Download directory: %s\n\n", outputDir)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create output directory: %v\n", err)
		return
	}

	// Get JDK releases from the specified provider and find matching version
	var downloadURL string
	var filename string
	var foundVersion string

	switch provider {
	case "adoptium":
		releases, err := adoptium.GetAllJDKs()
		if err != nil {
			fmt.Printf("❌ Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findAdoptiumDownload(releases, version)

	case "azul":
		releases, err := azul.GetAzulJDKs()
		if err != nil {
			fmt.Printf("❌ Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findAzulDownload(releases, version)

	case "liberica":
		releases, err := liberica.GetLibericaJDKs()
		if err != nil {
			fmt.Printf("❌ Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findLibericaDownload(releases, version)

	case "private":
		releases, err := private.GetPrivateJDKs()
		if err != nil {
			fmt.Printf("❌ Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findPrivateDownload(releases, version)

	default:
		fmt.Printf("❌ Unknown provider: %s\n", provider)
		fmt.Println("💡 Available providers: adoptium, azul, liberica, private")
		return
	}

	if downloadURL == "" {
		fmt.Printf("❌ JDK version %s not found in %s provider\n", version, provider)
		fmt.Println("💡 Try running 'jvm remote-list' to see available versions")
		return
	}

	if filename == "" {
		filename = fmt.Sprintf("openjdk-%s.tar.gz", version)
	}

	outputPath := filepath.Join(outputDir, filename)

	fmt.Printf("📦 Found JDK %s\n", foundVersion)
	fmt.Printf("🔗 Download URL: %s\n", downloadURL)
	fmt.Printf("💾 Saving to: %s\n", outputPath)

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("⚠️  File already exists: %s\n", filename)
	}

	// Ask for confirmation
	fmt.Print("\n🤔 Do you want to proceed with the download? (y/N): ")
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("❌ Download cancelled by user")
		return
	}

	fmt.Println()

	// Download the file
	if err := downloadFile(downloadURL, outputPath); err != nil {
		fmt.Printf("❌ Download failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Download completed successfully!\n")
	fmt.Printf("📁 File saved to: %s\n", outputPath)

	// Show file info
	if fileInfo, err := os.Stat(outputPath); err == nil {
		fmt.Printf("📏 File size: %.2f MB\n", float64(fileInfo.Size())/1024/1024)
		fmt.Printf("🕒 Download time: %s\n", time.Now().Format("15:04:05"))
	}
}

// downloadFile downloads a file from URL to the specified path with progress indication
func downloadFile(url, filepath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Minute * 30, // 30 minutes timeout for large files
	}

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "JVM-Manager/1.0")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create the output file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer out.Close()

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	var downloaded int64

	// Create a buffer for copying
	buffer := make([]byte, 32*1024) // 32KB buffer

	fmt.Println("⬇️ Downloading...")
	startTime := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("writing to file: %w", writeErr)
			}
			downloaded += int64(n)

			// Show progress if we know the content length
			if contentLength > 0 {
				progress := float64(downloaded) / float64(contentLength) * 100
				elapsed := time.Since(startTime)
				speed := float64(downloaded) / elapsed.Seconds() / 1024 / 1024 // MB/s

				fmt.Printf("\r📊 Progress: %.1f%% (%.2f MB / %.2f MB) - Speed: %.2f MB/s",
					progress,
					float64(downloaded)/1024/1024,
					float64(contentLength)/1024/1024,
					speed,
				)
			} else {
				// Show downloaded amount without percentage
				elapsed := time.Since(startTime)
				speed := float64(downloaded) / elapsed.Seconds() / 1024 / 1024 // MB/s

				fmt.Printf("\r📊 Downloaded: %.2f MB - Speed: %.2f MB/s",
					float64(downloaded)/1024/1024,
					speed,
				)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading response: %w", err)
		}
	}

	fmt.Println() // New line after progress
	return nil
}

// getDefaultDownloadDir returns the default download directory: ~/.jvm/versions
func getDefaultDownloadDir() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}

	jvmDir := filepath.Join(currentUser.HomeDir, ".jvm", "versions")
	return jvmDir, nil
}

// findAdoptiumDownload searches for a download URL in Adoptium releases with improved version matching
func findAdoptiumDownload(releases []adoptium.AdoptiumResponse, version string) (string, string, string) {
	runtime := getRuntimeInfo()

	// Parse target version
	targetMajor, targetMinor, targetPatch := parseVersionNumber(version)

	var bestMatch adoptium.AdoptiumResponse
	var bestBinary struct {
		OS      string `json:"os"`
		Arch    string `json:"architecture"`
		Package struct {
			Link string `json:"link"`
		} `json:"package"`
	}
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		releaseVersion := release.VersionData.OpenJDKVersion
		major, minor, patch := adoptium.ParseVersion(releaseVersion)

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Find the best binary for this release (prefer current OS/arch)
		for _, binary := range release.Binaries {
			if binary.OS == runtime.OS && binary.Arch == runtime.Arch {
				if !found || shouldPreferVersion(releaseVersion, bestMatch.VersionData.OpenJDKVersion) {
					bestMatch = release
					bestBinary = binary
					found = true
					break // Found perfect match
				}
			}
		}

		// If no perfect match, try any compatible binary
		if !found {
			for _, binary := range release.Binaries {
				if strings.Contains(binary.Package.Link, ".zip") {
					bestMatch = release
					bestBinary = binary
					found = true
					break
				}
			}
		}
	}

	if !found {
		return "", "", ""
	}

	url := bestBinary.Package.Link
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	return url, filename, bestMatch.VersionData.OpenJDKVersion
}

// findAzulDownload searches for a download URL in Azul releases with improved version matching
func findAzulDownload(releases []azul.AzulPackage, version string) (string, string, string) {
	runtime := getRuntimeInfo()

	// Parse target version
	targetMajor, targetMinor, targetPatch := parseVersionNumber(version)

	var bestMatch azul.AzulPackage
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		if len(release.JavaVersion) == 0 {
			continue
		}

		major := release.JavaVersion[0]
		minor := 0
		if len(release.JavaVersion) > 1 {
			minor = release.JavaVersion[1]
		}
		patch := 0
		if len(release.JavaVersion) > 2 {
			patch = release.JavaVersion[2]
		}

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Check if compatible with our platform or is a zip file
		if strings.Contains(strings.ToLower(release.Name), runtime.OS) || strings.HasSuffix(release.DownloadURL, ".zip") {
			bestMatch = release
			found = true
			break // Take the first match for now
		}
	}

	if !found {
		return "", "", ""
	}

	url := bestMatch.DownloadURL
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	// Format version string
	var versionStr string
	if len(bestMatch.JavaVersion) > 0 {
		versionStr = fmt.Sprintf("%d", bestMatch.JavaVersion[0])
		if len(bestMatch.JavaVersion) > 1 {
			versionStr += fmt.Sprintf(".%d", bestMatch.JavaVersion[1])
		}
		if len(bestMatch.JavaVersion) > 2 {
			versionStr += fmt.Sprintf(".%d", bestMatch.JavaVersion[2])
		}
	}

	return url, filename, versionStr
}

// findLibericaDownload searches for a download URL in Liberica releases with improved version matching
func findLibericaDownload(releases []liberica.LibericaRelease, version string) (string, string, string) {
	// Parse target version
	targetMajor, targetMinor, targetPatch := parseVersionNumber(version)

	var bestMatch liberica.LibericaRelease
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		major, minor, patch := liberica.ParseLibericaVersion(release.Version)

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Take the first match
		bestMatch = release
		found = true
		break
	}

	if !found {
		return "", "", ""
	}

	url := bestMatch.DownloadURL
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	return url, filename, bestMatch.Version
}

// findPrivateDownload searches for a download URL in Private releases with improved version matching
func findPrivateDownload(releases []private.PrivateRelease, version string) (string, string, string) {
	// Parse target version
	targetMajor, targetMinor, targetPatch := parseVersionNumber(version)

	var bestMatch private.PrivateRelease
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		// Simple version parsing for private releases
		versionParts := strings.Split(release.Version, ".")
		if len(versionParts) == 0 {
			continue
		}

		major, err := strconv.Atoi(versionParts[0])
		if err != nil {
			continue
		}

		minor := 0
		if len(versionParts) > 1 {
			if m, err := strconv.Atoi(versionParts[1]); err == nil {
				minor = m
			}
		}

		patch := 0
		if len(versionParts) > 2 {
			if p, err := strconv.Atoi(versionParts[2]); err == nil {
				patch = p
			}
		}

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Take the first match
		bestMatch = release
		found = true
		break
	}

	if !found {
		return "", "", ""
	}

	url := bestMatch.DownloadURL
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	return url, filename, bestMatch.Version
}
