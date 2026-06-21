package topk

import (
	"sort"
)

func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	// Count the frequency of each word
	freq := make(map[string]int)
	for _, word := range words {
		freq[word]++
	}

	// Sort by frequency descending, then by lex order ascending
	sort.Slice(freq, func(i, j int) bool {
		if freq[i] < freq[j] {
			return true
		} else if freq[i] > freq[j] {
			return false
		}
		return freq[i] == freq[j]
	})

	// Extract words with the top k frequencies
	result := make([]string, k)
	for i := 0; i < k; i++ {
		result[i] = string(freq[result[i]])
	}

	return result
}
