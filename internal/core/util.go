package core

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"unicode"

	"github.com/errata-ai/vale/v3/internal/nlp"
)

var defaultIgnoreDirectories = []string{
	"node_modules", ".git",
}
var spaces = regexp.MustCompile(" +")
var reANSI = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
var sanitizer = strings.NewReplacer(
	"&rsquo;", "'",
	"\r\n", "\n",
	"\r", "\n")

// CapFirst capitalizes the first letter of a string.
func CapFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Sanitize prepares text for our check functions.
func Sanitize(txt string) string {
	// TODO: symbols?
	return sanitizer.Replace(txt)
}

// StripANSI removes all ANSI characters from the given string.
func StripANSI(s string) string {
	return reANSI.ReplaceAllString(s, "")
}

// WhitespaceToSpace converts newlines and multiple spaces (e.g., "  ") into a
// single space.
func WhitespaceToSpace(msg string) string {
	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = spaces.ReplaceAllString(msg, " ")
	return msg
}

// ShouldIgnoreDirectory will check if directory should be ignored
func ShouldIgnoreDirectory(directoryName string) bool {
	for _, directory := range defaultIgnoreDirectories {
		if directory == directoryName {
			return true
		}
	}
	return false
}

// ToSentence converts a slice of terms into sentence.
func ToSentence(words []string, andOrOr string) string {
	l := len(words)

	if l == 1 {
		return fmt.Sprintf("'%s'", words[0])
	} else if l == 2 {
		return fmt.Sprintf("'%s' or '%s'", words[0], words[1])
	}

	wordsForSentence := []string{}
	for _, w := range words {
		wordsForSentence = append(wordsForSentence, fmt.Sprintf("'%s'", w))
	}

	wordsForSentence[l-1] = andOrOr + " " + wordsForSentence[l-1]
	return strings.Join(wordsForSentence, ", ")
}

// IsLetter returns `true` if s contains all letter characters and false if not.
func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsCode returns `true` if s is a code-like token.
func IsCode(s string) bool {
	for _, r := range s {
		if r != '*' && r != '@' {
			return false
		}
	}
	return true
}

// IsPhrase returns `true` is s is a phrase-like token.
//
// This is used to differentiate regex tokens from non-regex.
func IsPhrase(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && r != ' ' && !unicode.IsDigit(r) && r != '-' {
			return false
		}
	}
	return true
}

// InRange determines if the range r contains the integer n.
func InRange(n int, r []int) bool {
	return len(r) == 2 && (r[0] <= n && n <= r[1])
}

// Which checks for the existence of any command in `cmds`.
func Which(cmds []string) string {
	for _, cmd := range cmds {
		path, err := exec.LookPath(cmd)
		if err == nil {
			return path
		}
	}
	return ""
}

// CondSprintf is sprintf, ignores extra arguments.
func CondSprintf(format string, v ...interface{}) string {
	v = append(v, "")
	format += fmt.Sprint("%[", len(v), "]s")
	return fmt.Sprintf(format, v...)
}

// FormatMessage inserts `subs` into `msg`.
func FormatMessage(msg string, subs ...string) string {
	return CondSprintf(msg, StringsToInterface(subs)...)
}

// Substitute replaces the substring `sub` with a string of asterisks.
func Substitute(src, sub string, char rune) (string, bool) {
	idx := strings.Index(src, sub)
	if idx < 0 {
		return src, false
	}
	repl := strings.Map(func(r rune) rune {
		if r != '\n' {
			return char
		}
		return r
	}, sub)
	return strings.Replace(src, sub, repl, 1), true
}

// StringsToInterface converts a slice of strings to an interface.
func StringsToInterface(strings []string) []interface{} {
	intf := make([]interface{}, len(strings))
	for i, v := range strings {
		intf[i] = v
	}
	return intf
}

// Indent adds padding to every line of `text`.
func Indent(text, indent string) string {
	if text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			result += indent + j + "\n"
		}
		return result
	}
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}

// IsDir determines if the path given by `filename` is a directory.
func IsDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

// FileExists determines if the path given by `filename` exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// StringInSlice determines if `slice` contains the string `a`.
func StringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if a == b {
			return true
		}
	}
	return false
}

// IntInSlice determines if `slice` contains the int `a`.
func IntInSlice(a int, slice []int) bool {
	for _, b := range slice {
		if a == b {
			return true
		}
	}
	return false
}

// AllStringsInSlice determines if `slice` contains the `strings`.
func AllStringsInSlice(strings []string, slice []string) bool {
	for _, s := range strings {
		if !StringInSlice(s, slice) {
			return false
		}
	}
	return true
}

// SplitLines splits on CRLF, CR not followed by LF, and LF.
func SplitLines(data []byte, atEOF bool) (adv int, token []byte, err error) { //nolint:nonamedreturns
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\n' {
			return i + 1, data[0:i], nil
		}
		adv = i + 1
		if len(data) > i+1 && data[i+1] == '\n' {
			adv++
		}
		return adv, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func normalizePath(path string) string {
	// expand tilde
	homedir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return homedir
	} else if strings.HasPrefix(path, filepath.FromSlash("~/")) {
		path = filepath.Join(homedir, path[2:])
	}
	return path
}

func TextToContext(text string, meta *nlp.Info) []nlp.TaggedWord {
	context := []nlp.TaggedWord{}

	for idx, line := range strings.Split(text, "\n") {
		plain := stripMarkdown(line)

		pos := 0
		for _, tok := range nlp.TextToTokens(plain, meta) {
			if strings.TrimSpace(tok.Text) != "" {
				s := strings.Index(line[pos:], tok.Text) + len(line[:pos])
				if !StringInSlice(tok.Tag, []string{"''", "``"}) {
					context = append(context, nlp.TaggedWord{
						Line:  idx + 1,
						Token: tok,
						Span:  []int{s + 1, s + len(tok.Text)},
					})
				}
				pos = s
				line, _ = Substitute(line, tok.Text, '*')
			}
		}
	}

	return context
}

func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllStringSubmatchIndex(str, -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			if v[i] == -1 || v[i+1] == -1 {
				groups = append(groups, "")
			} else {
				groups = append(groups, str[v[i]:v[i+1]])
			}
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func HasAnySuffix(s string, suffixes []string) bool {
	n := len(s)
	for _, suffix := range suffixes {
		if n > len(suffix) && strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// ReplaceExt replaces the extension of `fp` with `ext` if the extension of
// `fp` is in `formats`.
//
// This is used in places where we need to normalize file extensions (e.g.,
// `foo.mdx` -> `foo.md`) in order to respect format associations.
func ReplaceExt(fp string, formats map[string]string) string {
	var ext string

	old := filepath.Ext(fp)
	if normed, found := formats[strings.Trim(old, ".")]; found {
		ext = "." + normed
		fp = fp[0:len(fp)-len(old)] + ext
	}

	return fp
}

// FindProcess checks if a process with the given PID exists.
func FindProcess(pid int) *os.Process {
	if pid <= 0 {
		return nil
	}

	p, err := os.FindProcess(pid)
	if runtime.GOOS != "windows" {
		err = p.Signal(os.Signal(syscall.Signal(0)))
	}

	if err != nil {
		return nil
	}

	return p
}

// UniqueStrings returns a new slice with all duplicate strings removed.
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}
