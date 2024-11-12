package queue

import "sync"

type LinkQueue struct {
	Root  *LinkNode
	Size  int
	Mutex sync.Mutex
}

type LinkNode struct {
	NextNode *LinkNode
	Value    string
}

func (q *LinkQueue) Add(value string) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	if q.Root == nil {
		root := new(LinkNode)
		root.Value = value

		q.Root = root
	} else {
		newNode := new(LinkNode)
		newNode.Value = value

		nowNode := q.Root
		for nowNode.NextNode != nil {
			nowNode = nowNode.NextNode
		}
		nowNode.NextNode = newNode
	}

	q.Size++

}

func (q *LinkQueue) Remove() string {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	if q.Size == 0 {
		panic("queue is empty")
	}

	res := q.Root.Value

	q.Root = q.Root.NextNode

	return res
}
