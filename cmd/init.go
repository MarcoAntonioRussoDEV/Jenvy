package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// Init inizializza l'ambiente JVM, incluso il completamento automatico per tutte le shell
func Init() {
	fmt.Println("🚀 Initializing Java Version Manager...")

	// Verifica se jvm è nel PATH
	if !isJvmInPath() {
		fmt.Println("⚠️  [WARNING] JVM executable not found in PATH")
		fmt.Println("💡 [INFO] Consider adding the JVM directory to your PATH for global access")
	}

	// Crea directory di configurazione se non esiste
	if err := createConfigDirectory(); err != nil {
		fmt.Printf("❌ [ERROR] Failed to create config directory: %v\n", err)
	} else {
		fmt.Println("✅ [SUCCESS] Configuration directory ready")
	}

	// Installa completamento per tutte le shell
	fmt.Println("\n📝 Setting up shell completions...")
	InstallCompletionForAllShells()

	// Controlla configurazione esistente
	fmt.Println("\n🔍 Checking current configuration...")
	if hasExistingConfig() {
		fmt.Println("✅ [SUCCESS] Configuration file found")
	} else {
		fmt.Println("💡 [INFO] No configuration file found - creating default configuration")
		if err := createDefaultConfig(); err != nil {
			fmt.Printf("⚠️  [WARNING] Failed to create default config: %v\n", err)
		} else {
			fmt.Println("✅ [SUCCESS] Default configuration created")
		}
	}

	fmt.Println("\n🎉 Java Version Manager initialization complete!")
	fmt.Println("💡 [INFO] Available commands:")
	fmt.Println("   jvm list              - Show installed JDK versions")
	fmt.Println("   jvm remote-list       - Show available JDK versions")
	fmt.Println("   jvm download <version> - Download and install a JDK version")
	fmt.Println("   jvm use <version>     - Set JAVA_HOME system-wide")
	fmt.Println("   jvm use <version>     - Set active JDK version")
	fmt.Println("   jvm help              - Show detailed help")
	fmt.Println("\n🔧 Restart your terminal to enable tab completion!")
}

// isJvmInPath controlla se l'eseguibile jvm è nel PATH
func isJvmInPath() bool {
	// Ottieni la directory dell'eseguibile corrente
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	exeDir := filepath.Dir(exePath)
	pathEnv := os.Getenv("PATH")

	paths := filepath.SplitList(pathEnv)
	for _, path := range paths {
		if filepath.Clean(path) == filepath.Clean(exeDir) {
			return true
		}
	}

	return false
}

// createConfigDirectory crea la directory di configurazione
func createConfigDirectory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".jvm")
	return os.MkdirAll(configDir, 0755)
}

// hasExistingConfig controlla se esiste già un file di configurazione
func hasExistingConfig() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	configPath := filepath.Join(homeDir, ".jvm", "config.json")
	_, err = os.Stat(configPath)
	return err == nil
}

// createDefaultConfig crea una configurazione di default
func createDefaultConfig() error {
	defaultConfig := `{
  "defaultProvider": "adoptium",
  "downloadPath": "",
  "privateRepositories": [],
  "lastCheck": "",
  "autoUpdate": true,
  "preferLTS": true
}`

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".jvm", "config.json")
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
