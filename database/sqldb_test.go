package database

import (
	"testing"
)

var dbConf = &DBConfig{
	DBHost:     "168.168.0.10",
	DBPort:     "3306",
	DBName:     "information_schema",
	DBUser:     "root",
	DBPassword: "kKie93jgUrn!k",
	DBOpenSize: 200,
	DBIdleSize: 100,
	DBDebug:    true,
}

// 测试生成数据表结构体
func TestBuildTableStruct(t *testing.T) {

	BuildTableStruct("t_agent", "pcbx_life_insurance", dbConf)
}
