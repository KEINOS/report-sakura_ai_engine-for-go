package topk

import (
	"container/heap"
)

// wordFreq holds a word and its frequency
type wordFreq struct {
	word  string
	count int
}

// freqHeap implements heap.Interface for wordFreq items
// Orders by ascending frequency, then descending lexicographical order
type freqHeap []*wordFreq

func (h freqHeap) Len() int { return len(h) }

func (h freqHeap) Less(i, j int) bool {
	if h[i].count == h[j].count {
		return h[i].word > h[j].word // reverse for min-heap based on lex order
	}
	return h[i].count < h[j].count // min-heap based on count
}

func (h freqHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *freqHeap) Push(x interface{}) {
	*h = append(*h, x.(*wordFreq))
}

func (h *freqHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// TopKFrequent returns at most k unique words ordered by descending frequency,
// with ties broken by ascending lexical order.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	// Count frequencies
	freqMap := make(map[string]int, len(words))
	for _, word := range words {
		freqMap[word]++
	}

	// Use a min-heap of size k to track top k elements
	h := &freqHeap{}
	heap.Init(h)

	for word, count := range freqMap {
		heap.Push(h, &wordFreq{word, count})
		if h.Len() > k {
			heap.Pop(h)
		}
	}

	// Extract results in correct order
	result := make([]string, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(*wordFreq).word
	}

	return result
}
