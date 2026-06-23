package topk

import (
	"container/heap"
	"sort"
)

// TopKFrequent returns at most k unique words from the input slice ordered by
// descending frequency. Words with the same frequency are ordered by ascending
// lexical order. If k <= 0 or the input slice is empty, a non‑nil empty slice is
// returned. The input slice is never mutated.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return make([]string, 0)
	}

	// Count frequencies.
	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++
	}

	// Min‑heap that keeps the current k best items.
	h := &minHeap{}
	heap.Init(h)

	// Helper to decide if a is better than b.
	better := func(a, b item) bool {
		if a.freq != b.freq {
			return a.freq > b.freq // higher frequency is better
		}
		return a.word < b.word // lexicographically smaller is better
	}

	// Populate the heap with at most k elements.
	for w, f := range freq {
		it := item{word: w, freq: f}
		if h.Len() < k {
			heap.Push(h, it)
			continue
		}
		if better(it, (*h)[0]) {
			(*h)[0] = it
			heap.Fix(h, 0)
		}
	}

	// Extract heap contents.
	resItems := make([]item, h.Len())
	for i := range resItems {
		resItems[i] = heap.Pop(h).(item)
	}

	// Sort by descending frequency, then ascending lexical order.
	sort.Slice(resItems, func(i, j int) bool {
		if resItems[i].freq != resItems[j].freq {
			return resItems[i].freq > resItems[j].freq
		}
		return resItems[i].word < resItems[j].word
	})

	// Build the result slice.
	out := make([]string, len(resItems))
	for i, it := range resItems {
		out[i] = it.word
	}
	return out
}

// item stores a word together with its frequency.
type item struct {
	word string
	freq int
}

// minHeap implements a min‑heap where the "worst" element (lowest frequency,
// and for ties the lexicographically largest word) is at the root.
type minHeap []item

func (h minHeap) Len() int { return len(h) }

func (h minHeap) Less(i, j int) bool {
	if h[i].freq != h[j].freq {
		return h[i].freq < h[j].freq // lower frequency is "less"
	}
	return h[i].word > h[j].word // larger word is "less" (worse)
}

func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(item))
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	it := old[n-1]
	*h = old[:n-1]
	return it
}
