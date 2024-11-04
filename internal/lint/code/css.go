package code

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/css"
)

func CSS() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`/\*!?|\*/`),
		Parser:  css.GetLanguage(),
		Queries: []string{`(comment)+ @comment`},
		Padding: func(s string) int {
			return computePadding(s, []string{"/*"})
		},
	}
}
