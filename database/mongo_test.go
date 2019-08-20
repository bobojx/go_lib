package database

import (
	"fmt"
	"testing"
)

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
