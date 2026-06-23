package topk

import (
	"container/heap"
	"sort"
)

// TopKFrequent returns at most k unique words from the input slice ordered by
// descending frequency. Words with the same frequency are ordered by ascending
// lexical order. If k <= 0 or the input slice is empty, a non‑nil empty slice is
// returned. The input slice is not mutated.
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

	for w, f := range freq {
		heap.Push(h, entry{word: w, freq: f})
		if h.Len() > k {
			heap.Pop(h) // discard the worst entry
		}
	}

	// Extract heap contents.
	tmp := make([]entry, h.Len())
	for i := len(tmp) - 1; i >= 0; i-- {
		tmp[i] = heap.Pop(h).(entry)
	}

	// Sort by descending frequency, then ascending lexical order.
	sort.Slice(tmp, func(i, j int) bool {
		if tmp[i].freq != tmp[j].freq {
			return tmp[i].freq > tmp[j].freq
		}
		return tmp[i].word < tmp[j].word
	})

	// Build the result slice.
	out := make([]string, len(tmp))
	for i, e := range tmp {
		out[i] = e.word
	}
	return out
}

// entry stores a word together with its frequency.
type entry struct {
	word string
	freq int
}

// minHeap implements a heap where the "smallest" element is the one with the
// lowest frequency, and for equal frequencies the lexicographically larger
// word is considered smaller (so it gets evicted first).
type minHeap []entry

func (h minHeap) Len() int { return len(h) }

func (h minHeap) Less(i, j int) bool {
	if h[i].freq != h[j].freq {
		return h[i].freq < h[j].freq // lower frequency is smaller
	}
	return h[i].word > h[j].word // larger word is smaller for tie‑break
}

func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(entry))
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
