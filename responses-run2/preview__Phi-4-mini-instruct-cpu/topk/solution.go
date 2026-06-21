package topk

import (
	"sort"
)

func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[word]++
	}

	type kv struct {
		word  string
		count int
	}

	var items []kv
	for word, count := range wordCount {
		items = append(items, kv{word, count})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].count == items[j].count {
			return items[i].word < items[j].word
		}
		return items[i].count > items[j].count
	})

	result := make([]string, 0, k)
	for i := 0; i < k && i < len(items); i++ {
		result = append(result, items[i].word)
	}

	return result
}
