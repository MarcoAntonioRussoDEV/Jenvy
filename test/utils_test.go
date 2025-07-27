package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jenvy/internal/utils"
)

// TestParseVersionNumber verifica il parsing delle versioni Java
func TestParseVersionNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		major int
		minor int
		patch int
	}{
		{"Major only", "17", 17, 0, 0},
		{"Major.Minor", "17.0", 17, 0, 0},
		{"Full version", "21.0.2", 21, 0, 2},
		{"LTS version", "11.0.21", 11, 0, 21},
		{"Java 8 style", "1.8.0", 8, 0, 0},
		{"Java 8 with update", "1.8.0_452", 8, 0, 452},
		{"Java 8 with build", "1.8.0_452-b09", 8, 0, 452},
		{"Modern Java 8", "8.0.392", 8, 0, 392},
		{"Liberica Java 8", "8u352", 8, 0, 352},
		{"Complex version", "17.0.5+9", 17, 0, 5},
		{"Invalid version", "invalid", -1, -1, -1},
		{"Empty string", "", 0, -1, -1},
		{"Single number", "8", 8, 0, 0},
		{"Build suffix", "21.0.2+13", 21, 0, 2},
		{"Beta version", "22.0.0-ea", 22, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch := utils.ParseVersionNumber(tt.input)
			if major != tt.major || minor != tt.minor || patch != tt.patch {
				t.Errorf("ParseVersionNumber(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.input, major, minor, patch, tt.major, tt.minor, tt.patch)
			}
		})
	}
}

// TestIsLTSVersion verifica il rilevamento delle versioni LTS
func TestIsLTSVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"Java 8 modern", "8.0.392", true},
		{"Java 8 legacy", "1.8.0_452", true},
		{"Java 8 Liberica", "8u352", true},
		{"Java 11 LTS", "11.0.20", true},
		{"Java 17 LTS", "17.0.5", true},
		{"Java 21 LTS", "21.0.2", true},
		{"Java 19 non-LTS", "19.0.2", false},
		{"Java 20 non-LTS", "20.0.1", false},
		{"Java 22 non-LTS", "22.0.0", false},
		{"LTS marker explicit", "22-lts", true},
		{"LTS marker mixed case", "23.0.1-LTS", true},
		{"Single major LTS", "17", true},
		{"Single major non-LTS", "19", false},
		{"Invalid version", "invalid", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsLTSVersion(tt.version)
			if result != tt.expected {
				t.Errorf("IsLTSVersion(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

// TestIsValidJDKDirectory verifica la validazione delle directory JDK
func TestIsValidJDKDirectory(t *testing.T) {
	// Crea una directory temporanea per i test
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func(string) string
		expected bool
	}{
		{
			name: "Valid JDK directory",
			setup: func(baseDir string) string {
				jdkDir := filepath.Join(baseDir, "valid-jdk")
				os.MkdirAll(filepath.Join(jdkDir, "bin"), 0755)
				os.MkdirAll(filepath.Join(jdkDir, "lib"), 0755)
				// Crea java.exe
				javaExe := filepath.Join(jdkDir, "bin", "java.exe")
				file, _ := os.Create(javaExe)
				file.Close()
				return jdkDir
			},
			expected: true,
		},
		{
			name: "Missing bin directory",
			setup: func(baseDir string) string {
				jdkDir := filepath.Join(baseDir, "no-bin")
				os.MkdirAll(filepath.Join(jdkDir, "lib"), 0755)
				return jdkDir
			},
			expected: false,
		},
		{
			name: "Missing lib directory",
			setup: func(baseDir string) string {
				jdkDir := filepath.Join(baseDir, "no-lib")
				os.MkdirAll(filepath.Join(jdkDir, "bin"), 0755)
				javaExe := filepath.Join(jdkDir, "bin", "java.exe")
				file, _ := os.Create(javaExe)
				file.Close()
				return jdkDir
			},
			expected: false,
		},
		{
			name: "Missing java.exe",
			setup: func(baseDir string) string {
				jdkDir := filepath.Join(baseDir, "no-java-exe")
				os.MkdirAll(filepath.Join(jdkDir, "bin"), 0755)
				os.MkdirAll(filepath.Join(jdkDir, "lib"), 0755)
				return jdkDir
			},
			expected: false,
		},
		{
			name: "Non-existent directory",
			setup: func(baseDir string) string {
				return filepath.Join(baseDir, "non-existent")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := tt.setup(tempDir)
			result := utils.IsValidJDKDirectory(testDir)
			if result != tt.expected {
				t.Errorf("IsValidJDKDirectory(%q) = %v, want %v", testDir, result, tt.expected)
			}
		})
	}
}

// TestGetJenvyVersionsDirectory verifica la creazione del percorso directory versioni
func TestGetJenvyVersionsDirectory(t *testing.T) {
	dir, err := utils.GetJenvyVersionsDirectory()
	if err != nil {
		t.Fatalf("GetJenvyVersionsDirectory() failed: %v", err)
	}

	if dir == "" {
		t.Error("GetJenvyVersionsDirectory() returned empty string")
	}

	// Verifica che il path contenga .jenvy/versions
	if !filepath.IsAbs(dir) {
		t.Error("GetJenvyVersionsDirectory() should return absolute path")
	}

	expectedSuffix := filepath.Join(".jenvy", "versions")
	if !strings.HasSuffix(dir, expectedSuffix) {
		t.Errorf("GetJenvyVersionsDirectory() = %q, should end with %q", dir, expectedSuffix)
	}
}

// BenchmarkParseVersionNumber benchmark per performance del parsing versioni
func BenchmarkParseVersionNumber(b *testing.B) {
	versions := []string{"17", "17.0", "21.0.2", "11.0.21", "1.8.0_392"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		version := versions[i%len(versions)]
		utils.ParseVersionNumber(version)
	}
}

// BenchmarkIsValidJDKDirectory benchmark per performance della validazione directory
func BenchmarkIsValidJDKDirectory(b *testing.B) {
	// Setup: crea una directory JDK valida per il benchmark
	tempDir := b.TempDir()
	jdkDir := filepath.Join(tempDir, "benchmark-jdk")
	os.MkdirAll(filepath.Join(jdkDir, "bin"), 0755)
	os.MkdirAll(filepath.Join(jdkDir, "lib"), 0755)
	javaExe := filepath.Join(jdkDir, "bin", "java.exe")
	file, _ := os.Create(javaExe)
	file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.IsValidJDKDirectory(jdkDir)
	}
}

// TestFindSingleJDKInstallation verifica la ricerca di installazioni JDK
func TestFindSingleJDKInstallation(t *testing.T) {
	// Skip test se non in ambiente con JDK installati
	if testing.Short() {
		t.Skip("Skipping JDK installation test in short mode")
	}

	// Questo test richiede JDK effettivamente installati
	// Potrebbe fallire in ambienti CI senza JDK pre-installati
	_, err := utils.FindSingleJDKInstallation("999") // Versione che non dovrebbe esistere
	if err == nil {
		t.Error("FindSingleJDKInstallation('999') should return error for non-existent version")
	}
}
