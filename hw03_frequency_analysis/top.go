package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type wordCount struct {
	word  string
	count int
}

func Top10(str string) []string {
	words := strings.Fields(str)

	m := make(map[string]int)
	for _, word := range words {
		m[word]++
	}

	wc := make([]wordCount, 0, len(m))
	for word, count := range m {
		wc = append(wc, wordCount{word, count})
	}

	sort.Slice(wc, func(i, j int) bool {
		if wc[i].count != wc[j].count {
			return wc[i].count > wc[j].count
		}
		return wc[i].word < wc[j].word
	})

	result := make([]string, 0, 10)
	for i := 0; i < len(wc) && i < 10; i++ {
		result = append(result, wc[i].word)
	}

	return result
}
