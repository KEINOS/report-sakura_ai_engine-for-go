package topk

import (
	"container/heap"
	"sort"
)

// TopKFrequent returns the k most frequent words, ordered by frequency (desc)
// and then lexicographically (asc) for ties.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	// Count frequencies
	freq := make(map[string]int)
	for _, word := range words {
		freq[word]++
	}

	// Use a min-heap to keep track of top k elements
	h := &minHeap{}
	heap.Init(h)

	for word, count := range freq {
		if h.Len() < k {
			heap.Push(h, item{word, count})
		} else if count > h.Peek().count || (count == h.Peek().count && word < h.Peek().word) {
			heap.Pop(h)
			heap.Push(h, item{word, count})
		}
	}

	// Extract results and reverse to get descending order
	result := make([]string, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(item).word
	}

	return result
}

// item represents a word and its frequency
type item struct {
	word  string
	count int
}

// minHeap implements heap.Interface with min-heap behavior
// Items are ordered by count (ascending), then by word (ascending)
type minHeap struct {
	items []item
}

func (h *minHeap) Len() int           { return len(h.items) }
func (h *minHeap) Less(i, j int) bool { return h.items[i].count < h.items[j].count || (h.items[i].count == h.items[j].count && h.items[i].word > h.items[j].word) }
func (h *minHeap) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *minHeap) Push(x interface{}) {
	h.items = append(h.items, x.(item))
}

func (h *minHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	x := old[n-1]
	h.items = old[0 : n-1]
	return x
}

func (h *minHeap) Peek() item {
	return h.items[0]
}
