package database

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewBoltDB(t *testing.T) {
	db := NewBoltDB("../data/test.db", 0666, nil)
	err := db.Put("test", "name", "bobojx")
	fmt.Println(err)
	data, err := db.Get("test", "name")
	if err != nil {
		fmt.Println(err)
	}
	var val string
	_ = json.Unmarshal(data, &val)
	fmt.Println(val)

	//err = db.Delete("test", "name")
	//if err != nil {
	//	fmt.Println(err)
	//}
	err = db.ForEach("test", func(bytes []byte, bytes2 []byte) error {
		fmt.Println(bytes, bytes2)
		return nil
	})
}
