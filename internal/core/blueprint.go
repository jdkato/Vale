package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/tomwright/dasel/v2"
	"gopkg.in/yaml.v2"
)

type DaselValue = map[string]any

var blueprintEngines = []string{"tree-sitter", "dasel", "command"}

// A Query is a single query that we want to run against a document.
//
// The result of the query is optionally assigned the given scope.
type Query struct {
	Scope     string `yaml:"scope"`
	Operation string `yaml:"operation"`
}

// A Blueprint is a set of queries that we want to run against a document.
//
// The supported engines are:
//
// - `tree-sitter`
// - `dasel`
// - `command`
type Blueprint struct {
	Engine string  `yaml:"engine"`
	Steps  []Query `yaml:"steps"`
}

// A ScopedValues is a value that has been assigned a scope.
type ScopedValues struct {
	Scope  string
	Values []string
}

// NewBlueprint creates a new blueprint from the given path.
func NewBlueprint(path string) (*Blueprint, error) {
	var blueprint Blueprint

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &blueprint)
	if err != nil {
		return nil, err
	}

	if blueprint.Engine == "" {
		return nil, fmt.Errorf("missing parser")
	} else if !StringInSlice(blueprint.Engine, blueprintEngines) {
		return nil, fmt.Errorf("unsupported parser: %s", blueprint.Engine)
	}

	if len(blueprint.Steps) == 0 {
		return nil, fmt.Errorf("missing queries")
	}

	return &blueprint, nil
}

func (b *Blueprint) Apply(f *File) ([]ScopedValues, error) {
	found := []ScopedValues{}

	value, err := fileToValue(f)
	if err != nil {
		return nil, NewE100(f.Path, err)
	}

	for _, s := range b.Steps {
		selected, err := dasel.Select(value, s.Operation)
		if err != nil {
			return found, err
		}

		values := []string{}
		for _, v := range selected {
			values = append(values, v.String())
		}

		found = append(found, ScopedValues{
			Scope:  s.Scope,
			Values: values,
		})
	}

	return found, nil
}

func fileToValue(f *File) (DaselValue, error) {
	var value DaselValue

	contents := []byte(f.Content)
	switch f.RealExt {
	case ".json":
		err := json.Unmarshal(contents, &value)
		if err != nil {
			return nil, err
		}
	case ".yml":
		err := yaml.Unmarshal(contents, &value)
		if err != nil {
			return nil, err
		}
	case ".toml":
		err := toml.Unmarshal(contents, &value)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported file type")
	}

	return value, nil
}
