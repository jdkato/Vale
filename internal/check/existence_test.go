package check

import (
	"testing"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/nlp"
)

func makeExistence(tokens []string) (*Existence, error) {
	def := baseCheck{"tokens": tokens}

	cfg, err := core.NewConfig(&core.CLIFlags{})
	if err != nil {
		return nil, err
	}

	rule, err := NewExistence(cfg, def, "")
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func TestExistence(t *testing.T) {
	rule, err := makeExistence([]string{"test"})
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := core.NewConfig(&core.CLIFlags{})
	if err != nil {
		t.Fatal(err)
	}

	file, err := core.NewFile("", cfg)
	if err != nil {
		t.Fatal(err)
	}

	alerts, _ := rule.Run(nlp.NewBlock("", "This is a test.", ""), file, cfg)
	if len(alerts) != 1 {
		t.Errorf("expected one alert, not %v", alerts)
	}
}

func FuzzExistenceInit(f *testing.F) {
	f.Add("hello")
	f.Fuzz(func(_ *testing.T, s string) {
		_, _ = makeExistence([]string{s})
	})
}

func FuzzExistence(f *testing.F) {
	rule, err := makeExistence([]string{"test"})
	if err != nil {
		f.Fatal(err)
	}

	cfg, err := core.NewConfig(&core.CLIFlags{})
	if err != nil {
		f.Fatal(err)
	}

	file, err := core.NewFile("", cfg)
	if err != nil {
		f.Fatal(err)
	}

	f.Add("hello")
	f.Fuzz(func(_ *testing.T, s string) {
		_, _ = rule.Run(nlp.NewBlock("", s, ""), file, cfg)
	})
}
