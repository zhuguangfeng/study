package main

import (
	"fmt"
)

// 单向链表
type SinglyLinkList struct {
	Data     int
	NextNode *SinglyLinkList
}

func NewSinglyLinkList() *SinglyLinkList {
	return &SinglyLinkList{}
}

func main() {
	node := NewSinglyLinkList()
	node.Data = 0

	node1 := NewSinglyLinkList()
	node1.Data = 1

	node.NextNode = node1

	node2 := NewSinglyLinkList()
	node2.Data = 2

	node1.NextNode = node2

	node3 := NewSinglyLinkList()
	node3.Data = 3

	node2.NextNode = node3

	//顺序打印遍历数据
	for node != nil {
		fmt.Println(node.Data)
		node = node.NextNode
	}

	//找出第n个节点

}
