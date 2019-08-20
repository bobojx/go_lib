package database

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"go_lib/utils"
	"log"
	"os"
	"path"
	"time"
)

type BoltDB struct {
	db *bolt.DB
}

// 新建boltDB
func NewBoltDB(filepath string, mode os.FileMode, options *bolt.Options) *BoltDB {
	if !utils.PathExist(path.Dir(filepath)) {
		err := os.MkdirAll(path.Dir(filepath), 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	if mode == 0 {
		mode = 0600
	}

	if options == nil {
		options = &bolt.Options{Timeout: 1 * time.Second}
	}
	db, err := bolt.Open(filepath, mode, options)
	if err != nil {
		log.Fatal(err)
	}
	return &BoltDB{db: db}
}

// 数据库路径
func (b *BoltDB) Path() string {
	return b.db.Path()
}

// 关闭数据库
func (b *BoltDB) Close() error {
	return b.db.Close()
}

// 获取数据
func (b *BoltDB) Get(bucketName string, key string) ([]byte, error) {
	var val []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		btk := tx.Bucket([]byte(bucketName))
		if btk == nil {
			return nil
		}
		val = btk.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}

// 插入数据
func (b *BoltDB) Put(bucketName string, key string, val string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		btk, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		data, err := json.Marshal(val)
		if err != nil {
			return err
		}
		err = btk.Put([]byte(key), data)
		return err
	})
	return err
}

// 删除数据
func (b *BoltDB) Delete(bucketName string, key string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		btk, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		err = btk.Delete([]byte(key))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// 遍历一个bucket中的数据
func (b *BoltDB) ForEach(bucketName string, callback func([]byte, []byte) error) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		btk := tx.Bucket([]byte(bucketName))
		if btk == nil {
			return nil
		}
		return btk.ForEach(callback)
	})
	return err
}
