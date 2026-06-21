package topk

import (
	"container/heap"
	"sort"
)

type item struct {
	word string
	freq int
}

// minHeap orders by ascending frequency, and for ties by descending word.
// The root is always the least desirable item, making it easy to evict.
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
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// TopKFrequent returns at most k unique words ordered by descending frequency.
// Equal frequencies are resolved by ascending lexical order.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	freqs := make(map[string]int, len(words))
	for _, w := range words {
		freqs[w]++
	}

	h := &minHeap{}
	heap.Init(h)

	for w, f := range freqs {
		if h.Len() < k {
			heap.Push(h, item{word: w, freq: f})
			continue
		}
		root := (*h)[0]
		if f > root.freq || (f == root.freq && w < root.word) {
			heap.Pop(h)
			heap.Push(h, item{word: w, freq: f})
		}
	}

	items := make([]item, h.Len())
	copy(items, *h)

	sort.Slice(items, func(i, j int) bool {
		if items[i].freq == items[j].freq {
			return items[i].word < items[j].word
		}
		return items[i].freq > items[j].freq
	})

	res := make([]string, len(items))
	for i, it := range items {
		res[i] = it.word
	}
	return res
}
