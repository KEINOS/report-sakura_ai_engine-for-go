package topk

import "container/heap"

type wordCount struct {
	word  string
	count int
}

type minHeap []wordCount

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool {
	if h[i].count != h[j].count {
		return h[i].count < h[j].count
	}
	return h[i].word > h[j].word
}
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(wordCount)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	counts := make(map[string]int)
	for _, w := range words {
		counts[w]++
	}

	h := &minHeap{}
	heap.Init(h)

	for w, c := range counts {
		heap.Push(h, wordCount{w, c})
		if h.Len() > k {
			heap.Pop(h)
		}
	}

	res := make([]string, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		res[i] = heap.Pop(h).(wordCount).word
	}
	return res
}
