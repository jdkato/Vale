package code

import (
	"regexp"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/smacker/go-tree-sitter/javascript"
)

func JavaScript() *Language {
	return &Language{
		Delims: regexp.MustCompile(`//|/\*\*?|\*/`),
		Parser: javascript.GetLanguage(),
		//Cutset:  " *",
		Queries: []core.Scope{{Name: "", Expr: "(comment) @comment", Type: ""}},
		Padding: cStyle,
	}
}
