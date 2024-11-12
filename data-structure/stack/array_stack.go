package stack

import "sync"

// 数组栈
type ArrayStack struct {
	Array []string
	Size  int
	Mutex sync.Mutex
}

// 入栈
func (s *ArrayStack) Push(v string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Array = append(s.Array, v)
	s.Size++
}

// 出栈
func (s *ArrayStack) Pop() string {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	res := s.Array[s.Size-1]

	size := s.Size - 1
	newArray := make([]string, size, size)

	for i := 0; i < size; i++ {
		newArray[i] = s.Array[i]
	}

	s.Array = newArray
	s.Size = size

	return res
}

// 获取栈顶元素
func (s *ArrayStack) Peek() string {
	if s.Size == 0 {
		panic("stack is empty")
	}
	return s.Array[s.Size-1]
}
