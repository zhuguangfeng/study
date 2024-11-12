package main

type DoubleLinkList struct {
	NextNode *DoubleLinkList
	PrevNode *DoubleLinkList
	Data     int
}

func NewDoubleLinkList() *DoubleLinkList {
	return &DoubleLinkList{}
}

// 获取上一个节点
func (l *DoubleLinkList) Prev() *DoubleLinkList {
	p := l.PrevNode
	if p != nil && p != l {
		return p
	}
	return nil
}

// 获取下一个节点
func (l *DoubleLinkList) Next() *DoubleLinkList {
	n := l.NextNode
	if n != nil && n != l {
		return n
	}
	return nil
}
