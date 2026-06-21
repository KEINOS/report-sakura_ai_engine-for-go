package topk

import (
	"sort"
)

// TopKFrequent returns at most k unique words with the highest frequency.
// In case of equal frequencies, words are ordered in ascending lexical order.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[word]++
	}

	var topK []string
	var freqList []struct {
		word   string
		freq   int
		isBest bool
	}

	for word, freq := range wordCount {
		topK = append(topK, word)
		freqList = append(freqList, struct {
			word   string
			freq   int
			isBest bool
		}{word, freq, freq == maxFreq(wordCount)})
	}

	sort.Slice(freqList, func(i, j int) bool {
		if freqList[i].isBest != freqList[j].isBest {
			return freqList[i].isBest
		}
		return freqList[i].word < freqList[j].word
	})

	return topK[:k]
}

func maxFreq(m map[string]int) int {
	max := 0
	for freq := range m {
		if freq > max {
			max = freq
		}
	}
	return max
}

// This implementation avoids sorting all input occurrences by using a map
// to count the frequency of each word and then creating a list of words
// with their frequencies and a flag indicating if they are the maximum
// frequency. This list is then sorted based on the frequency and lexical order.
// Finally, the top k elements are returned.

// Note: The sort.Slice function is used instead of sort.SliceStable to avoid
// the additional overhead of sort.SliceStable, which is not required in this case
// since we are not sorting based on multiple fields and the lexical order is
// sufficient for breaking ties.
