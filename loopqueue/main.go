package main

import (
	"fmt"
	"github.com/pkg/errors"
)

/*
http://c.biancheng.net/cpp/html/2706.html
为了区分队空还是队满的情况，有三种处理方式：
1) 牺牲一个单元来区分队空和队满，入队时少用一个队列单元，这是一种较为普遍的 做法，约定以“队头指针在队尾指针的下一位置作为队满的标志”，如图3-7(d2)所示。
队满条件为：(Q.rearfl)%MaxSize==Q.front。
队空条件仍为：Q.front==Q.rear。
队列中元素的个数：(Q.rear-Q.front+MaxSize)%MaxSize
2) 类型中增设表示元素个数的数据成员。这样，则队空的条件为Q.Size==0，队满的条 件为 Q.size==MaxSize。这两种情况都有 Q.front=Q.rear。
3) 类型中增设tag数据成员，以区分是队满还是队空。tag等于0的情况下，若因删除导 致Q.front==Q.rear则为队空；tag等于1的情况下，若因插入导致Q.ftont==Q.rear则为队满。
*/

const MaxSize = 5

type queue struct {
	data        []int
	front, rear int
}

var (
	errEmpty = errors.New("队列已空")
	errFull  = errors.New("队列已满")
)

var Q *queue = new(queue)

func init() {
	Q.front = 0
	Q.rear = 0
	Q.data = make([]int, MaxSize)
}

func (q *queue) inQueue(data int) error {
	if (q.rear+1)%MaxSize == q.front {
		return errFull
	}
	q.data[q.rear] = data
	q.rear = (q.rear + 1 + MaxSize) % MaxSize
	return nil
}

func (q *queue) outQueue() (int, error) {
	if (q.rear-q.front+MaxSize)%MaxSize == 0 {
		return 0, errEmpty
	}
	temp := q.data[q.front]
	q.data[q.front] = 0
	q.front = (q.front + 1 + MaxSize) % MaxSize
	return temp, nil

}
func (q *queue) lengthQueue() int {
	return (q.rear - q.front + MaxSize) % MaxSize
}

func main() {
	Q.inQueue(1)
	Q.inQueue(2)
	Q.inQueue(3)
	Q.inQueue(4)

	fmt.Println(Q.outQueue())
	fmt.Println(Q.outQueue())
	fmt.Println(Q.lengthQueue())
	fmt.Println(Q.outQueue())
	fmt.Println(Q.outQueue())
	fmt.Println(Q.outQueue())
	fmt.Println(Q.outQueue())
	Q.inQueue(5)
	fmt.Println(Q.outQueue())

}
