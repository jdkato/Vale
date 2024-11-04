package code

import (
	"regexp"

	"github.com/jdkato/go-tree-sitter-julia/julia"
)

func Julia() *Language {
	return &Language{
		Delims: regexp.MustCompile(`#|#=|=#`),
		Parser: julia.GetLanguage(),
		Queries: []string{
			`(line_comment)+ @comment`,
			`(block_comment)+ @comment`,
		},
		Padding: func(s string) int {
			return computePadding(s, []string{"#", `#=`, `=#`})
		},
	}
}
