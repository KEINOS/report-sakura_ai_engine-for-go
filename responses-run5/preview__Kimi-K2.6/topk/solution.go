package topk

import (
	"container/heap"
	"sort"
)

type item struct {
	word string
	freq int
}

// minHeap orders by ascending frequency, then descending word.
// The root is the worst candidate among the current top-k.
type minHeap []item

func (h minHeap) Len() int { return len(h) }
func (h minHeap) Less(i, j int) bool {
	if h[i].freq == h[j].freq {
		return h[i].word > h[j].word
	}
	return h[i].freq < h[j].freq
}
func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x any) {
	*h = append(*h, x.(item))
}

func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	it := old[n-1]
	*h = old[:n-1]
	return it
}

// TopKFrequent returns at most k unique words ordered by descending frequency.
// Ties are broken by ascending lexical order.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++
	}

	if k >= len(freq) {
		unique := make([]string, 0, len(freq))
		for w := range freq {
			unique = append(unique, w)
		}
		sort.Slice(unique, func(i, j int) bool {
			wi, wj := unique[i], unique[j]
			if freq[wi] == freq[wj] {
				return wi < wj
			}
			return freq[wi] > freq[wj]
		})
		return unique
	}

	h := &minHeap{}
	for w, c := range freq {
		if h.Len() < k {
			heap.Push(h, item{word: w, freq: c})
			continue
		}
		if c > (*h)[0].freq || (c == (*h)[0].freq && w < (*h)[0].word) {
			heap.Pop(h)
			heap.Push(h, item{word: w, freq: c})
		}
	}

	candidates := *h
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].freq == candidates[j].freq {
			return candidates[i].word < candidates[j].word
		}
		return candidates[i].freq > candidates[j].freq
	})

	result := make([]string, len(candidates))
	for i, it := range candidates {
		result[i] = it.word
	}
	return result
}
