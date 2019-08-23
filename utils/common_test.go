package utils

import (
	"fmt"
	"testing"
)

func TestStruct2Map(t *testing.T) {

	var person struct {
		Name string `json:"name"`
		Sex  string `json:"sex"`
		Age  int    `json:"age"`
	}
	p := &person
	p.Name = "张三"
	p.Sex = "男"
	p.Age = 19
	data := Struct2Map(p, nil)
	fmt.Println(data)
	var p2 = person

	_ = Map2Struct(data, p2)
	fmt.Println(p2.Name)
	// 随机字符串
	str := RandString(18)
	fmt.Println(str)

}
