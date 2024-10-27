package main

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	// CommandMetadata represents the complete metadata for a command
	CommandMetadata struct {
		Autocompletion bool `yaml:"autocompletion,omitempty"`
		ArgParser      bool
		Category       []string `yaml:"category,omitempty"`
		Help           string   `yaml:"help,omitempty"`
		Syntax         Syntax   `yaml:"syntax,omitempty"`
	}

	// Syntax defines the command's parameter syntax
	Syntax struct {
		Parameters []Parameter `yaml:"parameters,omitempty"`
		Groups     []Group     `yaml:"groups,omitempty"`
	}

	// Parameter represents a single command parameter
	Parameter struct {
		Name               string                 `yaml:"name"`
		Aliases            []string               `yaml:"aliases,omitempty"`
		Description        string                 `yaml:"desc,omitempty"`
		Positional         bool                   `yaml:"positional,omitempty"`
		Required           bool                   `yaml:"required,omitempty"`
		Placeholder        string                 `yaml:"placeholder,omitempty"`
		Type               string                 `yaml:"type"`
		Values             []string               `yaml:"values,omitempty"`
		Default            interface{}            `yaml:"default,omitempty"`
		NumValues          string                 `yaml:"num_values,omitempty"`
		Delimiter          string                 `yaml:"delimiter,omitempty"`
		Last               bool                   `yaml:"last,omitempty"`
		Leftovers          bool                   `yaml:"leftovers,omitempty"`
		AllowHyphenValues  bool                   `yaml:"allow_hyphen_values,omitempty"`
		Requires           []string               `yaml:"requires,omitempty"`
		ConflictsWith      []string               `yaml:"conflicts_with,omitempty"`
		RequiredWithout    []string               `yaml:"required_without,omitempty"`
		RequiredWithoutAll []string               `yaml:"required_without_all,omitempty"`
		RequiredIfEq       map[string]interface{} `yaml:"required_if_eq,omitempty"`
		RequiredIfEqAll    map[string]interface{} `yaml:"required_if_eq_all,omitempty"`
	}

	// Group represents a group of parameters
	Group struct {
		Name          string   `yaml:"name"`
		Parameters    []string `yaml:"parameters"`
		Required      bool     `yaml:"required"`
		Multiple      bool     `yaml:"multiple"`
		Requires      []string `yaml:"requires"`
		ConflictsWith []string `yaml:"conflicts_with"`
	}
)

// WriteToFile writes the metadata to a YAML file
func (m *CommandMetadata) WriteToFile(filename string) error {
	var data bytes.Buffer

	// Configure the encoder
	yamlEncoder := yaml.NewEncoder(&data)
	yamlEncoder.SetIndent(2)

	// Encode the metadata
	err := yamlEncoder.Encode(m)
	if err != nil {
		return err
	}

	// Write the data to the file
	return os.WriteFile(filename, data.Bytes(), 0644)
}
