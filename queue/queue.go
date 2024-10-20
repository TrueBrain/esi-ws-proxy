package queue

import (
	"container/heap"
)

type item struct {
	value    any
	priority int
	index    int
}

type priorityQueue []*item

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

type PriorityQueue struct {
	pq priorityQueue
}

func NewPriorityQueue() PriorityQueue {
	return PriorityQueue{
		pq: make(priorityQueue, 0),
	}
}

func (pq *PriorityQueue) Push(value any, priority int) {
	heap.Push(&pq.pq, &item{
		value:    value,
		priority: priority,
	})
}

func (pq *PriorityQueue) Pop() any {
	return heap.Pop(&pq.pq).(*item).value
}

func (pq *PriorityQueue) Peek() any {
	if len(pq.pq) == 0 {
		return nil
	}
	return pq.pq[0].value
}
