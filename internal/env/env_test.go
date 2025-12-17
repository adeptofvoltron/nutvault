package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadEnvFile(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	
	content := `# This is a comment
API_KEY=secret123
DB_HOST=localhost
DB_PORT=5432

# Another comment
EMPTY_VAR=
QUOTED_VAR="value with spaces"
SINGLE_QUOTED='single quoted value'
`
	
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Read the file
	env, err := ReadEnvFile(envFile)
	if err != nil {
		t.Fatalf("ReadEnvFile failed: %v", err)
	}

	// Check file path
	if env.FilePath != envFile {
		t.Errorf("Expected FilePath %s, got %s", envFile, env.FilePath)
	}

	// Check variables
	expectedVars := map[string]string{
		"API_KEY":       "secret123",
		"DB_HOST":       "localhost",
		"DB_PORT":       "5432",
		"EMPTY_VAR":     "",
		"QUOTED_VAR":    "value with spaces",
		"SINGLE_QUOTED": "single quoted value",
	}

	for key, expectedValue := range expectedVars {
		value, exists := env.GetVariable(key)
		if !exists {
			t.Errorf("Variable %s not found", key)
			continue
		}
		if value != expectedValue {
			t.Errorf("Variable %s: expected %q, got %q", key, expectedValue, value)
		}
	}

	// Check line count (should have at least 8 lines: 2 comments, 6 variables, 1 empty, possibly trailing newline)
	if len(env.Lines) < 8 {
		t.Errorf("Expected at least 8 lines, got %d", len(env.Lines))
	}
}

func TestReadEnvFile_NonExistent(t *testing.T) {
	_, err := ReadEnvFile("/nonexistent/file.env")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name     string
		rawLine  string
		expected EnvLine
	}{
		{
			name:    "empty line",
			rawLine: "",
			expected: EnvLine{
				Type:    LineEmpty,
				RawLine: "",
			},
		},
		{
			name:    "comment line",
			rawLine: "# This is a comment",
			expected: EnvLine{
				Type:    LineComment,
				Comment: "# This is a comment",
				RawLine: "# This is a comment",
			},
		},
		{
			name:    "variable line",
			rawLine: "KEY=value",
			expected: EnvLine{
				Type:    LineVariable,
				Key:     "KEY",
				Value:   "value",
				RawLine: "KEY=value",
			},
		},
		{
			name:    "variable with spaces",
			rawLine: "  KEY  =  value  ",
			expected: EnvLine{
				Type:    LineVariable,
				Key:     "KEY",
				Value:   "value",
				RawLine: "  KEY  =  value  ",
			},
		},
		{
			name:    "quoted value",
			rawLine: `KEY="quoted value"`,
			expected: EnvLine{
				Type:    LineVariable,
				Key:     "KEY",
				Value:   "quoted value",
				RawLine: `KEY="quoted value"`,
			},
		},
		{
			name:    "single quoted value",
			rawLine: `KEY='single quoted'`,
			expected: EnvLine{
				Type:    LineVariable,
				Key:     "KEY",
				Value:   "single quoted",
				RawLine: `KEY='single quoted'`,
			},
		},
		{
			name:    "invalid format",
			rawLine: "not a valid line",
			expected: EnvLine{
				Type:    LineComment,
				Comment: "not a valid line",
				RawLine: "not a valid line",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLine(tt.rawLine, 0)
			if result.Type != tt.expected.Type {
				t.Errorf("Type: expected %v, got %v", tt.expected.Type, result.Type)
			}
			if result.Key != tt.expected.Key {
				t.Errorf("Key: expected %q, got %q", tt.expected.Key, result.Key)
			}
			if result.Value != tt.expected.Value {
				t.Errorf("Value: expected %q, got %q", tt.expected.Value, result.Value)
			}
		})
	}
}

func TestUnquoteValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no quotes", "value", "value"},
		{"double quotes", `"value"`, "value"},
		{"single quotes", `'value'`, "value"},
		{"unmatched quotes", `"value'`, `"value'`},
		{"short string", "a", "a"},
		{"empty", "", ""},
		{"quoted empty", `""`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unquoteValue(tt.input)
			if result != tt.expected {
				t.Errorf("unquoteValue(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQuoteValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple value", "value", "value"},
		{"with spaces", "value with spaces", `"value with spaces"`},
		{"empty", "", `""`},
		{"with quotes", `value"test`, `"value\"test"`},
		{"with dollar", "value$test", `"value$test"`},
		{"with tab", "value\ttest", `"value	test"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteValue(tt.input)
			if result != tt.expected {
				t.Errorf("quoteValue(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetVariable(t *testing.T) {
	env := &EnvFile{
		Variables: map[string]string{
			"KEY1": "value1",
			"KEY2": "value2",
		},
	}

	// Test existing variable
	value, exists := env.GetVariable("KEY1")
	if !exists {
		t.Error("Expected KEY1 to exist")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %q", value)
	}

	// Test non-existing variable
	_, exists = env.GetVariable("NONEXISTENT")
	if exists {
		t.Error("Expected NONEXISTENT to not exist")
	}
}

func TestSetVariable(t *testing.T) {
	env := &EnvFile{
		Lines:     []EnvLine{},
		Variables: make(map[string]string),
	}

	// Test adding new variable
	err := env.SetVariable("NEW_KEY", "new_value")
	if err != nil {
		t.Fatalf("SetVariable failed: %v", err)
	}

	value, exists := env.GetVariable("NEW_KEY")
	if !exists {
		t.Error("Expected NEW_KEY to exist after SetVariable")
	}
	if value != "new_value" {
		t.Errorf("Expected new_value, got %q", value)
	}

	// Test updating existing variable
	err = env.SetVariable("NEW_KEY", "updated_value")
	if err != nil {
		t.Fatalf("SetVariable failed: %v", err)
	}

	value, exists = env.GetVariable("NEW_KEY")
	if value != "updated_value" {
		t.Errorf("Expected updated_value, got %q", value)
	}

	// Test empty key
	err = env.SetVariable("", "value")
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestSetVariable_UpdateExisting(t *testing.T) {
	env := &EnvFile{
		Lines: []EnvLine{
			{Type: LineVariable, Key: "EXISTING", Value: "old_value", LineIndex: 0},
		},
		Variables: map[string]string{
			"EXISTING": "old_value",
		},
	}

	err := env.SetVariable("EXISTING", "new_value")
	if err != nil {
		t.Fatalf("SetVariable failed: %v", err)
	}

	if env.Lines[0].Value != "new_value" {
		t.Errorf("Expected line value to be updated to new_value, got %q", env.Lines[0].Value)
	}

	if env.Variables["EXISTING"] != "new_value" {
		t.Errorf("Expected variable map to be updated to new_value, got %q", env.Variables["EXISTING"])
	}
}

func TestFillEmptyVariables(t *testing.T) {
	env := &EnvFile{
		Lines: []EnvLine{
			{Type: LineVariable, Key: "FILLED", Value: "already_filled", LineIndex: 0},
			{Type: LineVariable, Key: "EMPTY", Value: "", LineIndex: 1},
			{Type: LineVariable, Key: "ANOTHER_EMPTY", Value: "   ", LineIndex: 2},
		},
		Variables: map[string]string{
			"FILLED":       "already_filled",
			"EMPTY":        "",
			"ANOTHER_EMPTY": "   ",
		},
	}

	values := map[string]string{
		"FILLED":       "should_not_change",
		"EMPTY":        "filled_value",
		"ANOTHER_EMPTY": "another_filled",
		"NOT_IN_FILE":  "should_not_be_added",
	}

	err := env.FillEmptyVariables(values)
	if err != nil {
		t.Fatalf("FillEmptyVariables failed: %v", err)
	}

	// Check that filled variable wasn't changed
	if env.Variables["FILLED"] != "already_filled" {
		t.Errorf("Expected FILLED to remain already_filled, got %q", env.Variables["FILLED"])
	}

	// Check that empty variables were filled
	if env.Variables["EMPTY"] != "filled_value" {
		t.Errorf("Expected EMPTY to be filled_value, got %q", env.Variables["EMPTY"])
	}

	if strings.TrimSpace(env.Variables["ANOTHER_EMPTY"]) == "" {
		t.Error("Expected ANOTHER_EMPTY to be filled")
	}

	// Check that variable not in file wasn't added
	if env.HasVariable("NOT_IN_FILE") {
		t.Error("Expected NOT_IN_FILE to not be added")
	}
}

func TestSaveEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	env := &EnvFile{
		FilePath: envFile,
		Lines: []EnvLine{
			{Type: LineComment, Comment: "# Comment", LineIndex: 0},
			{Type: LineVariable, Key: "KEY1", Value: "value1", LineIndex: 1},
			{Type: LineEmpty, LineIndex: 2},
			{Type: LineVariable, Key: "KEY2", Value: "value with spaces", LineIndex: 3},
		},
		Variables: map[string]string{
			"KEY1": "value1",
			"KEY2": "value with spaces",
		},
	}

	err := env.SaveEnvFile()
	if err != nil {
		t.Fatalf("SaveEnvFile failed: %v", err)
	}

	// Read back and verify
	loaded, err := ReadEnvFile(envFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if len(loaded.Lines) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(loaded.Lines))
	}

	value, exists := loaded.GetVariable("KEY1")
	if !exists || value != "value1" {
		t.Errorf("KEY1: expected value1, got %q (exists: %v)", value, exists)
	}

	value, exists = loaded.GetVariable("KEY2")
	if !exists || value != "value with spaces" {
		t.Errorf("KEY2: expected 'value with spaces', got %q (exists: %v)", value, exists)
	}
}

func TestSaveEnvFileTo(t *testing.T) {
	tmpDir := t.TempDir()
	originalFile := filepath.Join(tmpDir, "original.env")
	backupFile := filepath.Join(tmpDir, "backup.env")

	env := &EnvFile{
		FilePath: originalFile,
		Lines: []EnvLine{
			{Type: LineVariable, Key: "KEY", Value: "value", LineIndex: 0},
		},
		Variables: map[string]string{
			"KEY": "value",
		},
	}

	err := env.SaveEnvFileTo(backupFile)
	if err != nil {
		t.Fatalf("SaveEnvFileTo failed: %v", err)
	}

	// Verify backup file exists and has correct content
	loaded, err := ReadEnvFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	value, exists := loaded.GetVariable("KEY")
	if !exists || value != "value" {
		t.Errorf("Expected KEY=value in backup, got %q (exists: %v)", value, exists)
	}
}

func TestGetAllVariables(t *testing.T) {
	env := &EnvFile{
		Variables: map[string]string{
			"KEY1": "value1",
			"KEY2": "value2",
			"KEY3": "value3",
		},
	}

	all := env.GetAllVariables()
	if len(all) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(all))
	}

	// Verify it's a copy (modifying shouldn't affect original)
	all["NEW_KEY"] = "new_value"
	if env.HasVariable("NEW_KEY") {
		t.Error("Modifying returned map should not affect original")
	}
}

func TestHasVariable(t *testing.T) {
	env := &EnvFile{
		Variables: map[string]string{
			"EXISTS": "value",
		},
	}

	if !env.HasVariable("EXISTS") {
		t.Error("Expected EXISTS to be found")
	}

	if env.HasVariable("NOT_EXISTS") {
		t.Error("Expected NOT_EXISTS to not be found")
	}
}

func TestIsVariableEmpty(t *testing.T) {
	env := &EnvFile{
		Variables: map[string]string{
			"EMPTY":     "",
			"SPACES":    "   ",
			"NOT_EMPTY": "value",
		},
	}

	if !env.IsVariableEmpty("EMPTY") {
		t.Error("Expected EMPTY to be empty")
	}

	if !env.IsVariableEmpty("SPACES") {
		t.Error("Expected SPACES to be empty (only whitespace)")
	}

	if env.IsVariableEmpty("NOT_EMPTY") {
		t.Error("Expected NOT_EMPTY to not be empty")
	}

	if env.IsVariableEmpty("NOT_EXISTS") {
		t.Error("Expected NOT_EXISTS to not be empty (doesn't exist)")
	}
}

