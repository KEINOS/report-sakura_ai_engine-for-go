package topk

import (
	"container/heap"
	"sort"
)

// TopKFrequent returns at most k unique words from the input slice ordered by
// descending frequency. Words with the same frequency are ordered by ascending
// lexical order. If k <= 0 or words is empty, a non‑nil empty slice is returned.
// The input slice is never modified.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	// Count frequencies.
	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++
	}

	// Min‑heap that keeps the current top‑k entries.
	h := &minHeap{}
	heap.Init(h)

	for w, c := range freq {
		heap.Push(h, entry{word: w, freq: c})
		if h.Len() > k {
			heap.Pop(h) // discard the worst entry
		}
	}

	// Extract heap contents.
	top := make([]entry, h.Len())
	for i := len(top) - 1; i >= 0; i-- {
		top[i] = heap.Pop(h).(entry)
	}

	// Final ordering: descending frequency, then ascending lexical order.
	sort.Slice(top, func(i, j int) bool {
		if top[i].freq != top[j].freq {
			return top[i].freq > top[j].freq
		}
		return top[i].word < top[j].word
	})

	// Build result slice.
	res := make([]string, len(top))
	for i, e := range top {
		res[i] = e.word
	}
	return res
}

// entry stores a word together with its frequency.
type entry struct {
	word string
	freq int
}

// minHeap implements a heap where the "worst" element (smallest frequency,
// and for ties the lexicographically largest word) is at the root.
type minHeap []entry

func (h minHeap) Len() int { return len(h) }

func (h minHeap) Less(i, j int) bool {
	if h[i].freq != h[j].freq {
		return h[i].freq < h[j].freq // smaller frequency is worse
	}
	// For equal frequency, larger word is worse (so it gets popped first).
	return h[i].word > h[j].word
}

func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x any) {
	*h = append(*h, x.(entry))
}

func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
