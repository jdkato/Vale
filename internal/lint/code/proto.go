package code

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/protobuf"
)

func Protobuf() *Language {
	return &Language{
		Delims:  regexp.MustCompile(`//|/\*|\*/`),
		Parser:  protobuf.GetLanguage(),
		Queries: []string{`(comment)+ @comment`},
		Padding: cStyle,
	}
}
