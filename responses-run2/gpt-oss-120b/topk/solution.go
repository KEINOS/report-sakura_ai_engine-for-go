package topk

import (
	"container/heap"
	"sort"
)

// TopKFrequent returns at most k unique words from the input slice ordered by
// descending frequency. Words with the same frequency are ordered by ascending
// lexical order. If k <= 0 or the input slice is empty, a non‑nil empty slice is
// returned. The input slice is not modified.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	// Count frequencies.
	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++
	}

	// Min‑heap that keeps the k best entries.
	h := &minHeap{}
	heap.Init(h)

	for w, f := range freq {
		e := entry{word: w, freq: f}
		if h.Len() < k {
			heap.Push(h, e)
			continue
		}
		// If e is better than the worst element (heap root), replace it.
		if better(e, (*h)[0]) {
			heap.Pop(h)
			heap.Push(h, e)
		}
	}

	// Extract heap contents.
	res := make([]entry, h.Len())
	for i := len(res) - 1; i >= 0; i-- {
		res[i] = heap.Pop(h).(entry)
	}

	// Final ordering: descending frequency, then ascending lexical order.
	sort.Slice(res, func(i, j int) bool {
		if res[i].freq != res[j].freq {
			return res[i].freq > res[j].freq
		}
		return res[i].word < res[j].word
	})

	// Build the output slice.
	out := make([]string, len(res))
	for i, e := range res {
		out[i] = e.word
	}
	return out
}

// entry stores a word together with its frequency.
type entry struct {
	word string
	freq int
}

// better reports whether a should be preferred over b according to the
// problem ordering (higher frequency, then smaller lexical order).
func better(a, b entry) bool {
	if a.freq != b.freq {
		return a.freq > b.freq
	}
	return a.word < b.word
}

// minHeap implements a min‑heap where the "worst" element (lowest frequency,
// and for ties the lexicographically larger word) is at the root.
type minHeap []entry

func (h minHeap) Len() int { return len(h) }

func (h minHeap) Less(i, j int) bool {
	if h[i].freq != h[j].freq {
		return h[i].freq < h[j].freq // lower frequency is "less"
	}
	return h[i].word > h[j].word // larger word is "less" for tie‑break
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
