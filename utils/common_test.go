package utils

import (
	"fmt"
	"testing"
)

type Person struct {
	Name string `json:"name"`
	Sex  string `json:"sex"`
	Age  int    `json:"age"`
}

func TestStruct2Map(t *testing.T) {
	p := &Person{
		Name: "张三",
		Sex:  "男",
		Age:  19,
	}

	data := Struct2Map(p, nil)
	fmt.Println(data)
	var p2 = &Person{}

	_ = Map2Struct(data, p2)
	fmt.Println(p2.Name)
	// 随机字符串
	str := RandString(18)
	fmt.Println(str)

}
