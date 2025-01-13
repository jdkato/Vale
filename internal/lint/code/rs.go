package code

import (
	"regexp"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/smacker/go-tree-sitter/rust"
)

func Rust() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`/{2,3}!?`),
		Parser:  rust.GetLanguage(),
		Queries: []core.Scope{{Name: "", Expr: `(line_comment)+ @comment`, Type: ""}},
		Padding: func(s string) int {
			return computePadding(s, []string{"//", "//!", "///"})
		},
	}
}
