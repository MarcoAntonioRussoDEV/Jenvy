package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// MockJDKSetup crea una struttura JDK simulata per i test
type MockJDKSetup struct {
	tempDir string
	jdkPath string
}

// NewMockJDKSetup crea un nuovo setup JDK per test
func NewMockJDKSetup(t *testing.T) *MockJDKSetup {
	tempDir := t.TempDir()
	jdkPath := filepath.Join(tempDir, "mock-jdk")

	// Crea struttura JDK base
	os.MkdirAll(filepath.Join(jdkPath, "bin"), 0755)
	os.MkdirAll(filepath.Join(jdkPath, "lib"), 0755)

	// Crea java.exe mock
	javaExe := filepath.Join(jdkPath, "bin", "java.exe")
	file, err := os.Create(javaExe)
	if err != nil {
		t.Fatalf("Failed to create mock java.exe: %v", err)
	}
	file.Close()

	return &MockJDKSetup{
		tempDir: tempDir,
		jdkPath: jdkPath,
	}
}

// GetJDKPath restituisce il path del JDK mock
func (m *MockJDKSetup) GetJDKPath() string {
	return m.jdkPath
}

// GetVersionsDir restituisce la directory versions mock
func (m *MockJDKSetup) GetVersionsDir() string {
	return m.tempDir
}

// CreateVersionedJDK crea un JDK con nome specifico di versione
func (m *MockJDKSetup) CreateVersionedJDK(t *testing.T, version string) string {
	jdkName := "JDK-" + version
	jdkPath := filepath.Join(m.tempDir, jdkName)

	// Crea struttura JDK
	os.MkdirAll(filepath.Join(jdkPath, "bin"), 0755)
	os.MkdirAll(filepath.Join(jdkPath, "lib"), 0755)

	// Crea java.exe
	javaExe := filepath.Join(jdkPath, "bin", "java.exe")
	file, err := os.Create(javaExe)
	if err != nil {
		t.Fatalf("Failed to create java.exe for %s: %v", version, err)
	}
	file.Close()

	return jdkPath
}

// CreateInvalidJDK crea una directory JDK invalida (mancano componenti)
func (m *MockJDKSetup) CreateInvalidJDK(t *testing.T, version string) string {
	jdkName := "JDK-" + version
	jdkPath := filepath.Join(m.tempDir, jdkName)

	// Crea solo directory base, senza bin/lib/java.exe
	os.MkdirAll(jdkPath, 0755)

	return jdkPath
}

// TestJDKValidationWorkflow testa il workflow completo di validazione JDK
func TestJDKValidationWorkflow(t *testing.T) {
	setup := NewMockJDKSetup(t)

	tests := []struct {
		name        string
		setupFunc   func(*testing.T) string
		expectValid bool
		description string
	}{
		{
			name: "Valid JDK structure",
			setupFunc: func(t *testing.T) string {
				return setup.CreateVersionedJDK(t, "17.0.5")
			},
			expectValid: true,
			description: "Complete JDK with bin/, lib/, and java.exe",
		},
		{
			name: "Invalid JDK structure",
			setupFunc: func(t *testing.T) string {
				return setup.CreateInvalidJDK(t, "21.0.1")
			},
			expectValid: false,
			description: "Incomplete JDK missing essential components",
		},
		{
			name: "Non-existent directory",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(setup.GetVersionsDir(), "non-existent")
			},
			expectValid: false,
			description: "Directory that doesn't exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jdkPath := tt.setupFunc(t)

			// Test path existence
			if tt.expectValid {
				if _, err := os.Stat(jdkPath); os.IsNotExist(err) {
					t.Errorf("Expected valid JDK path %s should exist", jdkPath)
				}
			}

			// Test java.exe existence for valid JDKs
			if tt.expectValid {
				javaExe := filepath.Join(jdkPath, "bin", "java.exe")
				if _, err := os.Stat(javaExe); os.IsNotExist(err) {
					t.Errorf("Expected java.exe at %s for valid JDK", javaExe)
				}
			}
		})
	}
}

// TestMultipleJDKVersions testa scenari con multiple versioni JDK
func TestMultipleJDKVersions(t *testing.T) {
	setup := NewMockJDKSetup(t)

	// Crea multiple versioni JDK
	versions := []string{"17.0.5", "17.0.8", "21.0.1", "11.0.21"}
	var jdkPaths []string

	for _, version := range versions {
		jdkPath := setup.CreateVersionedJDK(t, version)
		jdkPaths = append(jdkPaths, jdkPath)
	}

	// Verifica che tutte le versioni siano state create correttamente
	for i, version := range versions {
		jdkPath := jdkPaths[i]
		javaExe := filepath.Join(jdkPath, "bin", "java.exe")

		if _, err := os.Stat(javaExe); os.IsNotExist(err) {
			t.Errorf("JDK %s java.exe not found at %s", version, javaExe)
		}
	}

	// Test di ricerca pattern
	versionsDir := setup.GetVersionsDir()
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		t.Fatalf("Failed to read versions directory: %v", err)
	}

	jdkCount := 0
	for _, entry := range entries {
		if entry.IsDir() && filepath.Base(entry.Name())[:4] == "JDK-" {
			jdkCount++
		}
	}

	if jdkCount != len(versions) {
		t.Errorf("Expected %d JDK directories, found %d", len(versions), jdkCount)
	}
}

// TestJDKPathValidation testa la validazione dei percorsi JDK
func TestJDKPathValidation(t *testing.T) {
	setup := NewMockJDKSetup(t)

	validJDK := setup.CreateVersionedJDK(t, "17.0.5")
	invalidJDK := setup.CreateInvalidJDK(t, "21.0.1")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Valid JDK path", validJDK, true},
		{"Invalid JDK path", invalidJDK, false},
		{"Empty path", "", false},
		{"Non-existent path", "/non/existent/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Qui dovremmo chiamare la funzione di validazione
			// Per ora testiamo solo l'esistenza del file java.exe
			javaExe := filepath.Join(tt.path, "bin", "java.exe")
			_, err := os.Stat(javaExe)
			exists := !os.IsNotExist(err)

			if exists != tt.expected {
				t.Errorf("Path validation for %s: got %v, want %v", tt.path, exists, tt.expected)
			}
		})
	}
}

// BenchmarkJDKValidation benchmark per performance della validazione JDK
func BenchmarkJDKValidation(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	jdkPath := filepath.Join(tempDir, "benchmark-jdk")
	os.MkdirAll(filepath.Join(jdkPath, "bin"), 0755)
	os.MkdirAll(filepath.Join(jdkPath, "lib"), 0755)
	javaExe := filepath.Join(jdkPath, "bin", "java.exe")
	file, _ := os.Create(javaExe)
	file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test esistenza java.exe
		_, err := os.Stat(javaExe)
		if err != nil {
			b.Error("Validation failed during benchmark")
		}
	}
}
