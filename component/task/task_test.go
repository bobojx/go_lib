package task

import (
	"fmt"
	"testing"
	"time"
)

func TestManage_Add(t *testing.T) {

	m := NewManage()

	// 每30分钟执行
	m.AddTask("* */30 * * * *", func(item *Item) bool {
		fmt.Println("begin execute function")
		return true
	}, nil, nil)

	// 每天1点整执行
	m.AddTask("0 0 0 1 * *", func(item *Item) bool {
		fmt.Println("begin execute function")
		return true
	}, nil, nil)

	// 2029年1月2日 15:14:40执行任务
	m.AddTask("40 14 15 2 1 2029", func(item *Item) bool {
		fmt.Println("begin execute function")
		return true
	}, nil, nil)

	//m.run2Sleep()
	m.run2Tick()
}

func TestManage_Add2(t *testing.T) {
	now := time.Now()
	fmt.Println(now.Add(time.Hour*19 + time.Minute + 57))
	//next := now.Add(time.Hour * 24)
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	fmt.Println(next.Sub(now))
}
