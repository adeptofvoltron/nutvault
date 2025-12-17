package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// Project represents a vault project that stores encrypted variables.
// It contains the project directory path, encryption key, and loaded variables.
type Project struct {
	projectDir string
	key        []byte
	variables  map[string]string
}

// generateDefaultKey generates a deterministic key from current user and host information.
// It uses UID, username, home directory, and machine-id/hostname to create a unique key.
func generateDefaultKey() ([]byte, error) {
	var keyParts []string

	// Get current user information
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	keyParts = append(keyParts, currentUser.Uid)
	keyParts = append(keyParts, currentUser.Username)
	keyParts = append(keyParts, currentUser.HomeDir)

	// Get machine identifier
	var machineID string
	if runtime.GOOS == "linux" {
		// Try to read machine-id on Linux
		if data, err := os.ReadFile("/etc/machine-id"); err == nil {
			machineID = strings.TrimSpace(string(data))
		} else {
			// Fallback to hostname
			if hostname, err := os.Hostname(); err == nil {
				machineID = hostname
			}
		}
	} else {
		// For other OS, use hostname
		if hostname, err := os.Hostname(); err == nil {
			machineID = hostname
		}
	}
	keyParts = append(keyParts, machineID)

	// Combine all parts and hash
	combined := strings.Join(keyParts, ":")
	hash := sha256.Sum256([]byte(combined))
	return hash[:], nil
}

// calculateProjectHash calculates the SHA256 hash of projectName + key,
// truncated to 16 hex characters.
func calculateProjectHash(projectName string, key []byte) string {
	data := append([]byte(projectName), key...)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])[:16]
}

// NewProject creates a new vault project or opens an existing one.
// If key is nil, a default key is generated deterministically from user and host info.
// The project directory is created at ~/.nutvault/projects/<projectName>.<hash>
//
// Example:
//   project, err := vault.NewProject("myproject", nil)
//   if err != nil {
//       log.Fatal(err)
//   }
func NewProject(projectName string, key []byte) (*Project, error) {
	// Use default key if none provided
	if key == nil {
		var err error
		key, err = generateDefaultKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate default key: %w", err)
		}
	}

	// Calculate project hash
	hash := calculateProjectHash(projectName, key)

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create project directory path
	projectDir := filepath.Join(homeDir, ".nutvault", "projects", fmt.Sprintf("%s.%s", projectName, hash))

	// Create project directory if it doesn't exist
	if err := os.MkdirAll(projectDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Initialize project
	project := &Project{
		projectDir: projectDir,
		key:        key,
		variables:  make(map[string]string),
	}

	// Try to load existing variables
	dataPath := filepath.Join(projectDir, "data.json")
	if _, err := os.Stat(dataPath); err == nil {
		if err := project.loadVariables(); err != nil {
			return nil, fmt.Errorf("failed to load existing variables: %w", err)
		}
	}

	return project, nil
}

// encryptData encrypts data using AES-GCM with the project key.
func (p *Project) encryptData(plaintext []byte) ([]byte, error) {
	// Use the project key directly (it's already 32 bytes from SHA256)
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptData decrypts data using AES-GCM with the project key.
func (p *Project) decryptData(ciphertext []byte) ([]byte, error) {
	// Use the project key directly (it's already 32 bytes from SHA256)
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce (first gcm.NonceSize() bytes)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// loadVariables loads and decrypts variables from the data.json file in the project directory.
func (p *Project) loadVariables() error {
	dataPath := filepath.Join(p.projectDir, "data.json")
	encryptedData, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("failed to read data.json: %w", err)
	}

	// Decrypt the data
	decryptedData, err := p.decryptData(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	if err := json.Unmarshal(decryptedData, &p.variables); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// saveVariables encrypts and saves variables to the data.json file in the project directory.
func (p *Project) saveVariables() error {
	dataPath := filepath.Join(p.projectDir, "data.json")
	
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(p.variables, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Encrypt the data
	encryptedData, err := p.encryptData(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Write encrypted data to file
	if err := os.WriteFile(dataPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write data.json: %w", err)
	}

	return nil
}

// SaveVariable saves a variable to the project vault.
// The variable is stored in memory and persisted to data.json file.
//
// Example:
//   err := project.SaveVariable("API_KEY", "secret123")
//   if err != nil {
//       log.Fatal(err)
//   }
func (p *Project) SaveVariable(key string, value string) error {
	if p.variables == nil {
		p.variables = make(map[string]string)
	}

	p.variables[key] = value

	if err := p.saveVariables(); err != nil {
		return fmt.Errorf("failed to save variable: %w", err)
	}

	return nil
}

// GetVariable retrieves a variable from the project vault.
// Returns the value, a boolean indicating if the key exists, and any error.
//
// Example:
//   value, exists, err := project.GetVariable("API_KEY")
//   if err != nil {
//       log.Fatal(err)
//   }
//   if exists {
//       fmt.Println("Value:", value)
//   }
func (p *Project) GetVariable(key string) (string, bool, error) {
	if p.variables == nil {
		// Try to load variables if not loaded
		if err := p.loadVariables(); err != nil {
			// If file doesn't exist, return empty map
			if os.IsNotExist(err) {
				p.variables = make(map[string]string)
				return "", false, nil
			}
			return "", false, fmt.Errorf("failed to load variables: %w", err)
		}
	}

	value, exists := p.variables[key]
	return value, exists, nil
}

// ClearProject removes the entire project directory and all its contents.
// This operation cannot be undone.
//
// Example:
//   err := project.ClearProject()
//   if err != nil {
//       log.Fatal(err)
//   }
func (p *Project) ClearProject() error {
	if p.projectDir == "" {
		return fmt.Errorf("project directory not set")
	}

	if err := os.RemoveAll(p.projectDir); err != nil {
		return fmt.Errorf("failed to remove project directory: %w", err)
	}

	// Clear in-memory variables
	p.variables = make(map[string]string)

	return nil
}

// GetProjectDir returns the project directory path.
func (p *Project) GetProjectDir() string {
	return p.projectDir
}

// GetAllVariables returns a map of all variables in the vault project.
func (p *Project) GetAllVariables() map[string]string {
	if p.variables == nil {
		// Try to load variables if not loaded
		if err := p.loadVariables(); err != nil {
			// If file doesn't exist, return empty map
			if os.IsNotExist(err) {
				p.variables = make(map[string]string)
				return p.variables
			}
			// For other errors, return empty map (caller should handle error)
			return make(map[string]string)
		}
	}

	// Return a copy to prevent external modifications
	result := make(map[string]string)
	for k, v := range p.variables {
		result[k] = v
	}
	return result
}

