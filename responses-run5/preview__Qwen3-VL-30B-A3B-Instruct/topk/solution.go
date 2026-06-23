package topk

import (
	"container/heap"
	"sort"
)

// TopKFrequent returns the k most frequent unique words, ordered by descending frequency
// and ascending lexical order for ties.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	// Count frequency of each word
	freq := make(map[string]int)
	for _, word := range words {
		freq[word]++
	}

	// Use a min-heap to keep track of top k elements
	// We use negative frequency to simulate max-heap behavior
	// and store (frequency, word) pairs
	h := &minHeap{}
	heap.Init(h)

	for word, count := range freq {
		heap.Push(h, [2]int{-count, len(word), word}) // negative count for max-heap
		if h.Len() > k {
			heap.Pop(h)
		}
	}

	// Extract results from heap and sort by frequency (descending) and word (ascending)
	result := make([]string, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).([2]int)[2]
	}

	// Sort by frequency (descending) and then by word (ascending)
	sort.Slice(result, func(i, j int) bool {
		countI := freq[result[i]]
		countJ := freq[result[j]]
		if countI == countJ {
			return result[i] < result[j]
		}
		return countI > countJ
	})

	return result
}

// minHeap implements heap.Interface for [2]int where the first element is the key
// We use negative frequency to simulate max-heap behavior
type minHeap [][3]int // [negativeFrequency, wordLength, word]

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { // Min-heap based on negative frequency
	if h[i][0] != h[j][0] {
		return h[i][0] < h[j][0]
	}
	if h[i][1] != h[j][1] {
		return h[i][1] < h[j][1]
	}
	return h[i][2] < h[j][2]
}
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.([3]int)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
