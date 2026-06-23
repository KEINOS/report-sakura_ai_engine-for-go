package topk

import "container/heap"

type wordFreq struct {
	word string
	freq int
}

// minHeap orders candidates so that the least desirable is at the root:
// lower frequency first, and for equal frequency larger word first.
type minHeap []wordFreq

func (h minHeap) Len() int { return len(h) }
func (h minHeap) Less(i, j int) bool {
	if h[i].freq == h[j].freq {
		return h[i].word > h[j].word
	}
	return h[i].freq < h[j].freq
}
func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x any) {
	*h = append(*h, x.(wordFreq))
}

func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// TopKFrequent returns at most k unique words from words ordered by
// descending frequency. Ties are broken by ascending lexical order.
func TopKFrequent(words []string, k int) []string {
	if k <= 0 || len(words) == 0 {
		return []string{}
	}

	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++
	}

	h := &minHeap{}
	heap.Init(h)

	for w, f := range freq {
		if h.Len() < k {
			heap.Push(h, wordFreq{word: w, freq: f})
			continue
		}
		if f > (*h)[0].freq || (f == (*h)[0].freq && w < (*h)[0].word) {
			heap.Pop(h)
			heap.Push(h, wordFreq{word: w, freq: f})
		}
	}

	// Pop yields ascending undesirability, so fill the result backwards
	// to obtain descending frequency and ascending lexical order.
	res := make([]wordFreq, h.Len())
	for i := len(res) - 1; i >= 0; i-- {
		res[i] = heap.Pop(h).(wordFreq)
	}

	out := make([]string, len(res))
	for i, wf := range res {
		out[i] = wf.word
	}
	return out
}
