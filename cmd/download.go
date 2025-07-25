package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	"jvm/utils"
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
		utils.PrintError("No JDK version specified")
		utils.PrintInfo("Usage: jvm download <version> [options]")
		utils.PrintInfo("Examples:")
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
		utils.PrintError(fmt.Sprintf("Failed to determine download directory: %v", dirErr))
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

	fmt.Printf("%s Searching for JDK version %s from provider: %s\n",
		utils.ColorText("[>]", utils.BrightCyan), version, provider)
	fmt.Printf("%s Download directory: %s\n\n",
		utils.ColorText("[>]", utils.BrightCyan), outputDir)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create output directory: %v", err))
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
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findAdoptiumDownload(releases, version)

	case "azul":
		releases, err := azul.GetAzulJDKs()
		if err != nil {
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findAzulDownload(releases, version)

	case "liberica":
		releases, err := liberica.GetLibericaJDKs()
		if err != nil {
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findLibericaDownload(releases, version)

	case "private":
		releases, err := private.GetPrivateJDKs()
		if err != nil {
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findPrivateDownload(releases, version)

	default:
		fmt.Printf("[ERROR] Unknown provider: %s\n", provider)
		fmt.Println("[INFO] Available providers: adoptium, azul, liberica, private")
		return
	}

	if downloadURL == "" {
		fmt.Printf("[ERROR] JDK version %s not found in %s provider\n", version, provider)
		fmt.Println("[INFO] Try running 'jvm remote-list' to see available versions")
		return
	}

	if filename == "" {
		filename = fmt.Sprintf("openjdk-%s.tar.gz", version)
	}

	// Create a version-specific subdirectory
	versionDir := fmt.Sprintf("JDK-%s", foundVersion)
	versionOutputDir := filepath.Join(outputDir, versionDir)

	// Create version-specific directory
	if err := os.MkdirAll(versionOutputDir, 0755); err != nil {
		fmt.Printf("[ERROR] Failed to create version directory: %v\n", err)
		return
	}

	outputPath := filepath.Join(versionOutputDir, filename)

	fmt.Printf("%s JDK %s\n", utils.ColorText("[FOUND]", utils.BrightGreen), foundVersion)
	fmt.Printf("%s Download URL: %s\n", utils.ColorText("[URL]", utils.BrightBlue), downloadURL)
	fmt.Printf("%s Version directory: %s\n", utils.ColorText("[DIR]", utils.BrightYellow), versionOutputDir)
	fmt.Printf("%s Saving to: %s\n", utils.ColorText("[FILE]", utils.BrightMagenta), outputPath)

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		utils.PrintWarning(fmt.Sprintf("File already exists: %s", filename))
	}

	// Ask for confirmation
	fmt.Print("\n[?] Do you want to proceed with the download? (y/N): ")
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		utils.PrintInfo("Download cancelled by user")
		return
	}

	fmt.Println()

	// Download the file
	if err := downloadFile(downloadURL, outputPath); err != nil {
		utils.PrintError(fmt.Sprintf("Download failed: %v", err))
		return
	}

	utils.PrintSuccess("Download completed successfully!")
	fmt.Printf("%s JDK %s saved to: %s\n",
		utils.ColorText("[OUTPUT]", utils.BrightGreen), foundVersion, versionOutputDir)
	fmt.Printf("%s Archive file: %s\n",
		utils.ColorText("[FILE]", utils.BrightBlue), filename)

	// Show file info
	if fileInfo, err := os.Stat(outputPath); err == nil {
		fmt.Printf("[SIZE] File size: %.2f MB\n", float64(fileInfo.Size())/1024/1024)
		fmt.Printf("[TIME] Download time: %s\n", time.Now().Format("15:04:05"))
	}

	// Extract the archive automatically
	fmt.Printf("\n[EXTRACT] Extracting JDK archive...\n")
	extractPath := versionOutputDir // Extract directly to version directory

	if err := extractArchive(outputPath, extractPath); err != nil {
		fmt.Printf("[WARN] Extraction failed: %v\n", err)
		fmt.Printf("[INFO] You can manually extract: %s\n", outputPath)
	} else {
		fmt.Printf("[SUCCESS] JDK extracted successfully to: %s\n", extractPath)

		// Try to find the actual JDK root directory and move it to a cleaner path
		jdkRootDir, err := findJDKRootDir(extractPath)
		if err == nil {
			// If we found a nested JDK directory, move its contents up one level
			if jdkRootDir != extractPath {
				if err := flattenJDKDirectory(jdkRootDir, extractPath, outputPath); err == nil {
					fmt.Printf("[READY] JDK ready at: %s\n", extractPath)
					fmt.Printf("[PATH] Add to PATH: %s\n", filepath.Join(extractPath, "bin"))
				} else {
					fmt.Printf("[READY] JDK ready at: %s\n", jdkRootDir)
					fmt.Printf("[PATH] Add to PATH: %s\n", filepath.Join(jdkRootDir, "bin"))
				}
			} else {
				fmt.Printf("[READY] JDK ready at: %s\n", extractPath)
				fmt.Printf("[PATH] Add to PATH: %s\n", filepath.Join(extractPath, "bin"))
			}
		}

		// Optionally remove the archive file
		fmt.Printf("\n[CLEAN] Remove archive file? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) == "y" {
			if err := os.Remove(outputPath); err == nil {
				fmt.Printf("[CLEAN] Archive file removed\n")
			}
		}
	}

	fmt.Println()
	fmt.Println("[INFO] Next steps:")
	fmt.Printf("   jvm list                   # View downloaded JDKs\n")
	fmt.Printf("   jvm extract %s            # Extract the archive (coming soon)\n", foundVersion)
	fmt.Printf("   jvm use %s                # Set as active JDK (coming soon)\n", foundVersion)
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

	fmt.Println("[DOWNLOAD] Downloading...")
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

				fmt.Printf("\r[DOWNLOAD] Progress: %.1f%% (%.2f MB / %.2f MB) - Speed: %.2f MB/s",
					progress,
					float64(downloaded)/1024/1024,
					float64(contentLength)/1024/1024,
					speed,
				)
			} else {
				// Show downloaded amount without percentage
				elapsed := time.Since(startTime)
				speed := float64(downloaded) / elapsed.Seconds() / 1024 / 1024 // MB/s

				fmt.Printf("\r[DOWNLOAD] Downloaded: %.2f MB - Speed: %.2f MB/s",
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

// extractArchive extracts ZIP or TAR.GZ archives
func extractArchive(archivePath, destPath string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(archivePath))
	if ext == ".zip" {
		return extractZip(archivePath, destPath)
	} else if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
		return extractTarGz(archivePath, destPath)
	}

	return fmt.Errorf("unsupported archive format: %s", ext)
}

// extractZip extracts a ZIP archive
func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Clean the file path to prevent zip slip attacks
		cleanPath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(cleanPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(cleanPath, 0755)
			continue
		}

		// Create the directories for file
		if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
			return err
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarGz extracts a TAR.GZ archive
func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Clean the file path to prevent tar slip attacks
		cleanPath := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(cleanPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create the directories for file
			if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
				return err
			}

			// Extract file
			outFile, err := os.Create(cleanPath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}

			outFile.Close()

			// Set file permissions
			if err := os.Chmod(cleanPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}

// findJDKRootDir finds the actual JDK root directory within the extracted path
func findJDKRootDir(extractPath string) (string, error) {
	// JDK archives often contain a single root directory like "jdk-17.0.5+8"
	entries, err := os.ReadDir(extractPath)
	if err != nil {
		return extractPath, err
	}

	// Look for a single directory that might be the JDK root
	var jdkDir string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this directory contains typical JDK structure (bin, lib, etc.)
			potentialJDKDir := filepath.Join(extractPath, entry.Name())
			if isJDKDirectory(potentialJDKDir) {
				jdkDir = potentialJDKDir
				break
			}
		}
	}

	if jdkDir == "" {
		// If no JDK-like subdirectory found, check if extractPath itself is a JDK
		if isJDKDirectory(extractPath) {
			return extractPath, nil
		}
		return extractPath, fmt.Errorf("could not locate JDK root directory")
	}

	return jdkDir, nil
}

// isJDKDirectory checks if a directory looks like a JDK installation
func isJDKDirectory(dir string) bool {
	// Check for typical JDK directories
	requiredDirs := []string{"bin", "lib"}
	for _, reqDir := range requiredDirs {
		if _, err := os.Stat(filepath.Join(dir, reqDir)); err != nil {
			return false
		}
	}

	// Check for java executable
	javaExe := "java"
	if runtime.GOOS == "windows" {
		javaExe = "java.exe"
	}

	if _, err := os.Stat(filepath.Join(dir, "bin", javaExe)); err != nil {
		return false
	}

	return true
}

// flattenJDKDirectory moves the contents of a nested JDK directory up to the parent level
func flattenJDKDirectory(jdkRootDir, targetDir, archivePath string) error {
	// Create a temporary directory to avoid conflicts
	tempDir := targetDir + "_temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Move JDK contents to temp directory
	entries, err := os.ReadDir(jdkRootDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(jdkRootDir, entry.Name())
		destPath := filepath.Join(tempDir, entry.Name())

		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move %s: %v", entry.Name(), err)
		}
	}

	// Remove the now empty nested directory structure
	if err := os.RemoveAll(filepath.Join(targetDir, filepath.Base(jdkRootDir))); err != nil {
		return err
	}

	// Move contents from temp to target directory
	tempEntries, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	for _, entry := range tempEntries {
		srcPath := filepath.Join(tempDir, entry.Name())
		destPath := filepath.Join(targetDir, entry.Name())

		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move %s to final location: %v", entry.Name(), err)
		}
	}

	return nil
}
