package bootstrap

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"istio.io/istio/pkg/env"
	"istio.io/istio/pkg/model"
)

// MockWriter is a mock implementation of io.Writer for testing.
type MockWriter struct {
	Buffer bytes.Buffer
}

func (m *MockWriter) Write(p []byte) (int, error) {
	return m.Buffer.Write(p)
}

// TestNewInstance tests the New function.
func TestNewInstance(t *testing.T) {
	t.Run("CreateInstanceSuccessfully", func(t *testing.T) {
		cfg := Config{
			// Initialize with necessary fields.
			Metadata: Metadata{
				ProxyConfig: model.NodeMetaProxyConfig{
					ConfigPath: "/tmp/config",
				},
			},
		}
		instance := New(cfg)
		if instance == nil {
			t.Fatal("Expected new Instance, got nil")
		}
	})
}

// TestWriteTo tests the WriteTo method of the instance.
func TestWriteTo(t *testing.T) {
	tests := []struct {
		name          string
		templateFile  string
		cfg           Config
		expectedError error
		setupFunc     func()
	}{
		{
			name:         "SuccessfulWrite",
			templateFile: "testdata/valid_template.json",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			expectedError: nil,
		},
		{
			name:         "TemplateFileNotFound",
			templateFile: "non_existent_template.json",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			expectedError: os.ErrNotExist,
		},
		{
			name:         "InvalidTemplateSyntax",
			templateFile: "testdata/invalid_template.json",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			expectedError: errors.New("template parsing error"),
			setupFunc: func() {
				// Setup invalid template content if necessary
			},
		},
		{
			name:         "TemplateExecutionError",
			templateFile: "testdata/template_with_error.json",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			expectedError: errors.New("execution error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			instance := New(tt.cfg)
			mockWriter := &MockWriter{}
			err := instance.WriteTo(tt.templateFile, mockWriter)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("Expected error '%v', got nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Fatalf("Expected error containing '%v', got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if mockWriter.Buffer.Len() == 0 {
					t.Fatal("Expected data to be written, got empty buffer")
				}
			}
		})
	}
}

// TestCreateFile tests the CreateFile method of the instance.
func TestCreateFile(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name          string
		cfg           Config
		overrideEnv   string
		expectedPath  string
		expectedError error
		setupFunc     func()
	}{
		{
			name: "CreateFileSuccessfullyWithDefaultTemplate",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			expectedPath:  "/tmp/config/envoy-rev.json",
			expectedError: nil,
		},
		{
			name: "CreateFileWithCustomConfigFile",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath:        "/tmp/config",
						CustomConfigFile: "custom_template.json",
					},
				},
			},
			expectedPath:  "/tmp/config/envoy-rev.json",
			expectedError: nil,
		},
		{
			name: "CreateFileWithEnvironmentOverride",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			overrideEnv:   "override_template.json",
			expectedPath:  "/tmp/config/envoy-rev.json",
			expectedError: nil,
		},
		{
			name: "CreateFileMkdirAllError",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/invalid/path/config",
					},
				},
			},
			expectedPath:  "",
			expectedError: errors.New("mkdir"),
		},
		{
			name: "CreateFileOsCreateError",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			expectedPath:  "",
			expectedError: errors.New("create file"),
			setupFunc: func() {
				// Mock os.Create to return an error
				originalCreate := osCreate
				osCreate = func(name string) (*os.File, error) {
					return nil, errors.New("create file error")
				}
				t.Cleanup(func() { osCreate = originalCreate })
			},
		},
		{
			name: "CreateFileWriteToError",
			cfg: Config{
				Metadata: Metadata{
					ProxyConfig: model.NodeMetaProxyConfig{
						ConfigPath: "/tmp/config",
					},
				},
			},
			expectedPath:  "",
			expectedError: errors.New("write to error"),
			setupFunc: func() {
				// Mock instance.WriteTo to return an error
				originalWriteTo := instanceWriteTo
				instanceWriteTo = func(templateFile string, w io.Writer) error {
					return errors.New("write to error")
				}
				t.Cleanup(func() { instanceWriteTo = originalWriteTo })
			},
		},
	}

	// Override environment variable if needed
	originalOverrideVar := overrideVar
	defer func() { overrideVar = originalOverrideVar }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.overrideEnv != "" {
				env.Set("ISTIO_BOOTSTRAP", tt.overrideEnv)
				t.Cleanup(func() { env.Set("ISTIO_BOOTSTRAP", "") })
			} else {
				env.Set("ISTIO_BOOTSTRAP", "")
			}

			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			cfg := tt.cfg
			// Replace ConfigPath with tempDir if needed
			if strings.Contains(cfg.Metadata.ProxyConfig.ConfigPath, "/tmp/config") {
				cfg.Metadata.ProxyConfig.ConfigPath = filepath.Join(tempDir, "config")
			}

			instance := New(cfg)
			outputPath, err := instance.CreateFile()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("Expected error containing '%v', got nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Fatalf("Expected error containing '%v', got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if outputPath != tt.expectedPath && !strings.HasSuffix(outputPath, "envoy-rev.json") {
					t.Fatalf("Expected output path '%s', got '%s'", tt.expectedPath, outputPath)
				}
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Fatalf("Expected output file '%s' to exist", outputPath)
				}
			}
		})
	}
}

// TestGetEffectiveTemplatePath tests the GetEffectiveTemplatePath function.
func TestGetEffectiveTemplatePath(t *testing.T) {
	tests := []struct {
		name         string
		pc           *model.NodeMetaProxyConfig
		envOverride  string
		expectedPath string
	}{
		{
			name: "CustomConfigFileSet",
			pc: &model.NodeMetaProxyConfig{
				CustomConfigFile: "custom_template.json",
			},
			envOverride:  "",
			expectedPath: "custom_template.json",
		},
		{
			name: "ProxyBootstrapTemplatePathSet",
			pc: &model.NodeMetaProxyConfig{
				ProxyBootstrapTemplatePath: "bootstrap_template.json",
			},
			envOverride:  "",
			expectedPath: "bootstrap_template.json",
		},
		{
			name: "DefaultConfigPath",
			pc: &model.NodeMetaProxyConfig{
				ConfigPath: "",
			},
			envOverride:  "",
			expectedPath: DefaultCfgDir,
		},
		{
			name: "EnvironmentOverrideSet",
			pc: &model.NodeMetaProxyConfig{
				CustomConfigFile:         "custom_template.json",
				ProxyBootstrapTemplatePath: "bootstrap_template.json",
			},
			envOverride:  "env_override.json",
			expectedPath: "env_override.json",
		},
	}

	originalOverrideVar := overrideVar
	defer func() { overrideVar = originalOverrideVar }()

	for _, tt := range tests {
		if tt.envOverride != "" {
			env.Set("ISTIO_BOOTSTRAP", tt.envOverride)
			tt.expectedPath = tt.envOverride
		} else {
			env.Set("ISTIO_BOOTSTRAP", "")
		}

		t.Run(tt.name, func(t *testing.T) {
			path := GetEffectiveTemplatePath(tt.pc)
			if path != tt.expectedPath {
				t.Errorf("Expected path '%s', got '%s'", tt.expectedPath, path)
			}
		})
	}
}

// TestConfigFile tests the configFile function.
func TestConfigFile(t *testing.T) {
	tests := []struct {
		name          string
		config        string
		templateFile  string
		expectedSuffix string
		expectedPath  string
	}{
		{
			name:          "JSONTemplate",
			config:        "/tmp/config",
			templateFile:  "template.json",
			expectedSuffix: "json",
			expectedPath:  "/tmp/config/envoy-rev.json",
		},
		{
			name:          "YAMLTmplTemplate",
			config:        "/tmp/config",
			templateFile:  "template.yaml.tmpl",
			expectedSuffix: "yaml",
			expectedPath:  "/tmp/config/envoy-rev.yaml",
		},
		{
			name:          "YAMLTemplate",
			config:        "/tmp/config",
			templateFile:  "template.yaml",
			expectedSuffix: "yaml",
			expectedPath:  "/tmp/config/envoy-rev.yaml",
		},
		{
			name:          "UnknownSuffixTemplate",
			config:        "/tmp/config",
			templateFile:  "template.txt",
			expectedSuffix: "json",
			expectedPath:  "/tmp/config/envoy-rev.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := configFile(tt.config, tt.templateFile)
			if path != tt.expectedPath {
				t.Errorf("Expected path '%s', got '%s'", tt.expectedPath, path)
			}
		})
	}
}

// TestNewTemplate tests the newTemplate function.
func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name          string
		templatePath  string
		templateContent string
		expectedError bool
	}{
		{
			name:          "ValidTemplate",
			templatePath:  "testdata/valid_template.json",
			templateContent: `{{ .Key }}`,
			expectedError: false,
		},
		{
			name:          "InvalidTemplatePath",
			templatePath:  "non_existent_template.json",
			templateContent: "",
			expectedError: true,
		},
		{
			name:          "InvalidTemplateSyntax",
			templatePath:  "testdata/invalid_syntax_template.json",
			templateContent: `{{ .Key `,
			expectedError: true,
		},
		{
			name:          "EmptyTemplate",
			templatePath:  "testdata/empty_template.json",
			templateContent: ``,
			expectedError: false,
		},
	}

	// Setup test data
	for _, tt := range tests {
		if tt.templateContent != "" {
			err := os.MkdirAll(filepath.Dir(tt.templatePath), 0755)
			if err != nil {
				t.Fatalf("Failed to create directory for template: %v", err)
			}
			err = os.WriteFile(tt.templatePath, []byte(tt.templateContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write template file: %v", err)
			}
			t.Cleanup(func() {
				os.Remove(tt.templatePath)
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := newTemplate(tt.templatePath)
			if tt.expectedError {
				if err == nil {
					t.Fatalf("Expected error for template '%s', got nil", tt.templatePath)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for template '%s': %v", tt.templatePath, err)
				}
				if tmpl == nil {
					t.Fatal("Expected a non-nil template")
				}
			}
		})
	}
}

// TestToJSON tests the toJSON function.
func TestToJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		expectedJSON string
	}{
		{
			name:         "NilInput",
			input:        nil,
			expectedJSON: "{}",
		},
		{
			name:         "ValidInput",
			input:        map[string]string{"key": "value"},
			expectedJSON: `{"key":"value"}`,
		},
		{
			name:         "MarshalError",
			input:        func() {}, // Functions cannot be marshaled to JSON
			expectedJSON: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toJSON(tt.input)
			if result != tt.expectedJSON {
				t.Errorf("Expected JSON '%s', got '%s'", tt.expectedJSON, result)
			}
		})
	}
}