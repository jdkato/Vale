package lint

import (
	"fmt"
	"strings"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/glob"
)

func (l *Linter) lintData(f *core.File) error {
	for syntax, blueprint := range l.Manager.Config.Blueprints {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return err
		} else if sec.Match(f.Path) {
			found, berr := blueprint.Apply(f)
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

func (l *Linter) lintScopedValues(f *core.File, values []core.ScopedValues) error {
	var err error
	// We want to set up our processing servers as if we were dealing with
	// a directory since we likely have many fragments to convert.
	l.HasDir = true

	wholeFile := f.Content
	last := 0
	index := 0

	for _, match := range values {
		l.SetMetaScope(match.Scope)
		for _, v := range match.Values {
			i, line := findLineBySubstring(wholeFile, v, index)
			if i == 0 {
				return core.NewE100(f.Path, fmt.Errorf("'%s' not found", v))
			}
			index = i

			f.SetText(v)
			f.SetNormedExt(match.Format)

			switch match.Format {
			case "md":
				err = l.lintMarkdown(f)
			case "rst":
				err = l.lintRST(f)
			case "html":
				err = l.lintHTML(f)
			case "org":
				err = l.lintOrg(f)
			case "adoc":
				err = l.lintADoc(f)
			default:
				err = l.lintLines(f)
			}

			size := len(f.Alerts)
			if size != last {
				padding := strings.Index(line, v)
				f.Alerts = adjustPos(f.Alerts, last, i, padding)
			}
			last = size
		}
	}

	return err
}

func findLineBySubstring(s, sub string, last int) (int, string) {
	if strings.Count(sub, "\n") > 0 {
		sub = strings.Split(sub, "\n")[0]
	}

	for i, line := range strings.Split(s, "\n") {
		if i >= last && strings.Contains(line, sub) {
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
