package code

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/yaml"
)

func YAML() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`#`),
		Parser:  yaml.GetLanguage(),
		Queries: []string{`(comment)+ @comment`},
		Padding: func(s string) int {
			return computePadding(s, []string{"#"})
		},
	}
}
