package lib

import (
	"errors"
	"fmt"
)

// GoTickets 代表Goroutine票池的接口
type GoTickets interface {
	// 拿走一张票
	Take()
	// 归还一张票
	Return()
	// 票池是否已被激活
	Active() bool
	// 票的总数
	Total() uint32
	// 剩余的票数
	Remainder() uint32
}

// myGoTickets 表示Goroutine票池的实现
type myGoTickets struct {
	total    uint32        // 票的总数
	ticketCh chan struct{} // 票的容器
	active   bool          // 票池是否已被激活
}

// NewGoTickets 会新建一个Goroutine票池
func NewGoTickets(total uint32) (GoTickets, error) {
	gt := myGoTickets{}
	if !gt.init(total) {
		errMsg := fmt.Sprintf(
			"The goroutine ticket pool can NOT be initialized!(total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	return &gt, nil
}

// init 初始化Goroutine票池
func (gt *myGoTickets) init(total uint32) bool {
	if gt.active {
		return false
	}
	if total == 0 {
		return false
	}
	ch := make(chan struct{}, total)
	n := int(total)
	for i := 0; i < n; i++ {
		ch <- struct{}{} // 填满票池
	}
	gt.ticketCh = ch
	gt.total = total
	gt.active = true
	return true
}

// Take 拿走一张票
func (gt *myGoTickets) Take() {
	<-gt.ticketCh
}

// Return 归还一张票
func (gt *myGoTickets) Return() {
	gt.ticketCh <- struct{}{}
}

// Active 查看票池是否已激活
func (gt *myGoTickets) Active() bool {
	return gt.active
}

// Total 票池总数
func (gt *myGoTickets) Total() uint32 {
	return gt.total
}

// Remainder 票池剩余票数
func (gt *myGoTickets) Remainder() uint32 {
	return uint32(len(gt.ticketCh))
}
