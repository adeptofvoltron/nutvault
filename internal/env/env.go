package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// EnvLine represents a single line in an .env file.
// It can be a variable assignment, a comment, or an empty line.
type EnvLine struct {
	Type      LineType
	Key       string
	Value     string
	Comment   string
	RawLine   string
	LineIndex int
}

// LineType represents the type of line in an .env file.
type LineType int

const (
	LineVariable LineType = iota // KEY=value
	LineComment                  // # comment
	LineEmpty                    // empty line
)

// EnvFile represents a parsed .env file with all its lines preserved.
type EnvFile struct {
	FilePath string
	Lines    []EnvLine
	Variables map[string]string // Quick lookup map
}

// ReadEnvFile reads and parses an .env file from the given path.
// Returns an EnvFile struct with all lines preserved (including comments and empty lines).
//
// Example:
//   envFile, err := env.ReadEnvFile(".env")
//   if err != nil {
//       log.Fatal(err)
//   }
func ReadEnvFile(filePath string) (*EnvFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	envFile := &EnvFile{
		FilePath:  filePath,
		Lines:    make([]EnvLine, 0),
		Variables: make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	lineIndex := 0

	for scanner.Scan() {
		rawLine := scanner.Text()
		line := parseLine(rawLine, lineIndex)
		envFile.Lines = append(envFile.Lines, line)

		// Add to variables map if it's a variable line
		if line.Type == LineVariable {
			envFile.Variables[line.Key] = line.Value
		}

		lineIndex++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read .env file: %w", err)
	}

	return envFile, nil
}

// parseLine parses a single line from an .env file.
func parseLine(rawLine string, lineIndex int) EnvLine {
	trimmed := strings.TrimSpace(rawLine)

	// Empty line
	if trimmed == "" {
		return EnvLine{
			Type:      LineEmpty,
			RawLine:   rawLine,
			LineIndex: lineIndex,
		}
	}

	// Comment line
	if strings.HasPrefix(trimmed, "#") {
		return EnvLine{
			Type:      LineComment,
			Comment:   trimmed,
			RawLine:   rawLine,
			LineIndex: lineIndex,
		}
	}

	// Variable line: KEY=value or KEY="value" or KEY='value'
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) != 2 {
		// Invalid format, treat as comment
		return EnvLine{
			Type:      LineComment,
			Comment:   trimmed,
			RawLine:   rawLine,
			LineIndex: lineIndex,
		}
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Remove quotes if present
	value = unquoteValue(value)

	return EnvLine{
		Type:      LineVariable,
		Key:       key,
		Value:     value,
		RawLine:   rawLine,
		LineIndex: lineIndex,
	}
}

// unquoteValue removes surrounding quotes from a value if present.
func unquoteValue(value string) string {
	if len(value) < 2 {
		return value
	}

	// Check for double quotes
	if (value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'') {
		return value[1 : len(value)-1]
	}

	return value
}

// quoteValue adds quotes to a value if it contains spaces or special characters.
func quoteValue(value string) string {
	// If value contains spaces, quotes, or special characters, wrap in quotes
	if strings.ContainsAny(value, " \t\"'$") || value == "" {
		// Escape any existing quotes
		escaped := strings.ReplaceAll(value, "\"", "\\\"")
		return fmt.Sprintf(`"%s"`, escaped)
	}
	return value
}

// GetVariable retrieves a variable value from the .env file.
// Returns the value and a boolean indicating if the variable exists.
//
// Example:
//   value, exists := envFile.GetVariable("API_KEY")
//   if exists {
//       fmt.Println("Value:", value)
//   }
func (e *EnvFile) GetVariable(key string) (string, bool) {
	value, exists := e.Variables[key]
	return value, exists
}

// SetVariable sets or updates a variable in the .env file.
// If the variable already exists, its value is overwritten.
// If the variable doesn't exist, it's added at the end of the file.
//
// Example:
//   err := envFile.SetVariable("API_KEY", "new-secret-value")
//   if err != nil {
//       log.Fatal(err)
//   }
func (e *EnvFile) SetVariable(key string, value string) error {
	// Normalize key (remove spaces)
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	// Check if variable already exists
	for i := range e.Lines {
		if e.Lines[i].Type == LineVariable && e.Lines[i].Key == key {
			// Update existing variable
			e.Lines[i].Value = value
			e.Variables[key] = value
			return nil
		}
	}

	// Variable doesn't exist, add it at the end
	newLine := EnvLine{
		Type:      LineVariable,
		Key:       key,
		Value:     value,
		LineIndex: len(e.Lines),
	}
	e.Lines = append(e.Lines, newLine)
	e.Variables[key] = value

	return nil
}

// FillEmptyVariables fills empty variable values with provided values.
// Only variables that are already defined but have empty values are filled.
// Variables that don't exist in the file are not added.
//
// Example:
//   values := map[string]string{
//       "API_KEY": "secret123",
//       "DB_HOST": "localhost",
//   }
//   err := envFile.FillEmptyVariables(values)
//   if err != nil {
//       log.Fatal(err)
//   }
func (e *EnvFile) FillEmptyVariables(values map[string]string) error {
	for i := range e.Lines {
		if e.Lines[i].Type == LineVariable {
			key := e.Lines[i].Key
			currentValue := strings.TrimSpace(e.Lines[i].Value)

			// If variable is empty and we have a value to fill
			if currentValue == "" {
				if newValue, exists := values[key]; exists && newValue != "" {
					e.Lines[i].Value = newValue
					e.Variables[key] = newValue
				}
			}
		}
	}

	return nil
}

// SaveEnvFile saves the .env file to disk, preserving comments and formatting.
//
// Example:
//   err := envFile.SaveEnvFile()
//   if err != nil {
//       log.Fatal(err)
//   }
func (e *EnvFile) SaveEnvFile() error {
	return e.SaveEnvFileTo(e.FilePath)
}

// SaveEnvFileTo saves the .env file to a specific path.
//
// Example:
//   err := envFile.SaveEnvFileTo(".env.backup")
//   if err != nil {
//       log.Fatal(err)
//   }
func (e *EnvFile) SaveEnvFileTo(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, line := range e.Lines {
		var lineToWrite string

		switch line.Type {
		case LineEmpty:
			lineToWrite = ""
		case LineComment:
			lineToWrite = line.Comment
		case LineVariable:
			// Format: KEY=value (with quotes if needed)
			quotedValue := quoteValue(line.Value)
			lineToWrite = fmt.Sprintf("%s=%s", line.Key, quotedValue)
		}

		if _, err := writer.WriteString(lineToWrite + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	return nil
}

// GetAllVariables returns a map of all variables in the .env file.
func (e *EnvFile) GetAllVariables() map[string]string {
	result := make(map[string]string)
	for k, v := range e.Variables {
		result[k] = v
	}
	return result
}

// HasVariable checks if a variable exists in the .env file.
func (e *EnvFile) HasVariable(key string) bool {
	_, exists := e.Variables[key]
	return exists
}

// IsVariableEmpty checks if a variable exists and has an empty value.
func (e *EnvFile) IsVariableEmpty(key string) bool {
	value, exists := e.Variables[key]
	return exists && strings.TrimSpace(value) == ""
}

