package database

import (
	"fmt"
	"testing"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Sex  string `json:"sex"`
}

func TestInitMysqlDb(t *testing.T) {
	p := Person{Name: "张三", Age: 18, Sex: "男"}
	data, err := ConvertData(&p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
}
