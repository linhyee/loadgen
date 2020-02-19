package lib

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestGoTickets(t *testing.T) {
	total := 10
	gotickets, err := NewGoTickets(uint32(total))
	if err != nil {
		fmt.Println(err)
		return
	}
	//消费者
	go func(t GoTickets) {
		for {
			//出现的个数也是随机的
			for i := 0; i < rand.Intn(10); i++ {
				go func(t GoTickets) {
					//取票
					t.Take()
					//窗口处理时长
					time.Sleep(1 * time.Second)
					//归还票
					t.Return()
				}(gotickets)
			}
			//假定窗口在3秒范围内随机出现消费者
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
		}
	}(gotickets)
	//业务限定时间
	timer := time.NewTimer(15 * time.Second)
END:
	for {
		select {
		case <-timer.C:
			fmt.Println("[End]")
			break END
		default:
			//每秒播报剩余票数
			time.Sleep(1 * time.Second)
			fmt.Printf("GoTickets status:%v, Total: %d, Remainder:%d\n",
				gotickets.Active(), gotickets.Total(), gotickets.Remainder())
		}
	}
}
