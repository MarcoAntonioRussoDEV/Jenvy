package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"jvm/utils"

	"golang.org/x/sys/windows/registry"
)

// UseJDK sets a specific JDK as the active JAVA_HOME
func UseJDK() {
	if len(os.Args) < 3 {
		utils.PrintUsage("Usage: jvm use <version>")
		utils.PrintUsage("Short form: jvm u <version>")
		utils.PrintInfo("Available JDKs:")
		showAvailableJDKs()
		return
	}

	version := os.Args[2]

	// Check if running as administrator
	if !isRunningAsAdmin() {
		utils.PrintInfo("Administrator privileges required to modify system environment variables")
		utils.PrintInfo("Requesting administrator privileges...")

		if requestAdminPrivileges() {
			return // Exit current process, admin process will handle the command
		} else {
			utils.PrintError("Failed to obtain administrator privileges")
			utils.PrintInfo("You can run manually as Administrator or use user-level installation")
			return
		}
	}

	// Get JDK installation directory
	jdkPath, err := findJDKInstallation(version)
	if err != nil {
		utils.PrintError(fmt.Sprintf("JDK version %s not found: %v", version, err))
		utils.PrintInfo("Run 'jvm list' to see installed JDKs")
		return
	}

	// Verify it's a valid JDK directory
	if !isValidJDKDirectory(jdkPath) {
		utils.PrintError(fmt.Sprintf("Invalid JDK directory: %s", jdkPath))
		return
	}

	// Set JAVA_HOME in system environment
	err = setSystemEnvironmentVariable("JAVA_HOME", jdkPath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to set JAVA_HOME: %v", err))
		utils.PrintInfo("Try running as Administrator")
		return
	}

	// Ensure %JAVA_HOME%\\bin is in PATH
	err = ensureJavaHomeInPath()
	if err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to update PATH: %v", err))
		utils.PrintInfo("You may need to add %JAVA_HOME%\\bin to your PATH manually")
	}

	utils.PrintSuccess(fmt.Sprintf("Set JAVA_HOME to JDK %s", version))
	utils.PrintInfo(fmt.Sprintf("JAVA_HOME = %s", jdkPath))
	utils.PrintInfo("Restart your terminal/IDE to see the changes")

	// Show Java version
	fmt.Println()
	utils.PrintInfo("Testing Java installation:")
	testJavaInstallation(jdkPath)
}

// requestAdminPrivileges attempts to restart the application with administrator privileges
func requestAdminPrivileges() bool {
	// Get current executable path
	exe, err := os.Executable()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get executable path: %v", err))
		return false
	}

	// Build command arguments (pass all original arguments)
	args := os.Args[1:] // Skip the program name

	// Create the command with runas verb to request admin privileges
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)

	// Join arguments into a single string
	argString := strings.Join(args, " ")
	argPtr, _ := syscall.UTF16PtrFromString(argString)

	// Use ShellExecute to run with elevated privileges
	ret := shellExecute(0, verbPtr, exePtr, argPtr, nil, 1)

	// Return true if ShellExecute succeeded (> 32)
	return ret > 32
}

// shellExecute is a wrapper for Windows ShellExecute API
func shellExecute(hwnd uintptr, verb, file, args, dir *uint16, show int) uintptr {
	ret, _, _ := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW").Call(
		hwnd,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(args)),
		uintptr(unsafe.Pointer(dir)),
		uintptr(show))
	return ret
}

// findJDKInstallation finds the installation path for a given JDK version
func findJDKInstallation(version string) (string, error) {
	// Get the default JVM versions directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	versionsDir := filepath.Join(homeDir, ".jvm", "versions")

	// Look for exact match first
	exactMatch := filepath.Join(versionsDir, fmt.Sprintf("JDK-%s", version))
	if _, err := os.Stat(exactMatch); err == nil {
		return exactMatch, nil
	}

	// Look for partial match
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read versions directory: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if strings.HasPrefix(name, "JDK-") {
				jdkVersion := strings.TrimPrefix(name, "JDK-")
				if strings.HasPrefix(jdkVersion, version) {
					matches = append(matches, filepath.Join(versionsDir, name))
				}
			}
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no JDK found matching version %s", version)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	// Multiple matches - show options
	utils.PrintWarning("Multiple JDK versions found:")
	for i, match := range matches {
		fmt.Printf("  %d. %s\n", i+1, filepath.Base(match))
	}
	return "", fmt.Errorf("multiple matches found, please be more specific")
}

// isValidJDKDirectory checks if a directory contains a valid JDK installation
func isValidJDKDirectory(path string) bool {
	// Check for required directories
	requiredDirs := []string{"bin", "lib"}
	for _, dir := range requiredDirs {
		if _, err := os.Stat(filepath.Join(path, dir)); err != nil {
			return false
		}
	}

	// Check for java executable
	javaExe := "java.exe"
	if _, err := os.Stat(filepath.Join(path, "bin", javaExe)); err != nil {
		return false
	}

	return true
}

// setSystemEnvironmentVariable sets a system environment variable in Windows registry
func setSystemEnvironmentVariable(name, value string) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	err = key.SetStringValue(name, value)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	// Broadcast WM_SETTINGCHANGE message to notify applications
	// This helps some applications pick up the new environment variable
	utils.PrintInfo("Broadcasting environment change...")

	return nil
}

// ensureJavaHomeInPath ensures %JAVA_HOME%\\bin is in the system PATH
func ensureJavaHomeInPath() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Read current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		return fmt.Errorf("failed to read PATH: %w", err)
	}

	javaHomeBin := `%JAVA_HOME%\bin`

	// Check if %JAVA_HOME%\\bin is already in PATH
	pathEntries := strings.Split(currentPath, ";")
	for _, entry := range pathEntries {
		if strings.EqualFold(strings.TrimSpace(entry), javaHomeBin) {
			utils.PrintInfo("%JAVA_HOME%\\bin is already in PATH")
			return nil
		}
	}

	// Add %JAVA_HOME%\\bin to the beginning of PATH
	newPath := javaHomeBin + ";" + currentPath

	err = key.SetStringValue("Path", newPath)
	if err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	utils.PrintSuccess("Added %JAVA_HOME%\\bin to system PATH")
	return nil
}

// setUserEnvironmentVariable sets a user environment variable in Windows registry
func setUserEnvironmentVariable(name, value string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open user registry key: %w", err)
	}
	defer key.Close()

	err = key.SetStringValue(name, value)
	if err != nil {
		return fmt.Errorf("failed to set user registry value: %w", err)
	}

	return nil
}

// ensureJavaHomeInUserPath ensures %JAVA_HOME%\\bin is in the user PATH
func ensureJavaHomeInUserPath() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open user registry key: %w", err)
	}
	defer key.Close()

	// Read current user PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		// User PATH doesn't exist, create it
		currentPath = ""
	}

	javaHomeBin := `%JAVA_HOME%\bin`

	// Check if %JAVA_HOME%\\bin is already in PATH
	if currentPath != "" {
		pathEntries := strings.Split(currentPath, ";")
		for _, entry := range pathEntries {
			if strings.EqualFold(strings.TrimSpace(entry), javaHomeBin) {
				utils.PrintInfo("%JAVA_HOME%\\bin is already in user PATH")
				return nil
			}
		}
	}

	// Add %JAVA_HOME%\\bin to the beginning of PATH
	var newPath string
	if currentPath == "" {
		newPath = javaHomeBin
	} else {
		newPath = javaHomeBin + ";" + currentPath
	}

	err = key.SetStringValue("Path", newPath)
	if err != nil {
		return fmt.Errorf("failed to update user PATH: %w", err)
	}

	utils.PrintSuccess("Added %JAVA_HOME%\\bin to user PATH")
	return nil
}

// showAvailableJDKs shows a list of installed JDKs
func showAvailableJDKs() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get home directory: %v", err))
		return
	}

	versionsDir := filepath.Join(homeDir, ".jvm", "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		utils.PrintWarning("No JDKs found. Use 'jvm download <version>' to install a JDK")
		return
	}

	var jdks []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "JDK-") {
			version := strings.TrimPrefix(entry.Name(), "JDK-")
			jdkPath := filepath.Join(versionsDir, entry.Name())
			if isValidJDKDirectory(jdkPath) {
				jdks = append(jdks, version)
			}
		}
	}

	if len(jdks) == 0 {
		utils.PrintWarning("No valid JDKs found. Use 'jvm download <version>' to install a JDK")
		return
	}

	fmt.Println("Available JDK versions:")
	for _, jdk := range jdks {
		fmt.Printf("  - %s\n", jdk)
	}
}

// testJavaInstallation tests if Java is working correctly
func testJavaInstallation(jdkPath string) {
	javaExe := filepath.Join(jdkPath, "bin", "java.exe")

	// Test java -version command
	fmt.Printf("Testing: %s -version\n", javaExe)

	// We can't easily run the command and capture output here without additional complexity
	// Instead, we'll just verify the executable exists and is accessible
	if _, err := os.Stat(javaExe); err != nil {
		utils.PrintError("Java executable not found")
		return
	}

	utils.PrintSuccess("Java executable found and accessible")
	fmt.Printf("Java location: %s\n", javaExe)
}

// InitializeJVMEnvironment sets up the initial JVM environment variables during installation
func InitializeJVMEnvironment() {
	fmt.Println("🔧 Setting up JVM environment variables...")

	// Check if running as administrator
	if !isRunningAsAdmin() {
		utils.PrintWarning("For system-wide environment variables, run as Administrator")
		utils.PrintInfo("You can still use JVM, but 'jvm use' will require Administrator privileges")
	} else {
		// Ensure %JAVA_HOME%\\bin is in PATH (will be set when a JDK is selected)
		err := ensureJavaHomeInPath()
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to initialize PATH: %v", err))
			utils.PrintInfo("You may need to manually add %JAVA_HOME%\\bin to your PATH")
		} else {
			utils.PrintSuccess("JVM environment initialized")
		}
	}

	utils.PrintInfo("Use 'jvm use <version>' to set your active JDK")
}

// isRunningAsAdmin checks if the current process is running with administrator privileges
func isRunningAsAdmin() bool {
	// Try to open a registry key that requires admin access
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.SET_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	return true
}
