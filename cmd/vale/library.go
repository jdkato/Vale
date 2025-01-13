package main

import "encoding/json"

var library = "https://raw.githubusercontent.com/errata-ai/styles/master/library.json"

func getLibrary(_ string) ([]Style, error) {
	styles := []Style{}

	resp, err := fetchJSON(library)
	if err != nil {
		return styles, err
	} else if err = json.Unmarshal(resp, &styles); err != nil {
		return styles, err
	}

	return styles, err
}

func inLibrary(name, path string) string {
	lookup, err := getLibrary(path)
	if err != nil {
		return ""
	}

	for _, entry := range lookup {
		if name == entry.Name {
			return entry.URL
		}
	}

	return ""
}
