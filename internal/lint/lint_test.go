package lint

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/errata-ai/vale/v3/internal/core"
)

func TestGenderBias(t *testing.T) {
	reToMatches := map[string][]string{
		"(?:alumna|alumnus)":          {"alumna", "alumnus"},
		"(?:alumnae|alumni)":          {"alumnae", "alumni"},
		"(?:mother|father)land":       {"motherland", "fatherland"},
		"air(?:m[ae]n|wom[ae]n)":      {"airman", "airwoman", "airmen", "airwomen"},
		"anchor(?:m[ae]n|wom[ae]n)":   {"anchorman", "anchorwoman", "anchormen", "anchorwomen"},
		"camera(?:m[ae]n|wom[ae]n)":   {"cameraman", "camerawoman", "cameramen", "camerawomen"},
		"chair(?:m[ae]n|wom[ae]n)":    {"chairman", "chairwoman", "chairmen", "chairwomen"},
		"congress(?:m[ae]n|wom[ae]n)": {"congressman", "congresswoman", "congressmen", "congresswomen"},
		"door(?:m[ae]n|wom[ae]n)":     {"doorman", "doorwoman", "doormen", "doorwomen"},
		"drafts(?:m[ae]n|wom[ae]n)":   {"draftsman", "draftswoman", "draftsmen", "draftswomen"},
		"fire(?:m[ae]n|wom[ae]n)":     {"fireman", "firewoman", "firemen", "firewomen"},
		"fisher(?:m[ae]n|wom[ae]n)":   {"fisherman", "fisherwoman", "fishermen", "fisherwomen"},
		"fresh(?:m[ae]n|wom[ae]n)":    {"freshman", "freshwoman", "freshmen", "freshwomen"},
		"garbage(?:m[ae]n|wom[ae]n)":  {"garbageman", "garbagewoman", "garbagemen", "garbagewomen"},
		"mail(?:m[ae]n|wom[ae]n)":     {"mailman", "mailwoman", "mailmen", "mailwomen"},
		"middle(?:m[ae]n|wom[ae]n)":   {"middleman", "middlewoman", "middlemen", "middlewomen"},
		"news(?:m[ae]n|wom[ae]n)":     {"newsman", "newswoman", "newsmen", "newswomen"},
		"ombuds(?:man|woman)":         {"ombudsman", "ombudswoman"},
		"work(?:m[ae]n|wom[ae]n)":     {"workman", "workwoman", "workmen", "workwomen"},
		"police(?:m[ae]n|wom[ae]n)":   {"policeman", "policewoman", "policemen", "policewomen"},
		"repair(?:m[ae]n|wom[ae]n)":   {"repairman", "repairwoman", "repairmen", "repairwomen"},
		"sales(?:m[ae]n|wom[ae]n)":    {"salesman", "saleswoman", "salesmen", "saleswomen"},
		"service(?:m[ae]n|wom[ae]n)":  {"serviceman", "servicewoman", "servicemen", "servicewomen"},
		"steward(?:ess)?":             {"steward", "stewardess"},
		"tribes(?:m[ae]n|wom[ae]n)":   {"tribesman", "tribeswoman", "tribesmen", "tribeswomen"},
	}
	for re, matches := range reToMatches {
		regex := regexp.MustCompile(re)
		for _, match := range matches {
			if !regex.MatchString(match) {
				t.Errorf("expected = %v, got = %v", true, false)
			}
		}
	}
}

func initLinter() (*Linter, error) {
	cfg, err := core.NewConfig(&core.CLIFlags{})
	if err != nil {
		return nil, err
	}

	cfg.MinAlertLevel = 0
	cfg.GBaseStyles = []string{"Vale"}
	cfg.Flags.InExt = ".txt" // default value

	return NewLinter(cfg)
}

func benchmarkLint(b *testing.B, path string) {
	b.Helper()

	linter, err := initLinter()
	if err != nil {
		b.Fatal(err)
	}

	path, err = filepath.Abs(path)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		_, err = linter.Lint([]string{path}, "*")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLintRST(b *testing.B) {
	benchmarkLint(b, "../../testdata/fixtures/benchmarks/bench.rst")
}

func BenchmarkLintMD(b *testing.B) {
	benchmarkLint(b, "../../testdata/fixtures/benchmarks/bench.md")
}
