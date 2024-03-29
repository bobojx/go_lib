package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"go_lib/utils"
	"log"
	"reflect"
	"regexp"
	"strings"
)

type SqlDB struct {
	db        *sql.DB
	table     string
	debug     bool
	LastSql   string
	LastArgs  []interface{}
	lastError error
	query     interface{}
	tx        *sql.Tx
}

var SqlDrivers = make(map[string]*sql.DB)

// 数据库配置
type DBConfig struct {
	DBHost     string `json:"db_host" yaml:"db_host"`
	DBPort     string `json:"db_port" yaml:"db_port"`
	DBName     string `json:"db_name" yaml:"db_name"`
	DBUser     string `json:"db_user" yaml:"db_user"`
	DBPassword string `json:"db_password" yaml:"db_password"`
	DBOpenSize int    `json:"db_open_size" yaml:"db_open_size"` // 打开连接数
	DBIdleSize int    `json:"db_idle_size" yaml:"db_idle_size"` // 空闲连接数
	DBDebug    bool   `json:"db_debug" yaml:"db_debug"`
}

// 数据库字段
type DBColumn struct {
	Field string
	Icon  string
}

type DM map[string]interface{}

// 字段验证规则
var columnReg = regexp.MustCompile(`(.+?)\[(\+|-|!|>|<|>=|<=|like)]`)

// 初始化MySQL
func InitMysqlDb(conf *DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", conf.DBUser, conf.DBPassword, conf.DBHost, conf.DBPort, conf.DBName)

	if db, ok := SqlDrivers[dsn]; ok {
		return db, nil
	}
	fmt.Println(dsn)
	mysqlDb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// 设置最大打开连接数
	mysqlDb.SetMaxOpenConns(conf.DBOpenSize)
	// 设置最大空闲连接数
	mysqlDb.SetMaxIdleConns(conf.DBIdleSize)

	err = mysqlDb.Ping()
	if err != nil {
		return nil, err
	}
	SqlDrivers[dsn] = mysqlDb
	return mysqlDb, nil
}

// 实例化MySQL
func NewMysqlDB(conf *DBConfig) (*SqlDB, error) {
	SqlDrivers, err := InitMysqlDb(conf)
	if err != nil {
		return nil, err
	}
	mysql := &SqlDB{db: SqlDrivers, debug: conf.DBDebug}
	return mysql, nil
}

func (m *SqlDB) Table(tableName string) *DBTable {
	return NewDBTable(m, tableName)
}

// 开启事务
func (m *SqlDB) BeginTrans() error {
	var err error
	m.tx, err = m.db.Begin()
	return err
}

// 提交事务
func (m *SqlDB) Commit() error {
	err := m.tx.Commit()
	return err
}

// 回滚事务
func (m *SqlDB) Rollback() error {
	err := m.tx.Rollback()
	return err
}

// 新增
func (m *SqlDB) Insert(table string, orgData interface{}) (int, bool) {
	var columns []string
	var values []interface{}
	var valMask []string
	data, err := ConvertData(orgData)
	if err != nil {
		return 0, false
	}
	for k, v := range data {
		columns = append(columns, m.FormatColumn(k))
		values = append(values, v)
		valMask = append(valMask, "?")
	}
	sqlStr := fmt.Sprintf("INSERT INTO %s(%s) VALUE(%s)", table, strings.Join(columns, ","), strings.Join(valMask, ","))
	res, err := m.Exec(sqlStr)
	if err != nil {
		return 0, false
	}
	id, _ := res.LastInsertId()
	return int(id), false
}

// 删除
func (m *SqlDB) Delete(where utils.M, table string) (int, error) {
	var values []interface{}
	whereStr, whereVal := m.ProcessWhere(where, "AND", table)
	values = append(values, whereVal)
	sqlStr := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereStr)

	res, err := m.Exec(sqlStr, values...)
	if err != nil {
		return 0, err
	}
	// 返回受影响行数
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

// 更新
func (m *SqlDB) Update(data utils.M, where utils.M, table string) error {
	var values []interface{}
	var tmp []string
	for i, v := range data {
		filed := m.explainColumn(i)
		// 自增、自减
		if filed.Icon == "+" || filed.Icon == "-" {
			// name[+] => name = name + ?
			tmp = append(tmp, fmt.Sprintf("%s = %s %s ?", m.FormatColumn(filed.Field), m.FormatColumn(filed.Field), filed.Icon))
		} else {
			tmp = append(tmp, fmt.Sprintf("%s %s ?", m.FormatColumn(filed.Field), filed.Icon))
		}
		values = append(values, v)
	}
	// 组装SQL语句
	sqlStr := fmt.Sprintf("UPDATE %s SET %s", m.FormatColumn(table), strings.Join(tmp, ","))

	if where != nil {
		whereStr, whereVal := m.ProcessWhere(where, "AND", table)
		values = append(values, whereVal)
		sqlStr = fmt.Sprintf("%s WHERE %s", sqlStr, whereStr)
	}

	// 执行SQL
	_, err := m.Exec(sqlStr, values...)
	return err
}

// 查询
func (m *SqlDB) Query(sqlStr string, args ...interface{}) ([]utils.M, error) {
	rows, err := m.db.Query(sqlStr, args...)
	if err != nil {
		if m.debug {
			m.PrintError(err)
		}
		return nil, err
	}
	defer func() {
		_ = rows.Close()

	}()
	return m.FetchAll(rows)
}

func (m *SqlDB) QueryRow(sqlStr string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(sqlStr, args...)
}

// 获取所有数据
func (m *SqlDB) FetchAll(query *sql.Rows) ([]utils.M, error) {
	columns, _ := query.Columns()
	values := make([]interface{}, len(columns))
	scans := make([]interface{}, len(columns))
	for i := range values {
		scans[i] = &values[i]
	}
	results := make([]utils.M, 0)
	for query.Next() {
		if err := query.Scan(scans...); err != nil {
			return nil, err
		}
		row := utils.M{}
		for k, v := range values {
			key := columns[k]
			switch v.(type) {
			case []byte:
				row[key] = string(v.([]byte))
			default:
				row[key] = v
			}
		}
		results = append(results, row)
	}
	return results, nil
}

// 执行SQL
func (m *SqlDB) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	var res sql.Result
	var err error
	if m.tx != nil {
		res, err = m.tx.Exec(sqlStr, args...)
	} else {
		res, err = m.db.Exec(sqlStr, args...)
	}
	m.LastSql = sqlStr
	m.LastArgs = args
	if err != nil {
		if m.debug {
			m.PrintError(err)
		}
		return nil, err
	}
	return res, err
}

// 打印错误
func (m *SqlDB) PrintError(err error) {
	fmt.Println(m.LastSql)
	fmt.Println(m.LastArgs)
	fmt.Println(err)
	m.lastError = err
}

// 获取最后一条错误
func (m *SqlDB) GetLastError() error {
	return m.lastError
}

// 处理where条件
func (m *SqlDB) ProcessWhere(where utils.M, icon string, table string) (string, []interface{}) {
	var whereStrings []string
	var values []interface{}
	for i, v := range where {
		if i == "AND" || i == "OR" {
			// 递归获取所有查询条件
			tmpWhere, val := m.ProcessWhere(v.(utils.M), i, table)
			whereStrings = append(whereStrings, tmpWhere)
			values = append(values, val...)
		} else {
			t := reflect.TypeOf(v).Kind()
			if t == reflect.Slice || t == reflect.Array {
				values = append(values, v.([]interface{})...)
				whereStrings = append(whereStrings, m.formatWhere(i, table, len(v.([]interface{}))))
			} else {
				values = append(values, v)
				whereStrings = append(whereStrings, m.formatWhere(i, table, 0))
			}
		}
	}
	wherePrefix := fmt.Sprintf(" %s ", icon)
	whereStr := fmt.Sprintf("(%s)", strings.Join(whereStrings, wherePrefix))
	return whereStr, values
}

// 格式化where条件
func (m *SqlDB) formatWhere(column string, table string, length int) string {
	filed := m.explainColumn(column)

	columnStr := m.FormatColumn(filed.Field)
	icon := filed.Icon
	var whereIcon string

	var formatStr string
	// 处理多个值的情况
	if length > 0 {
		var maskArgs []string
		whereIcon = "IN"
		for i := 0; i < length; i++ {
			maskArgs = append(maskArgs, "?")
		}
		if icon == "!" {
			whereIcon = "NOT IN"
		}
		formatStr = fmt.Sprintf("%s.%s %s (%s)", m.FormatColumn(table), columnStr, whereIcon, strings.Join(maskArgs, ","))
	} else {
		if icon == "!" {
			whereIcon = "!="
		} else {
			whereIcon = icon
		}
		formatStr = fmt.Sprintf("%s.%s %v ?", m.FormatColumn(table), columnStr, whereIcon)
	}
	return formatStr
}

// 解析字段
func (m *SqlDB) explainColumn(column string) *DBColumn {
	match := columnReg.FindStringSubmatch(column)
	filed := &DBColumn{}
	if len(match) > 0 {
		filed.Field = match[1]
		filed.Icon = match[2]
	} else {
		filed.Field = column
		filed.Icon = "="
	}
	return filed
}

// 格式化字段
func (m *SqlDB) FormatColumn(column string) string {
	return fmt.Sprintf("`%s`", column)
}

// 扫描数据到map中
func (m *SqlDB) scan2Map(scans []interface{}, columns []string) interface{} {
	obj := utils.M{}
	for i, v := range columns {
		var val interface{}
		obj[v] = val
		scans[i] = &val
	}
	return obj
}

// 扫描数据到结构体中
func (m *SqlDB) scan2Struct(t reflect.Type, scans []interface{}, columns []string) interface{} {
	obj := reflect.New(t).Interface()
	objVal := reflect.ValueOf(obj).Elem()
	for i, c := range columns {
		index := m.findTagOf(t, c)
		if index != -1 {
			scans[i] = objVal.Field(index).Addr().Interface()
		} else {
			var empty interface{}
			scans[i] = &empty
		}
	}
	return obj
}

// 扫描数据到任何类型中
func (m *SqlDB) scan2Any(scans []interface{}, columns []string, i interface{}) interface{} {
	if i == nil {
		return m.scan2Map(scans, columns)
	}
	t := reflect.TypeOf(i)
	switch t.Kind() {
	case reflect.Ptr:
		return m.scan2Struct(t.Elem(), scans, columns)
	case reflect.Struct:
		return m.scan2Struct(t, scans, columns)
	case reflect.Map:
		fallthrough
	default:
		return m.scan2Map(scans, columns)
	}
}

// 查找tag下标
func (m *SqlDB) findTagOf(t reflect.Type, colName string) int {
	for i := 0; i < t.NumField(); i++ {
		val, ok := t.Field(i).Tag.Lookup("json")
		if ok && val == colName {
			return i
		}
	}
	return -1
}

// 关闭数据库
func (m *SqlDB) Close() {
	err := m.db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// 处理数据
func ConvertData(orgData interface{}) (DM, error) {
	t := reflect.TypeOf(orgData)
	switch t.Kind() {
	case reflect.Map:
		if t.Name() == "DM" {
			return orgData.(DM), nil
		} else if t.Name() == "M" {
			return DM(orgData.(utils.M)), nil
		}
		return orgData.(map[string]interface{}), nil
	case reflect.Ptr:
		fallthrough
	case reflect.Struct:
		return utils.Struct2Map(orgData, nil), nil
	default:
		return nil, errors.New("not support this data")
	}
}

// 构建数据表结构体
func BuildTableStruct(tabName, dbName string, dbConf *DBConfig) {
	types := map[string]string{
		"int":      "int",
		"tinyint":  "int",
		"smallint": "int",
		"varchar":  "string",
		"char":     "string",
		"text":     "string",
		"tinytext": "string",
		"double":   "float64",
		"float":    "float64",
	}

	db, err := NewMysqlDB(dbConf)
	if err != nil {
		panic(err)
	}
	res, err := db.Table("COLUMNS").Where(utils.M{"TABLE_NAME": tabName, "TABLE_SCHEMA": dbName}, "").Order(utils.M{"ORDINAL_POSITION": "ASC"}).Query().Result()
	if err != nil {
		panic(err)
	}

	var tmp []string

	var (
		colName    string
		colType    string
		colComment string
	)

	for _, row := range res {
		colName = row["COLUMN_NAME"].(string)
		colType = row["DATA_TYPE"].(string)
		colComment = row["COLUMN_COMMENT"].(string)
		tmp = append(tmp, fmt.Sprintf("\t%v %v `json:\"%v\" bson:\"%v\"` // %v", utils.Under2Hump(colName), types[colType], colName, colName, colComment))
	}

	fmt.Println(fmt.Sprintf("type %s struct {", tabName))
	for _, v := range tmp {
		fmt.Println(v)
	}
	fmt.Println("}")
}
