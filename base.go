package environment

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	envVarRegex      = regexp.MustCompile(`\${([a-zA-Z_][a-zA-Z0-9_]*)}`)
	validEnvVarRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

func fillSpecification[T any](instance *T, paths ...string) error {
	envVars := make(map[string]string, len(paths))
	for _, path := range paths {
		fileVars, err := loadEnv(path)
		if err != nil {
			return fmt.Errorf("error loading .env file: %v", err)
		}
		for k, v := range fileVars {
			envVars[k] = v
		}
	}

	if err := parseEnv(instance, envVars); err != nil {
		return fmt.Errorf("field load environment: %v", err)
	}
	return nil
}

func loadEnv(filename string) (map[string]string, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s does not exist", filename)
		}
		return nil, err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatalf("failed to close env file: %v", err)
		}
	}(file)

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	var buffer bytes.Buffer
	var multiline bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasSuffix(line, "\\") {
			multiline = true
			buffer.WriteString(strings.TrimSuffix(line, "\\"))
			continue
		}

		if multiline {
			buffer.WriteString(line)
			line = buffer.String()
			buffer.Reset()
			multiline = false
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if !validEnvVarRegex.MatchString(key) {
			return nil, fmt.Errorf("invalid environment variable name: %s", key)
		}

		value := processValue(strings.TrimSpace(parts[1]))
		value = expandEnvVars(value, envVars)
		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}

func processValue(value string) string {
	if value == "" {
		return value
	}

	value = strings.NewReplacer(
		`\n`, "\n",
		`\t`, "\t",
		`\r`, "\r",
		`\\`, `\`,
	).Replace(value)

	if quote := value[0]; (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
		value = value[1 : len(value)-1]
	}

	return value
}

func expandEnvVars(value string, envVars map[string]string) string {
	return envVarRegex.ReplaceAllStringFunc(value, func(match string) string {
		varName := match[2 : len(match)-1]
		if val, exists := envVars[varName]; exists {
			return val
		}
		if val, exists := os.LookupEnv(varName); exists {
			return val
		}
		return match
	})
}

func parseEnv(cfg interface{}, envVars map[string]string) error {
	val := reflect.ValueOf(cfg).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		if field.Kind() == reflect.Struct {
			if err := parseEnv(field.Addr().Interface(), envVars); err != nil {
				return err
			}
			if customParser, ok := field.Addr().Interface().(CustomParser); ok {
				envValue, err := getValueFromEnvOrFile(structField, envVars)
				if err != nil {
					return err
				}
				if err := customParser.ParseEnv(envValue); err != nil {
					return err
				}
			}
			continue
		}

		envValue, err := getValueFromEnvOrFile(structField, envVars)
		if err != nil {
			return err
		}

		if envValue == "" {
			continue
		}

		if err := setValue(field, envValue); err != nil {
			return fmt.Errorf("error setting field %s: %v", structField.Name, err)
		}
	}
	return nil
}

func getValueFromEnvOrFile(structField reflect.StructField, envVars map[string]string) (string, error) {
	envTag := structField.Tag.Get("env")
	if envTag == "" {
		return "", nil
	}

	if val, exists := envVars[envTag]; exists {
		return val, nil
	}
	if val := os.Getenv(envTag); val != "" {
		return val, nil
	}
	if structField.Tag.Get("required") == "true" {
		return "", fmt.Errorf("required environment variable %s is missing", envTag)
	}
	return structField.Tag.Get("default"), nil
}

func setValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		}
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	case reflect.Slice:
		elements := strings.Split(value, ",")
		slice := reflect.MakeSlice(field.Type(), len(elements), len(elements))
		for i, elem := range elements {
			elem = strings.TrimSpace(elem)
			if err := setValue(slice.Index(i), elem); err != nil {
				return err
			}
		}
		field.Set(slice)
	case reflect.Map:
		var m map[string]string
		if err := json.Unmarshal([]byte(value), &m); err != nil {
			return err
		}
		field.Set(reflect.ValueOf(m))
	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}
	return nil
}

type CustomParser interface {
	ParseEnv(value string) error
}
