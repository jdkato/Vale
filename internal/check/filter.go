package check

import (
	"fmt"
	"os"
	"strings"

	"github.com/expr-lang/expr"

	"github.com/errata-ai/vale/v3/internal/core"
)

func filter(mgr *Manager) (map[string]Rule, error) {
	stringOrPath := mgr.Config.Flags.Filter

	if stringOrPath == "" {
		return mgr.rules, nil
	} else if !core.FileExists(stringOrPath) {
		name := stringOrPath

		stringOrPath = core.FindAsset(mgr.Config, name)
		if stringOrPath == "" {
			return nil, fmt.Errorf("filter '%s' not found", name)
		}
	}

	b, err := os.ReadFile(stringOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read filter '%s': %w", stringOrPath, err)
	}

	filter := string(b)
	// .Name, .Level -> override
	// .Scope, .Message, .Description, .Extends, .Link
	//
	// The idea here should be simple: we read the ini file and apply overrides
	// (where needed) from the user-given filter. The order is always:
	//
	// ini -> filter
	//
	// The key is that the *filter* always has the last say -- in terms of what
	// rules run and at what level.
	//
	// NOTE: This means that filtered results can only ever be a *subset* of
	// the would-be results since we're filtering on checks loaded based on the
	// ini config.

	env := FilterEnv{}
	for _, rule := range mgr.rules {
		env.Rules = append(env.Rules, rule.Fields())
	}
	code := fmt.Sprintf(`filter(Rules, {%s})`, filter)

	program, err := expr.Compile(code, expr.Env(env))
	if err != nil {
		return mgr.rules, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return mgr.rules, err
	}

	filtered := map[string]Rule{}
	for _, entry := range output.([]interface{}) {
		rule, _ := entry.(Definition)

		name := rule.Name
		if strings.Count(name, ".") > 1 {
			// TODO: See lint.go#249.
			list := strings.Split(name, ".")
			name = strings.Join([]string{list[0], list[1]}, ".")
		}

		// NOTE: We can't simply assume that what the filter returns should be
		// run -- it depends on the *intent* of the filter.
		//
		// If the filter *only* sets `.Level`, then, for example, the output
		// could contain rules that match the new level but are disabled in the
		// `.vale.ini`.
		//
		// TODO: If checking for the existence of, e.g., `.Level` enough?
		// Should we use `program.Constants`?

		if strings.Contains(code, ".Level") {
			lvl := core.LevelToInt[rule.Level]
			if lvl < mgr.Config.MinAlertLevel {
				mgr.Config.MinAlertLevel = lvl
			}
		}

		/*
			if strings.Contains(code, ".Name") {
				mgr.Config.GChecks[name] = true
				for _, v := range mgr.Config.SChecks {
					v[name] = true
				}
			}*/

		filtered[name] = mgr.rules[name]
	}

	return filtered, nil
}
