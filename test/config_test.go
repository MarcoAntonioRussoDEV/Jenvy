package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestConfigurationHandling testa la gestione della configurazione
func TestConfigurationHandling(t *testing.T) {
	// Crea directory temporanea per i test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Test: creazione configurazione
	t.Run("Create Config", func(t *testing.T) {
		config := map[string]interface{}{
			"private_endpoint": "https://nexus.company.com/api/jdk",
			"private_token":    "test-token-123",
			"provider":         "adoptium",
		}

		configData, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		err = os.WriteFile(configPath, configData, 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Verifica che il file sia stato creato
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}
	})

	// Test: lettura configurazione
	t.Run("Read Config", func(t *testing.T) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var config map[string]interface{}
		err = json.Unmarshal(data, &config)
		if err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		// Verifica campi
		if config["private_endpoint"] != "https://nexus.company.com/api/jdk" {
			t.Error("private_endpoint not correctly stored")
		}

		if config["private_token"] != "test-token-123" {
			t.Error("private_token not correctly stored")
		}
	})

	// Test: aggiornamento configurazione
	t.Run("Update Config", func(t *testing.T) {
		// Leggi configurazione esistente
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var config map[string]interface{}
		err = json.Unmarshal(data, &config)
		if err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		// Modifica configurazione
		config["provider"] = "azul"
		config["private_token"] = "new-token-456"

		// Scrivi configurazione aggiornata
		updatedData, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal updated config: %v", err)
		}

		err = os.WriteFile(configPath, updatedData, 0644)
		if err != nil {
			t.Fatalf("Failed to write updated config: %v", err)
		}

		// Verifica modifiche
		data, err = os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read updated config: %v", err)
		}

		err = json.Unmarshal(data, &config)
		if err != nil {
			t.Fatalf("Failed to unmarshal updated config: %v", err)
		}

		if config["provider"] != "azul" {
			t.Error("provider was not updated correctly")
		}

		if config["private_token"] != "new-token-456" {
			t.Error("private_token was not updated correctly")
		}
	})
}

// TestPrivateRepositoryConfig testa la configurazione repository privati
func TestPrivateRepositoryConfig(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		endpoint string
		token    string
		valid    bool
	}{
		{
			name:     "Valid HTTPS endpoint with token",
			endpoint: "https://nexus.company.com/api/jdk",
			token:    "abc123xyz",
			valid:    true,
		},
		{
			name:     "Valid HTTPS endpoint without token",
			endpoint: "https://repository.internal.com/jdk-api",
			token:    "",
			valid:    true,
		},
		{
			name:     "Invalid HTTP endpoint",
			endpoint: "http://insecure.com/api",
			token:    "token123",
			valid:    false,
		},
		{
			name:     "Empty endpoint",
			endpoint: "",
			token:    "token123",
			valid:    false,
		},
		{
			name:     "Invalid URL format",
			endpoint: "not-a-url",
			token:    "token123",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{
				"private_endpoint": tt.endpoint,
				"private_token":    tt.token,
			}

			// Validazione endpoint (simulata)
			isValidEndpoint := tt.endpoint != "" &&
				len(tt.endpoint) > 8 &&
				tt.endpoint[:8] == "https://" // Solo HTTPS è valido

			if isValidEndpoint != tt.valid {
				t.Errorf("Endpoint validation for %s: got %v, want %v",
					tt.endpoint, isValidEndpoint, tt.valid)
			}

			if tt.valid {
				// Se valido, salva configurazione
				configData, err := json.Marshal(config)
				if err != nil {
					t.Fatalf("Failed to marshal config: %v", err)
				}

				testConfigPath := filepath.Join(tempDir, tt.name+".json")
				err = os.WriteFile(testConfigPath, configData, 0644)
				if err != nil {
					t.Fatalf("Failed to write config: %v", err)
				}

				// Verifica salvataggio
				if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
					t.Error("Valid config was not saved")
				}
			}
		})
	}
}

// TestConfigurationValidation testa la validazione della configurazione
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		valid  bool
	}{
		{
			name: "Complete valid config",
			config: map[string]interface{}{
				"private_endpoint": "https://nexus.company.com/api/jdk",
				"private_token":    "valid-token",
				"provider":         "adoptium",
			},
			valid: true,
		},
		{
			name: "Missing private_token",
			config: map[string]interface{}{
				"private_endpoint": "https://nexus.company.com/api/jdk",
				"provider":         "adoptium",
			},
			valid: true, // token is optional
		},
		{
			name: "Invalid provider",
			config: map[string]interface{}{
				"private_endpoint": "https://nexus.company.com/api/jdk",
				"private_token":    "valid-token",
				"provider":         "invalid-provider",
			},
			valid: false,
		},
		{
			name:   "Empty config",
			config: map[string]interface{}{},
			valid:  true, // empty config is valid (uses defaults)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validazione simulata
			isValid := true

			// Verifica provider se presente
			if provider, exists := tt.config["provider"]; exists {
				validProviders := []string{"adoptium", "azul", "liberica", "private"}
				providerValid := false
				for _, valid := range validProviders {
					if provider == valid {
						providerValid = true
						break
					}
				}
				if !providerValid {
					isValid = false
				}
			}

			if isValid != tt.valid {
				t.Errorf("Config validation for %s: got %v, want %v",
					tt.name, isValid, tt.valid)
			}
		})
	}
}

// TestDefaultConfiguration testa la configurazione di default
func TestDefaultConfiguration(t *testing.T) {
	defaultConfig := map[string]interface{}{
		"provider": "adoptium",
		"lts_only": false,
	}

	// Verifica campi default
	if defaultConfig["provider"] != "adoptium" {
		t.Error("Default provider should be 'adoptium'")
	}

	if defaultConfig["lts_only"] != false {
		t.Error("Default lts_only should be false")
	}

	// Test: unione con configurazione utente
	userConfig := map[string]interface{}{
		"private_endpoint": "https://company.com/api",
		"lts_only":         true,
	}

	// Simula merge di configurazione
	mergedConfig := make(map[string]interface{})
	for k, v := range defaultConfig {
		mergedConfig[k] = v
	}
	for k, v := range userConfig {
		mergedConfig[k] = v
	}

	// Verifica merge
	if mergedConfig["provider"] != "adoptium" {
		t.Error("Default provider should be preserved")
	}

	if mergedConfig["lts_only"] != true {
		t.Error("User lts_only should override default")
	}

	if mergedConfig["private_endpoint"] != "https://company.com/api" {
		t.Error("User private_endpoint should be added")
	}
}

// TestConfigurationPersistence testa la persistenza della configurazione
func TestConfigurationPersistence(t *testing.T) {
	tempDir := t.TempDir()

	// Simula ottenimento directory di configurazione
	configDir := filepath.Join(tempDir, ".jvm")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	// Test persistenza attraverso restart simulati
	configs := []map[string]interface{}{
		{
			"provider": "adoptium",
			"version":  "1.0.0",
		},
		{
			"provider":         "azul",
			"private_endpoint": "https://company.com/api",
			"version":          "1.0.1",
		},
		{
			"provider":      "liberica",
			"private_token": "updated-token",
			"version":       "1.0.2",
		},
	}

	for i, config := range configs {
		t.Run(fmt.Sprintf("Persistence_Test_%d", i+1), func(t *testing.T) {
			// Salva configurazione
			configData, err := json.Marshal(config)
			if err != nil {
				t.Fatalf("Failed to marshal config: %v", err)
			}

			err = os.WriteFile(configPath, configData, 0644)
			if err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			// Simula restart: rileggi configurazione
			data, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read config after restart: %v", err)
			}

			var loadedConfig map[string]interface{}
			err = json.Unmarshal(data, &loadedConfig)
			if err != nil {
				t.Fatalf("Failed to unmarshal config after restart: %v", err)
			}

			// Verifica persistenza
			for key, expectedValue := range config {
				if loadedConfig[key] != expectedValue {
					t.Errorf("Config key %s: got %v, want %v",
						key, loadedConfig[key], expectedValue)
				}
			}
		})
	}
}

// BenchmarkConfigurationOperations benchmark per operazioni di configurazione
func BenchmarkConfigurationOperations(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	config := map[string]interface{}{
		"private_endpoint": "https://nexus.company.com/api/jdk",
		"private_token":    "benchmark-token",
		"provider":         "adoptium",
		"lts_only":         true,
	}

	b.Run("Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(config)
			if err != nil {
				b.Error("Marshal failed during benchmark")
			}
		}
	})

	// Salva configurazione per test di lettura
	configData, _ := json.Marshal(config)
	os.WriteFile(configPath, configData, 0644)

	b.Run("Unmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, err := os.ReadFile(configPath)
			if err != nil {
				b.Error("ReadFile failed during benchmark")
			}

			var loadedConfig map[string]interface{}
			err = json.Unmarshal(data, &loadedConfig)
			if err != nil {
				b.Error("Unmarshal failed during benchmark")
			}
		}
	})
}
