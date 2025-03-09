package lint

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/nlp"
	"github.com/errata-ai/vale/v3/internal/system"
)

func (l Linter) lintMDX(f *core.File) error {
	var html string
	var err error

	exe := system.Which([]string{"mdx2vast"})
	if exe == "" {
		return core.NewE100("lintMDX", errors.New("mdx2vast not found"))
	}

	s, err := l.Transform(f)
	if err != nil {
		return err
	}

	html, err = callVast(f, s, exe)
	if err != nil {
		return core.NewE100(f.Path, err)
	}

	// NOTE: This is required to avoid finding matches inside info strings. For
	// example, if we're looking for 'json' we many incorrectly report the
	// location as being in an infostring like '```json'.
	//
	// See https://github.com/errata-ai/vale/v2/issues/248.
	body := reExInfo.ReplaceAllStringFunc(f.Content, func(m string) string {
		parts := strings.Split(m, "`")

		// This ensures that we respect the number of opening backticks, which
		// could be more than 3.
		//
		// See https://github.com/errata-ai/vale/v2/issues/271.
		tags := strings.Repeat("`", len(parts)-1)
		span := strings.Repeat("*", nlp.StrLen(parts[len(parts)-1]))

		return tags + span
	})

	// NOTE: This is required to avoid finding matches inside link references.
	body = reLinkRef.ReplaceAllStringFunc(body, func(m string) string {
		return "][" + strings.Repeat("*", nlp.StrLen(m)-3) + "]"
	})
	body = reLinkDef.ReplaceAllStringFunc(body, func(m string) string {
		return "[" + strings.Repeat("*", nlp.StrLen(m)-3) + "]:"
	})

	// NOTE: This is required to avoid finding matches inside ordered lists.
	body = reNumericList.ReplaceAllStringFunc(body, func(m string) string {
		return strings.Repeat("*", nlp.StrLen(m))
	})

	f.Content = body
	return l.lintHTMLTokens(f, []byte(html), 0)
}

func callVast(_ *core.File, text, exe string) (string, error) {
	var out bytes.Buffer
	var eut bytes.Buffer

	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = &out
	cmd.Stderr = &eut

	if err := cmd.Run(); err != nil {
		return "", errors.New(eut.String())
	}

	return out.String(), nil
}
