package main

import (
	"encoding/json"
	"os"
	"text/template"

	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"
)

var funcs = template.FuncMap{}

func init() {
	funcs["red"] = func(s string) string {
		return pterm.Red(s)
	}
	funcs["blue"] = func(s string) string {
		return pterm.Blue(s)
	}
	funcs["yellow"] = func(s string) string {
		return pterm.Yellow(s)
	}
	funcs["underline"] = func(s string) string {
		return pterm.Underscore.Sprint(s)
	}
	funcs["newTable"] = func(wrap bool) *tablewriter.Table {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetAutoWrapText(wrap)
		return table
	}
	funcs["addRow"] = func(t *tablewriter.Table, r []string) *tablewriter.Table {
		t.Append(r)
		return t
	}
	funcs["renderTable"] = func(t *tablewriter.Table) *tablewriter.Table {
		t.Render()
		t.ClearRows()
		return t
	}
	funcs["jsonEscape"] = func(i string) string {
		b, err := json.Marshal(i)
		if err != nil {
			panic(err)
		}
		return string(b[1 : len(b)-1])
	}
}
