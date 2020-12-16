package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
	Titles        []string
	ContentIndex  []int
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.SearchWithTitles(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	ns := []byte(strings.ToLower(s.CompleteWorks))
	s.SuffixArray = suffixarray.New(ns)
	if err := s.IndexTitles(dat); err != nil {
		return err
	}
	return nil
}

func (s *Searcher) Search(query string) []string {
	nq := strings.ToLower(query)
	idxs := s.SuffixArray.Lookup([]byte(nq), -1)
	results := []string{}
	for _, idx := range idxs {
		results = append(results, s.CompleteWorks[idx-250:idx+250])
	}
	return results
}

type titledContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func (s *Searcher) SearchWithTitles(query string) []titledContent {
	nq := strings.ToLower(query)
	idxs := s.SuffixArray.Lookup([]byte(nq), -1)
	var res []titledContent
	for _, idx := range idxs {
		res = append(res, titledContent{s.TitleLookUp(idx), s.CompleteWorks[idx-250 : idx+250]})
	}
	return res
}

//var reTitle = regexp.MustCompile(`by William Shakespeare\s+Contents\s+([\s\S]*)\n{4,}`)
var reIdxTitle = regexp.MustCompile(`by William Shakespeare\s+Contents\s+([\s\S]*)THE SONNETS`)

func (s *Searcher) IndexTitles(b []byte) error {
	// Extract table of contents string to get titles.
	sa := suffixarray.New(b)
	tocIdxs := sa.FindAllIndex(reIdxTitle, -1)
	if len(tocIdxs) == 0 {
		return fmt.Errorf("could not find table of contents")
	}
	toc := tocIdxs[0]
	endOfToC := toc[1]

	s.Titles = findTitles(reIdxTitle, s.CompleteWorks[toc[0]:endOfToC])
	s.ContentIndex = findContentIndex(sa, s.Titles)

	return nil
}

func (s *Searcher) TitleLookUp(idx int) string {
	for i, title := range s.Titles {
		startIdx := s.ContentIndex[i]
		if startIdx == -1 {
			continue
		}

		// Are we at the end of the index stack?
		if len(s.ContentIndex)-1 == i {
			if startIdx > idx {
				continue
			}

			return title
		}

		// Do we have detected end index?
		endIdx := s.ContentIndex[i+1]
		if endIdx == -1 {
			continue
		}

		if startIdx < idx && endIdx > idx {
			return title
		}
	}

	// Could not detect title.
	return ""
}

// findTitles Clean and collect title strings.
func findTitles(re *regexp.Regexp, str string) []string {
	var ts []string
	s := re.FindStringSubmatch(str)
	for _, t := range strings.Split(s[1], "\n") {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}

		ts = append(ts, t)
	}

	return ts
}

// findContentIndex find index of matching title to get the index range of the title content.
func findContentIndex(sa *suffixarray.Index, titles []string) []int {
	cir := make([]int, len(titles))
	for i, tt := range titles {
		re := regexp.MustCompile(fmt.Sprintf(`%s`, tt))
		// Only get the 1st 2 match and assumed the first one was from table of contents and
		// 2nd from the start of the content which is normally starts with the title.
		idxs := sa.FindAllIndex(re, 2)
		if len(idxs) < 2 {
			cir[i] = -1
			continue
		}

		// Content index range starts from detected title and next to
		cir[i] = idxs[1][0]
	}

	return cir
}
