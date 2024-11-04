package code

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/javascript"
)

func JavaScript() *Language {
	return &Language{
		Delims: regexp.MustCompile(`//|/\*\*?|\*/`),
		Parser: javascript.GetLanguage(),
		//Cutset:  " *",
		Queries: []string{`(comment)+ @comment`},
		Padding: cStyle,
	}
}
