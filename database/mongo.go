package database

import (
	"fmt"
	"gopkg.in/mgo.v2"
)

type MongoDBConfig struct {
	DBHost     string `json:"db_host" yaml:"db_host"`
	DBPort     string `json:"db_port" yaml:"db_port"`
	DBUser     string `json:"db_user" yaml:"db_user"`
	DBPassword string `json:"db_password" yaml:"db_password"`
	DBAuth     string `json:"db_auth" yaml:"db_auth"`
	DBName     string `json:"db_name" yaml:"db_name"`
	DBPoolSize int    `json:"db_pool_size" yaml:"db_pool_size"`
}

type MongoDB struct {
	isOpen  bool
	dbName  string
	session *mgo.Session
	db      *mgo.Database
}

// 全局缓存
var globalSession *mgo.Session

// 初始化数据库
func InitMongo(conf *MongoDBConfig) error {
	var err error
	fmt.Println(conf.BuildDsn())
	globalSession, err = mgo.Dial(conf.BuildDsn())
	if err != nil {
		return err
	}
	// 设置单个服务器中使用的最大套接字数
	globalSession.SetPoolLimit(conf.DBPoolSize)
	// 限制与使用给定标记配置的服务器的通信
	globalSession.SelectServers()
	return nil
}

// 创建连接配置
func (mc *MongoDBConfig) BuildDsn() string {
	if mc.DBAuth == "" {
		return fmt.Sprintf("mongodb://%s:%s", mc.DBHost, mc.DBPort)
	}
	return fmt.Sprintf("mongodb://%s:%s@%s:%s?authSource=%s", mc.DBUser, mc.DBPassword, mc.DBHost, mc.DBPort, mc.DBAuth)
}

// 创建一个连接
func NewMongoDB(dbName string) *MongoDB {
	return &MongoDB{dbName: dbName}
}

// 打开一个连接
func (m *MongoDB) Open(dbName string) *MongoDB {
	if !m.isOpen {
		m.session = globalSession.Clone()
		m.isOpen = true
		m.db = m.session.DB(dbName)
	} else {
		m.db = m.session.DB(dbName)
	}
	return m
}

// 获取table
func (m *MongoDB) Table(tabName string) *mgo.Collection {
	if !m.isOpen {
		m.Open(m.dbName)
	}
	return m.db.C(tabName)
}

// 关闭连接
func (m *MongoDB) Close() {
	if m.isOpen {
		m.isOpen = false
		m.session.Close()
	}
}
