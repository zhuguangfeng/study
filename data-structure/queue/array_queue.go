package queue

import "sync"

// 数组队列 先进先出
type ArrayQueue struct {
	Array []string
	Size  int
	Mutex sync.Mutex
}

// 入列
func (q *ArrayQueue) Add(element string) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	q.Array = append(q.Array, element)
	q.Size++
}

// 出列
func (q *ArrayQueue) Remove() string {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	if q.Size == 0 {
		panic("Queue is empty")
	}

	res := q.Array[0]

	// 直接原位移动
	//for i := 1; i < len(q.Array); i++ {
	//	q.Array[i-1] = q.Array[i]
	//}
	//q.Array = q.Array[0 : len(q.Array)-1]

	//创建新的数组
	newArray := make([]string, q.Size-1)
	for i := 1; i < q.Size-1; i++ {
		newArray[i-1] = q.Array[i]
	}
	q.Array = newArray
	q.Size--
	return res
}
