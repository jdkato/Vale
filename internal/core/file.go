package core

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jdkato/twine/summarize"

	"github.com/errata-ai/vale/v3/internal/glob"
	"github.com/errata-ai/vale/v3/internal/nlp"
)

var commentControlRE = regexp.MustCompile(`^vale (.+\..+|[^.]+) = (YES|NO|on|off)$`)

var commentStyleRE = regexp.MustCompile(`^vale styles? = (.*)$`)

// A File represents a linted text file.
type File struct {
	NLP        nlp.Info          // -
	Summary    bytes.Buffer      // holds content to be included in summarization checks
	Alerts     []Alert           // all alerts associated with this file
	BaseStyles []string          // base style assigned in .vale
	Lines      []string          // the File's Content split into lines
	Sequences  []string          // tracks various info (e.g., defined abbreviations)
	Content    string            // the raw file contents
	Format     string            // 'code', 'markup' or 'prose'
	NormedExt  string            // the normalized extension (see util/format.go)
	Path       string            // the full path
	NormedPath string            // the normalized path
	Transform  string            // XLST transform
	RealExt    string            // actual file extension
	Checks     map[string]bool   // syntax-specific checks assigned in .vale
	ChkToCtx   map[string]string // maps a temporary context to a particular check
	Comments   map[string]bool   // comment control statements
	Metrics    map[string]int    // count-based metrics
	history    map[string]int    // -
	limits     map[string]int    // -
	simple     bool              // -
	Lookup     bool              // -
}

// NewFile initializes a File.
func NewFile(src string, config *Config) (*File, error) {
	var format, ext string
	var fbytes []byte
	var lookup bool

	if FileExists(src) {
		fbytes, _ = os.ReadFile(src)
		if config.Flags.InExt != ".txt" {
			ext, format = FormatFromExt(config.Flags.InExt, config.Formats)
		} else {
			ext, format = FormatFromExt(src, config.Formats)
		}
	} else {
		ext, format = FormatFromExt(config.Flags.InExt, config.Formats)
		fbytes = []byte(src)
		src = "stdin" + config.Flags.InExt
		lookup = true
	}
	filepaths := []string{src}

	normed := ReplaceExt(src, config.Formats)
	if normed != src {
		// NOTE: In retrospect, this was a mistake: we should NOT normalize
		// the extension with respect to the `.vale.ini` file.
		//
		// The `.vale.ini` file should reflect the actual file extensions (as
		// they appear on disk). Unfortunately, changing this behavior entirely
		// would break backwards compatibility with many configurations.
		//
		// So, as a workaround, we check both cases. This means that there are
		// two cases:
		//
		// - No assigned format: No change (no normed path).
		//
		// - Assigned format: We can reference the file using the normed path
		// (old behavior) or the actual path (desired behavior).
		//
		// See also `Linter.skip`.
		filepaths = append(filepaths, normed)
	}

	baseStyles := config.GBaseStyles
	checks := make(map[string]bool)

	for _, fp := range filepaths {
		for _, sec := range config.StyleKeys {
			if pat, found := config.SecToPat[sec]; found && pat.Match(fp) {
				baseStyles = config.SBaseStyles[sec]
			}
		}

		for _, sec := range config.RuleKeys {
			if pat, found := config.SecToPat[sec]; found && pat.Match(fp) {
				for k, v := range config.SChecks[sec] {
					checks[k] = v
				}
			}
		}
	}

	lang := "en"
	for syntax, code := range config.FormatToLang {
		sec, err := glob.Compile(syntax)
		if err != nil {
			return &File{}, err
		} else if sec.Match(src) {
			lang = code
			break
		}
	}

	transform := ""
	for sec, p := range config.Stylesheets {
		pat, err := glob.Compile(sec)
		if err != nil {
			return &File{}, NewE100(src, err)
		} else if pat.Match(src) {
			transform = p
			break
		}
	}
	content := Sanitize(string(fbytes))

	// NOTE: We need to perform a clone here because we perform inplace editing
	// of the files contents that we don't want reflected in `lines`.
	//
	// See lint/walk.go.
	lines := strings.SplitAfter(strings.Clone(content), "\n")

	file := File{
		NormedExt: ext, Format: format, RealExt: filepath.Ext(src),
		BaseStyles: baseStyles, Checks: checks, Lines: lines, Content: content,
		Comments: make(map[string]bool), history: make(map[string]int),
		simple: config.Flags.Simple, Transform: transform,
		limits: make(map[string]int), Path: src, Metrics: make(map[string]int),
		NLP:    nlp.Info{Endpoint: config.NLPEndpoint, Lang: lang},
		Lookup: lookup, NormedPath: normed,
	}

	return &file, nil
}

// SortedAlerts returns all of f's alerts sorted by line and column.
func (f *File) SortedAlerts() []Alert {
	sort.Sort(ByPosition(f.Alerts))
	return f.Alerts
}

// ComputeMetrics returns all of f's metrics.
func (f *File) ComputeMetrics() (map[string]interface{}, error) {
	params := map[string]interface{}{}

	doc := summarize.NewDocument(f.Summary.String())
	if doc.NumWords == 0 {
		return params, nil
	}

	for k, v := range f.Metrics {
		if strings.HasPrefix(k, "table") {
			continue
		}
		k = strings.ReplaceAll(k, ".", "_")
		params[k] = float64(v)
	}

	params["complex_words"] = doc.NumComplexWords
	params["long_words"] = doc.NumLongWords
	params["paragraphs"] = doc.NumParagraphs - 1
	params["sentences"] = doc.NumSentences
	params["characters"] = doc.NumCharacters
	params["words"] = doc.NumWords
	params["polysyllabic_words"] = doc.NumPolysylWords
	params["syllables"] = doc.NumSyllables

	return params, nil
}

// FindLoc calculates the line and span of an Alert.
func (f *File) FindLoc(ctx, s string, pad, count int, a Alert) (int, []int) {
	var length int
	var lines []string

	for _, s := range a.Offset {
		ctx, _ = Substitute(ctx, s, '@')
	}

	pos, substring := initialPosition(ctx, s, a)
	if pos < 0 {
		// Shouldn't happen ...
		return pos, []int{0, 0}
	}

	loc := a.Span
	if f.Format == "markup" && !f.simple || f.Format == "fragment" {
		lines = f.Lines
	} else {
		lines = strings.SplitAfter(ctx, "\n")
	}

	counter := 0
	for idx, l := range lines {
		length = nlp.StrLen(l)
		if (counter + length) >= pos {
			loc[0] = (pos - counter) + pad
			loc[1] = loc[0] + nlp.StrLen(substring) - 1
			extent := length + pad
			if loc[1] > extent {
				loc[1] = extent
			} else if loc[1] <= 0 {
				loc[1] = 1
			}
			return count - (len(lines) - (idx + 1)), loc
		}
		counter += length
	}

	return count, loc
}

func (f *File) assignLoc(ctx string, blk nlp.Block, pad int, a Alert) (int, []int) {
	loc := a.Span
	for idx, l := range strings.SplitAfter(ctx, "\n") {
		if loc[0] < 0 || loc[1] < 0 {
			continue
		}
		// NOTE: This fixes #473, but the real issue is that `blk.Line` is
		// wrong. This seems related to `location.go#41`, but I'm not sure.
		//
		// At the very least, this change includes a representative test case
		// and a temporary fix.
		exact := len(l) > loc[1] && l[loc[0]:loc[1]] == a.Match
		if exact || idx == blk.Line {
			length := nlp.StrLen(l)
			pos, substring := initialPosition(l, blk.Text, a)

			loc[0] = pos + pad
			loc[1] = pos + nlp.StrLen(substring) - 1

			extent := length + pad
			if loc[1] > extent {
				loc[1] = extent
			} else if loc[1] <= 0 {
				loc[1] = 1
			}

			return idx + 1, loc
		}
	}
	return blk.Line + 1, a.Span
}

// SetText updates the file's content, lines, and history.
func (f *File) SetText(s string) {
	f.Content = s
	f.Lines = strings.SplitAfter(s, "\n")
	f.history = map[string]int{}
}

// AddAlert calculates the in-text location of an Alert and adds it to a File.
func (f *File) AddAlert(a Alert, blk nlp.Block, lines, pad int, lookup bool) {
	ctx := blk.Context
	if old, ok := f.ChkToCtx[a.Check]; ok {
		ctx = old
	}

	// NOTE: If the `ctx` document is large (as could be the case with
	// `scope: raw`) this is *slow*. Thus, the cap at 1k.
	//
	// TODO: Actually fix this.
	if len(a.Offset) == 0 && strings.Count(ctx, a.Match) > 1 && len(ctx) < 1000 {
		a.Offset = append(a.Offset, strings.Fields(ctx[0:a.Span[0]])...)
	}

	if !lookup {
		a.Line, a.Span = f.assignLoc(ctx, blk, pad, a)
	}
	if (!lookup && a.Span[0] < 0) || lookup {
		a.Line, a.Span = f.FindLoc(ctx, blk.Text, pad, lines, a)
	}

	if a.Span[0] > 0 {
		f.ChkToCtx[a.Check], _ = Substitute(ctx, a.Match, '#')
		if !a.Hide {
			// Ensure that we're not double-reporting an Alert:
			entry := strings.Join([]string{
				strconv.Itoa(a.Line),
				strconv.Itoa(a.Span[0]),
				a.Check}, "-")

			if _, found := f.history[entry]; !found {
				// Check rule-assigned limits for reporting:
				count, occur := f.limits[a.Check]
				if (!occur || a.Limit == 0) || count < a.Limit {
					f.Alerts = append(f.Alerts, a)

					f.history[entry] = 1
					if a.Limit > 0 {
						f.limits[a.Check]++
					}
				}
			}
		}
	}
}

// UpdateComments sets a new status based on comment.
func (f *File) UpdateComments(comment string) {
	if comment == "vale off" { //nolint:gocritic
		f.Comments["off"] = true
	} else if comment == "vale on" {
		f.Comments["off"] = false
	} else if commentControlRE.MatchString(comment) {
		check := commentControlRE.FindStringSubmatch(comment)
		if len(check) == 3 {
			f.Comments[check[1]] = (check[2] == "NO" || check[2] == "off")
		}
	} else if commentStyleRE.MatchString(comment) {
		for _, style := range f.BaseStyles {
			f.Comments[style] = true
		}
		check := commentStyleRE.FindStringSubmatch(comment)
		for _, style := range strings.Split(check[1], ", ") {
			f.Comments[style] = false
		}
	}
}

// QueryComments checks if there has been an in-text comment for this check.
func (f *File) QueryComments(check string) bool {
	if f.Comments["off"] {
		return true
	}
	if style, _, ok := strings.Cut(check, "."); ok {
		if status := f.Comments[style]; status {
			return true
		}
	}
	if status := f.Comments[check]; status {
		return true
	}
	return false
}

// ResetComments resets the state of all checks back to active.
func (f *File) ResetComments() {
	for check := range f.Comments {
		if check != "off" {
			f.Comments[check] = false
		}
	}
}
