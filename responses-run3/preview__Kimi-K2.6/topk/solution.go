package topk

import (
	"container/heap"
	"sort"
)

type entry struct {
	word string
	freq int
}

type minHeap []entry

func (h minHeap) Len() int { return len(h) }
func (h minHeap) Less(i, j int) bool {
	if h[i].freq == h[j].freq {
		return h[i].word > h[j].word
	}
	return h[i].freq < h[j].freq
}
func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(entry))
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	e := old[n-1]
	*h = old[:n-1]
	return e
}

// TopKFrequent returns at most k unique words from words ordered by descending
// frequency. Words with equal frequency are ordered by ascending lexical order.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	freqs := make(map[string]int, len(words))
	for _, w := range words {
		freqs[w]++
	}

	unique := len(freqs)
	if k > unique {
		k = unique
	}

	entries := make([]entry, 0, k)

	if k == unique {
		for w, f := range freqs {
			entries = append(entries, entry{word: w, freq: f})
		}
	} else {
		h := &minHeap{}
		heap.Init(h)

		for w, f := range freqs {
			if h.Len() < k {
				heap.Push(h, entry{word: w, freq: f})
				continue
			}
			top := (*h)[0]
			if f > top.freq || (f == top.freq && w < top.word) {
				heap.Pop(h)
				heap.Push(h, entry{word: w, freq: f})
			}
		}

		for h.Len() > 0 {
			entries = append(entries, heap.Pop(h).(entry))
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].freq == entries[j].freq {
			return entries[i].word < entries[j].word
		}
		return entries[i].freq > entries[j].freq
	})

	result := make([]string, len(entries))
	for i, e := range entries {
		result[i] = e.word
	}
	return result
}
