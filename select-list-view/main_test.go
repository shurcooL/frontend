package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dgryski/go-trigram"
	"github.com/shurcooL/go/u/u5"
)

func init() {
	const path = "/Users/Dmitri/Dropbox/Work/2013/Data Sets/all-Go-packages.json"

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var importers u5.Importers
	if err := json.NewDecoder(f).Decode(&importers); err != nil {
		panic(err)
	}

	ss = make([]string, len(importers.Results))
	for i, entry := range importers.Results {
		ss[i] = entry.Path
	}

	// ---

	idx = trigram.NewIndex(ss)
}

var idx trigram.Index
var ss []string

func BenchmarkContains(b *testing.B) {
	for i := 0; i < b.N; i++ {

		var found []string

		filter := "shurcooL"
		//lowerFilter := strings.ToLower(filter)
		lowerFilter := filter

		for _, header := range ss {
			if filter != "" && !strings.Contains(header, lowerFilter) {
				continue
			}

			found = append(found, header)
		}

		if len(found) != 146 {
			b.Error("unexpected found:", found)
		}

	}
}

func BenchmarkTrigram(b *testing.B) {
	for i := 0; i < b.N; i++ {

		var found []string

		filter := "shurcooL"
		//lowerFilter := strings.ToLower(filter)
		lowerFilter := filter

		q := idx.Query(lowerFilter)

		for _, v := range q {
			if filter != "" && !strings.Contains(ss[v], lowerFilter) {
				continue
			}

			found = append(found, ss[v])
		}

		if len(found) != 146 {
			b.Error("unexpected found:", found)
		}

	}
}
