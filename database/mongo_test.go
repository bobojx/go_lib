package database

import (
	"fmt"
	"go_lib/utils"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

type User struct {
	Name  string `json:"name" bson:"name"`
	Sex   string `json:"sex" bson:"sex"`
	Age   int    `json:"age" bson:"age"`
	Phone string `json:"phone" bson:"phone"`
}

func init() {
	mc := &MongoDBConfig{
		DBHost:     "127.0.0.1",
		DBPort:     "27017",
		DBAuth:     "",
		DBUser:     "",
		DBPassword: "",
		DBPoolSize: 10,
	}
	err := InitMongo(mc)
	if err != nil {
		fmt.Println(err)
	}
}

func TestInitMongo(t *testing.T) {
	mc := &MongoDBConfig{
		DBHost:     "127.0.0.1",
		DBPort:     "27017",
		DBAuth:     "",
		DBUser:     "",
		DBPassword: "",
		DBPoolSize: 10,
	}
	err := InitMongo(mc)
	if err != nil {
		fmt.Println(err)
	}
}

func TestNewCollection(t *testing.T) {
	m := NewMongoDB("t_user")
	defer m.Close()
	c := NewCollection(m, "t_user")

	u := User{
		Name:  "麻子",
		Sex:   "男",
		Age:   24,
		Phone: "18888888888",
	}
	// 插入操作
	err := c.Insert(u)
	if err != nil {
		fmt.Println(err)
	}
}

func TestCollection_Find(t *testing.T) {
	m := NewMongoDB("t_user")
	defer m.Close()
	c := NewCollection(m, "t_user")

	result := &User{}
	err := c.Find(bson.M{"name": "张三"}, result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)

	// 查询所有
	list, err := c.Query(bson.M{}, 1, 0, nil, nil, nil)
	if err != nil {
		fmt.Println(err)
	}
	utils.PrintAny(list)
}
