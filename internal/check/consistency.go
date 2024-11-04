package check

import (
	"fmt"
	"strings"

	"github.com/errata-ai/regexp2"
	"github.com/mitchellh/mapstructure"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/nlp"
)

type step struct {
	pattern *regexp2.Regexp
	subs    []string
}

// Consistency ensures that the keys and values of Either don't both exist.
type Consistency struct {
	Definition `mapstructure:",squash"`
	steps      []step
	// `either` (`map`): A map of `option 1: option 2` pairs, of which only one
	// may appear.
	Either map[string]string
	// `nonword` (`bool`): Removes the default word boundaries (`\b`).
	Nonword bool
	// `ignorecase` (`bool`): Makes all matches case-insensitive.
	Ignorecase bool
}

// NewConsistency creates a new `consistency`-based rule.
func NewConsistency(cfg *core.Config, generic baseCheck, path string) (Consistency, error) {
	var chkRE string

	rule := Consistency{}
	name, _ := generic["name"].(string)

	err := mapstructure.WeakDecode(generic, &rule)
	if err != nil {
		return rule, readStructureError(err, path)
	}

	err = checkScopes(rule.Scope, path)
	if err != nil {
		return rule, err
	}

	regex := makeRegexp(
		cfg.WordTemplate,
		rule.Ignorecase,
		func() bool { return !rule.Nonword },
		func() string { return "" }, true)

	chkKey := strings.Split(name, ".")[1]
	count := 0
	for v1, v2 := range rule.Either {
		count += 2

		subs := []string{
			fmt.Sprintf("%s%d", chkKey, count),
			fmt.Sprintf("%s%d", chkKey, count+1)}

		chkRE = fmt.Sprintf("(?P<%s>%s)|(?P<%s>%s)", subs[0], v1, subs[1], v2)
		chkRE = fmt.Sprintf(regex, chkRE)

		re, errc := regexp2.CompileStd(chkRE)
		if errc != nil {
			return rule, core.NewE201FromPosition(errc.Error(), path, 1)
		}

		rule.Extends = name
		rule.Name = fmt.Sprintf("%s.%s", name, v1)
		rule.steps = append(rule.steps, step{pattern: re, subs: subs})
	}

	return rule, nil
}

// Run looks for inconsistent use of a user-defined regex.
func (o Consistency) Run(blk nlp.Block, f *core.File, cfg *core.Config) ([]core.Alert, error) {
	alerts := []core.Alert{}

	loc := []int{}
	txt := blk.Text

	for _, s := range o.steps {
		matches := s.pattern.FindAllStringSubmatchIndex(txt, -1)
		for _, submat := range matches {
			for idx, mat := range submat {
				if mat != -1 && idx > 0 && idx%2 == 0 {
					loc = []int{mat, submat[idx+1]}
					f.Sequences = append(
						f.Sequences,
						s.pattern.SubexpNames()[idx/2])
				}
			}
		}

		if matches != nil && core.AllStringsInSlice(s.subs, f.Sequences) {
			o.Name = o.Extends

			a, err := makeAlert(o.Definition, loc, txt, cfg)
			if err != nil {
				return alerts, err
			}
			alerts = append(alerts, a)
		}
	}

	return alerts, nil
}

// Fields provides access to the internal rule definition.
func (o Consistency) Fields() Definition {
	return o.Definition
}

// Pattern is the internal regex pattern used by this rule.
func (o Consistency) Pattern() string {
	return ""
}
