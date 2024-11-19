package main

import (
	"fmt"
	"sync"
)

// 二叉树
type TreeNode struct {
	Data      string    //数据
	LeftNode  *TreeNode //左子树
	RightNode *TreeNode //右子树
}

type LinkNode struct {
	Next  *LinkNode
	Value *TreeNode
}
type LinkQueue struct {
	root *LinkNode
	size int
	lock sync.Mutex
}

// 层次遍历 使用广度遍历的方法
func LayerOrder(treeNode *TreeNode) {
	if treeNode == nil {
		return
	}

	queue := new(LinkQueue)

	//跟节点先入队
	queue.Add(treeNode)

	for queue.size > 0 {
		element := queue.Remove()

		//先打印节点值
		fmt.Println(element.Data)

		//左子树非空，入队列
		if element.LeftNode != nil {
			queue.Add(element.LeftNode)
		}

		//右子树非空入队列
		if element.RightNode != nil {
			queue.Add(element.RightNode)
		}
	}
}

func (queue *LinkQueue) Add(node *TreeNode) {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	//如果栈顶为空，那么增加节点
	if queue.root == nil {
		queue.root = new(LinkNode)
		queue.root.Value = node
	} else {
		//否则新元素插入链表的末尾
		newNode := new(LinkNode)
		newNode.Value = node

		//一直遍历到链表的尾部
		nowNode := queue.root
		for nowNode.Next != nil {
			nowNode = nowNode.Next
		}

		//新节点放在链表尾部
		nowNode.Next = newNode
	}

	queue.size += 1
}

// 出队
func (queue *LinkQueue) Remove() *TreeNode {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	//队中元素已空
	if queue.size == 0 {
		return nil
	}

	//顶部元素出队
	topNode := queue.root
	v := topNode.Value

	//将顶部元素的后继链接上
	queue.root = topNode.Next

	//队列中元素-1
	queue.size -= 1
	return v
}

func (queue *LinkQueue) Size() int {
	return queue.size
}

func main() {
	t := &TreeNode{Data: "A"}
	t.LeftNode = &TreeNode{Data: "B"}
	t.RightNode = &TreeNode{Data: "C"}
	t.LeftNode.LeftNode = &TreeNode{Data: "D"}
	t.LeftNode.RightNode = &TreeNode{Data: "E"}
	t.RightNode.LeftNode = &TreeNode{Data: "F"}

	fmt.Println("\n层次排序")
	LayerOrder(t)
}

//// 先序遍历
//func PreOrder(tree *TreeNode) {
//	if tree == nil {
//		return
//	}
//	//先打印根节点
//	fmt.Println(tree.Data)
//	//在打印左子树
//	PreOrder(tree.LeftNode)
//	//再打印右子树
//	PreOrder(tree.RightNode)
//}
//
//// 中序遍历
//func MidOrder(tree *TreeNode) {
//	if tree == nil {
//		return
//	}
//	//先打印左子树
//	MidOrder(tree.LeftNode)
//	//再打印根节点
//	fmt.Println(tree.Data)
//	//再打印右子树
//	MidOrder(tree.RightNode)
//}
//
//// 后序遍历
//func PostOrder(tree *TreeNode) {
//	if tree == nil {
//		return
//	}
//
//	// 先打印左子树
//	PostOrder(tree.LeftNode)
//	// 再打印右字树
//	PostOrder(tree.RightNode)
//	// 再打印根节点
//	fmt.Print(tree.Data, " ")
//}
