package lint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/tomwright/dasel/v2"
	"gopkg.in/yaml.v2"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/glob"
)

type DaselValue = map[string]any

type DaselQuery struct {
	name   string `yaml:"name"`
	Query  string `yaml:"query"`
	Scope  string `yaml:"scope"`
	Format string `yaml:"format"`
}

type Blueprint struct {
	Select []DaselQuery `yaml:"select"`
}

type ScopedValue struct {
	Scope  string
	Value  dasel.Values
	Format string
}

func (l *Linter) lintData(f *core.File) error {
	value, err := fileToValue(f)
	if err != nil {
		return core.NewE100(f.Path, err)
	}

	for syntax, blueprint := range l.Manager.Config.Blueprints {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return err
		} else if sec.Match(f.Path) {
			query, err := readBlueprint(blueprint, l.Manager.Config)
			if err != nil {
				return core.NewE201FromTarget(
					err.Error(),
					fmt.Sprintf("Blueprint = %s", blueprint),
					l.Manager.Config.RootINI,
				)
			}

			found, berr := applyBlueprint(value, query)
			if berr != nil {
				return core.NewE201FromTarget(
					berr.Error(),
					fmt.Sprintf("Blueprint = %s", blueprint),
					l.Manager.Config.RootINI,
				)
			}

			return l.lintScopedValues(f, found)
		}
	}

	return nil
}

func (l *Linter) lintScopedValues(f *core.File, values []ScopedValue) error {
	var err error
	// We want to set up our processing servers as if we were dealing with
	// a directory since we likely have many fragments to convert.
	l.HasDir = true

	wholeFile := f.Content

	last := 0
	for _, value := range values {
		for _, v := range value.Value {
			i, line := findLineBySubstring(wholeFile, v.String())
			if i == 0 {
				return core.NewE100(f.Path, fmt.Errorf("'%s' not found", v.String()))
			}
			f.SetText(v.String())

			switch f.NormedExt {
			case ".md":
				err = l.lintMarkdown(f)
			case ".rst":
				err = l.lintRST(f)
			case ".xml":
				err = l.lintADoc(f)
			case ".html":
				err = l.lintHTML(f)
			case ".org":
				err = l.lintOrg(f)
			default:
				err = l.lintLines(f)
			}

			size := len(f.Alerts)
			if size != last {
				padding := strings.Index(line, v.String())
				f.Alerts = adjustPos(f.Alerts, last, i, padding)
			}
			last = size
		}
	}

	return err
}

func applyBlueprint(value DaselValue, blueprint *Blueprint) ([]ScopedValue, error) {
	found := []ScopedValue{}

	for _, s := range blueprint.Select {
		values, err := dasel.Select(value, s.Query)
		if err != nil {
			return found, err
		}
		found = append(found, ScopedValue{
			Scope:  s.Scope,
			Value:  values,
			Format: s.Format,
		})
	}

	return found, nil
}

func readBlueprint(blueprint string, cfg *core.Config) (*Blueprint, error) {
	var query Blueprint

	file := core.FindConfigAsset(cfg, blueprint+".yml", core.BlueprintsDir)
	if file == "" {
		return nil, fmt.Errorf("blueprint '%s' not found", blueprint)
	}

	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, &query)
	if err != nil {
		return nil, err
	}

	return &query, nil
}

func fileToValue(f *core.File) (DaselValue, error) {
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

func findLineBySubstring(s, sub string) (int, string) {
	for i, line := range strings.Split(s, "\n") {
		if strings.Contains(line, sub) {
			return i + 1, line
		}
	}
	return 0, ""
}

func adjustPos(alerts []core.Alert, last, line, padding int) []core.Alert {
	for i := range alerts {
		if i >= last {
			alerts[i].Line += line - 1
			alerts[i].Span = []int{
				alerts[i].Span[0] + padding,
				alerts[i].Span[1] + padding,
			}
		}
	}
	return alerts
}
