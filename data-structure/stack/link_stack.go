package stack

import "sync"

// 链表栈
type LinkStack struct {
	Root  *LinkNode
	Size  int
	Mutex sync.Mutex
}

type LinkNode struct {
	Next  *LinkNode
	Value string
}

// 出栈
func (s *LinkStack) Push(v string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Root == nil {
		root := new(LinkNode)
		root.Value = v
		s.Root = root
	} else {
		prevNode := s.Root

		root := new(LinkNode)
		root.Value = v

		root.Next = prevNode.Next
		s.Root = root
	}
	s.Size++

}

func (s *LinkStack) Pop() string {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Root == nil {
		panic("stack is empty")
	}

	res := s.Root.Value

	s.Root = s.Root.Next

	s.Size--

	return res
}

// 获取栈顶元素
func (s *LinkStack) Top() string {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Root == nil {
		panic("stack is empty")
	}

	return s.Root.Value
}
