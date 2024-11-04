package lint

import (
	"strings"
	"testing"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_applyPatterns(t *testing.T) {
	cases := []struct {
		description string
		conf        core.Config
		exts        extensionConfig
		content     string
		expected    string
	}{
		{
			description: "MDX comment in markdown, custom comment delimiter",
			conf: core.Config{
				CommentDelimiters: map[string][2]string{
					".md": [2]string{"{/*", "*/}"},
				},
			},
			exts: extensionConfig{".md", ".md"},
			content: `---
title: Example page
description: Example page
---

This is the intro pagragraph.

{/* This is a comment */}
`,
			expected: strings.ReplaceAll(`
@@@
title: Example page
description: Example page
@@@


This is the intro pagragraph.

<!-- This is a comment -->
`, "@", "`"),
		},
		{
			description: "MDX comment in markdown, no custom comment delimiter",
			conf:        core.Config{},
			exts:        extensionConfig{".md", ".md"},
			content: `---
title: Example page
description: Example page
---

This is the intro pagragraph.

{/* This is a comment */}
`,
			expected: strings.ReplaceAll(`
@@@
title: Example page
description: Example page
@@@


This is the intro pagragraph.

{/* This is a comment */}
`, "@", "`"),
		},
		{
			description: "multiline MDX comment in markdown, custom comment delimiter",
			conf: core.Config{
				CommentDelimiters: map[string][2]string{
					".md": [2]string{"{/*", "*/}"},
				},
			},
			exts: extensionConfig{".md", ".md"},
			content: `---
title: Example page
description: Example page
---

This is the intro pagragraph.

{/* 
This is a comment 
*/}
`,
			expected: strings.ReplaceAll(`
@@@
title: Example page
description: Example page
@@@


This is the intro pagragraph.

<!-- 
This is a comment 
-->
`, "@", "`"),
		},
		{
			description: "token ignore in cc file",
			content:     "Call \\c func to start the process.",
			conf: core.Config{
				TokenIgnores: map[string][]string{
					"*.cc": []string{`(\\c \w+)`},
				},
				Formats: map[string]string{
					"cc": "md",
				},
			},
			exts:     extensionConfig{".md", ".cc"},
			expected: "Call `\\c func` to start the process.",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			s, err := applyPatterns(&c.conf, c.exts, c.content)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, s)
		})
	}
}

func Test_applyPatterns_errors(t *testing.T) {
	cases := []struct {
		description string
		conf        core.Config
		exts        extensionConfig
		content     string
		expectedErr string
	}{
		{
			description: "only one delimiter",
			conf: core.Config{
				CommentDelimiters: map[string][2]string{
					".md": [2]string{"{/*", ""},
				},
			},
			exts: extensionConfig{".md", ".md"},
			content: `---
title: Example page
description: Example page
---

This is the intro pagragraph.

{/* This is a comment */}
`,
			expectedErr: "",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			_, err := applyPatterns(&c.conf, c.exts, c.content)
			assert.ErrorContains(t, err, c.expectedErr)
		})
	}
}

// TODO: Test for expected errors resulting from applyPatterns
