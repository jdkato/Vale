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
			found, err := blueprint.Apply(f)
			if err != nil {
				return core.NewE201FromTarget(
					err.Error(),
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
	for _, matches := range values {
		l.SetMetaScope(matches.Scope)
		for _, v := range matches.Values {
			i, line := findLineBySubstring(wholeFile, v)
			if i == 0 {
				return core.NewE100(f.Path, fmt.Errorf("'%s' not found", v))
			}
			f.SetText(v)

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
				padding := strings.Index(line, v)
				f.Alerts = adjustPos(f.Alerts, last, i, padding)
			}
			last = size
		}
	}

	return err
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
