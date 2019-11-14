package component

import (
	"fmt"
	"testing"
)

func TestNewPool(t *testing.T) {
	total := 0
	p := NewPool(10, func(obj ...interface{}) bool {
		return sum(obj[0].(int), &total)
	})
	var num []interface{}
	for i := 1; i <= 100; i++ {
		num = append(num, i)
	}
	p.AddTaskInterface(num)
	p.FinishCallBack(func() {
		fmt.Println(total)
	})
	p.Start()
}

func sum(obj int, total *int) bool {
	*total += obj
	return true
}
