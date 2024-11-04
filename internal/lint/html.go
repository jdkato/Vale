package lint

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/errata-ai/regexp2"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/glob"
)

var reFrontMatter = regexp.MustCompile(
	`^(?s)(?:---|\+\+\+)\n(.+?)\n(?:---|\+\+\+)`)

var heading = regexp.MustCompile(`^h\d$`)

func (l *Linter) lintHTML(f *core.File) error {
	if l.Manager.Config.Flags.Built != "" {
		return l.lintTxtToHTML(f)
	}
	return l.lintHTMLTokens(f, []byte(f.Content), 0)
}

type extensionConfig struct {
	Normed, Real string
}

var blockDelimiters = map[string]string{
	".adoc": "\n----\n$1\n----\n",
	".md":   "\n```\n$1\n```\n",
	".rst":  "\n::\n\n%s\n",
	".org":  orgExample,
}

func applyBlockPatterns(c *core.Config, exts extensionConfig, content string) (string, error) {
	block, ok := blockDelimiters[exts.Normed]
	if !ok {
		return content, fmt.Errorf("ignore patterns are not supported in '%s' files", exts.Normed)
	}

	// TODO: Should we assume this?
	s := reFrontMatter.ReplaceAllString(content, block)

	for syntax, regexes := range c.BlockIgnores {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return s, err
		} else if sec.Match(exts.Normed) || sec.Match(exts.Real) {
			for _, r := range regexes {
				pat, errc := regexp2.CompileStd(r)
				if errc != nil { //nolint:gocritic
					return s, core.NewE201FromTarget(
						errc.Error(),
						r,
						c.Flags.Path,
					)
				} else if strings.HasSuffix(exts.Normed, ".rst") {
					// HACK: We need to add padding for the literal block.
					for _, c := range pat.FindAllStringSubmatch(s, -1) {
						sec := fmt.Sprintf(block, core.Indent(c[0], "    "))
						s = strings.Replace(s, c[0], sec, 1)
					}
				} else {
					s, err = pat.Replace(s, block, 0, -1)
					if err != nil {
						return s, core.NewE201FromTarget(
							err.Error(),
							r,
							c.Flags.Path,
						)
					}
				}
			}
		}
	}
	return s, nil
}

var inlineDelimiters = map[string]string{
	".adoc": "`$1`",
	".md":   "`$1`",
	".rst":  "``$1``",
	".org":  "=$1=",
}

func applyInlinePatterns(c *core.Config, exts extensionConfig, content string) (string, error) {
	inline, ok := inlineDelimiters[exts.Normed]
	if !ok {
		return content, fmt.Errorf("ignore patterns are not supported in '%s' files", exts.Normed)
	}

	for syntax, regexes := range c.TokenIgnores {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return content, err
		} else if sec.Match(exts.Normed) || sec.Match(exts.Real) {
			for _, r := range regexes {
				pat, errc := regexp2.CompileStd(r)
				if errc != nil {
					return content, core.NewE201FromTarget(
						errc.Error(),
						r,
						c.Flags.Path,
					)
				}
				content, err = pat.Replace(content, inline, 0, -1)
				if err != nil {
					return content, core.NewE201FromTarget(
						err.Error(),
						r,
						c.Flags.Path,
					)
				}
			}
		}
	}
	return content, nil
}

// applyCommentPatterns replaces any custom comment delimiters with HTML comment
// tags based on the user configuration. This makes it possible to apply
// comment-based controls using custom comment delimiters.
func applyCommentPatterns(c *core.Config, exts extensionConfig, content string) (string, error) {
	for syntax, delims := range c.CommentDelimiters {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return content, err
		} else if sec.Match(exts.Normed) || sec.Match(exts.Real) {
			// This field was not assigned, so do nothing.
			if delims[0] == "" && delims[1] == "" {
				return content, nil
			}
			// Return an error if only one delimiter is configured
			if (delims[0] == "" && delims[1] != "") || (delims[0] != "" && delims[1] == "") {
				return content, fmt.Errorf("CommentDelimiters must be empty or have two values")
			}

			content = strings.ReplaceAll(content, delims[0], "<!--")
			content = strings.ReplaceAll(content, delims[1], "-->")
		}
	}
	return content, nil
}

func applyPatterns(c *core.Config, exts extensionConfig, content string) (string, error) {
	s, err := applyBlockPatterns(c, exts, content)
	if err != nil {
		return s, err
	}

	s, err = applyInlinePatterns(c, exts, s)
	if err != nil {
		return s, err
	}

	s, err = applyCommentPatterns(c, exts, s)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (l *Linter) lintTxtToHTML(f *core.File) error {
	html, err := os.ReadFile(l.Manager.Config.Flags.Built)
	if err != nil {
		return core.NewE100(f.Path, err)
	}
	return l.lintHTMLTokens(f, html, 0)
}
