package environment

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestLoadEnv(t *testing.T) {
	content := `TEST_KEY=test_value
ANOTHER_KEY=another_value`

	tmpFile, err := os.CreateTemp("", "*.env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	// nolint:govet // ignore
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	envVars, err := loadEnv(tmpFile.Name())
	if err != nil {
		t.Fatalf("loadEnv failed: %v", err)
	}

	expected := map[string]string{
		"TEST_KEY":    "test_value",
		"ANOTHER_KEY": "another_value",
	}
	if !reflect.DeepEqual(envVars, expected) {
		t.Errorf("Expected %v, got %v", expected, envVars)
	}
}

func TestProcessValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"quoted value"`, "quoted value"},
		{`'another quoted value'`, "another quoted value"},
		{`escaped\nvalue`, "escaped\nvalue"},
		{`escaped\tvalue`, "escaped\tvalue"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := processValue(test.input)
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	envVars := map[string]string{
		"EXISTING_VAR": "existing_value",
	}
	os.Setenv("EXISTING_ENV_VAR", "existing_env_value")

	tests := []struct {
		input    string
		expected string
	}{
		{"${EXISTING_VAR}", "existing_value"},
		{"${EXISTING_ENV_VAR}", "existing_env_value"},
		{"${NON_EXISTENT_VAR}", "${NON_EXISTENT_VAR}"},
		{"no vars", "no vars"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := expandEnvVars(test.input, envVars)
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestParseEnv(t *testing.T) {
	type Config struct {
		TestKey    string        `env:"TEST_KEY"`
		AnotherKey int           `env:"ANOTHER_KEY"`
		Duration   time.Duration `env:"DURATION"`
		Optional   string        `env:"OPTIONAL" default:"default_value"`
		Required   string        `env:"REQUIRED" required:"true"`
	}

	envVars := map[string]string{
		"TEST_KEY":    "test_value",
		"ANOTHER_KEY": "42",
		"DURATION":    "1h",
		"REQUIRED":    "required_value",
	}

	var cfg Config
	if err := parseEnv(&cfg, envVars); err != nil {
		t.Fatalf("parseEnv failed: %v", err)
	}

	if cfg.TestKey != "test_value" {
		t.Errorf("Expected TestKey to be 'test_value', got %q", cfg.TestKey)
	}
	if cfg.AnotherKey != 42 {
		t.Errorf("Expected AnotherKey to be 42, got %d", cfg.AnotherKey)
	}
	if cfg.Duration != time.Hour {
		t.Errorf("Expected Duration to be 1h, got %v", cfg.Duration)
	}
	if cfg.Optional != "default_value" {
		t.Errorf("Expected Optional to be 'default_value', got %q", cfg.Optional)
	}
	if cfg.Required != "required_value" {
		t.Errorf("Expected Required to be 'required_value', got %q", cfg.Required)
	}
}

func TestSetValue(t *testing.T) {
	tests := []struct {
		name     string
		field    interface{}
		value    string
		expected interface{}
	}{
		{"string", new(string), "test", "test"},
		{"int", new(int), "42", 42},
		{"duration", new(time.Duration), "1h", time.Hour},
		{"bool", new(bool), "true", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := reflect.ValueOf(test.field).Elem()
			if err := setValue(field, test.value); err != nil {
				t.Errorf("setValue failed: %v", err)
			}
			if !reflect.DeepEqual(field.Interface(), test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, field.Interface())
			}
		})
	}
}
