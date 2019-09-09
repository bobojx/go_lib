package database

import (
	"fmt"
	"go_lib/utils"
	"regexp"
	"strings"
)

// 字段设置
type Field struct {
	Column string // 字段名
	Alias  string // 别名
	Func   string // 所用方法
}

// 表连接设置
type Join struct {
	TableTo    string
	ColumnTo   string
	TableFrom  string
	ColumnFrom string
	Key        string // 链接类型，left,right,online
}

// 表结构设置
type DBTable struct {
	where      utils.M // 条件
	whereStr   string
	joinStr    string
	groupStr   string        // 分组
	fieldStr   string        // 字段
	orderStr   string        // 排序
	limitStr   string        // 分页
	sqlStr     string        // 执行的SQL语句
	table      string        // 表名
	values     []interface{} // 查询值
	db         *SqlDB
	columnType interface{} // 字段类型
}

// 验证字段正则
var fieldReg = regexp.MustCompile(`(.+?)\[(.+?)\]`)

// 新建一个table处理类
func NewDBTable(db *SqlDB, table string) *DBTable {
	return &DBTable{table: table, db: db, fieldStr: "*"}
}

// 开启事务
func (t *DBTable) BeginTrans() error {
	return t.db.BeginTrans()
}

// 提交事务
func (t *DBTable) Commit() error {
	return t.db.Commit()
}

// 回滚事务
func (t *DBTable) Rollback() error {
	return t.db.Rollback()
}

// 解释字段
func (t *DBTable) explainField(field string) *Field {
	match := fieldReg.FindStringSubmatch(field)
	var (
		fieldStr string
		funcStr  string
	)

	if len(match) > 0 {
		fieldStr = match[1]
		funcStr = strings.ToUpper(match[2])
	} else {
		fieldStr = field
	}

	tmp := strings.Split(fieldStr, " ")
	f := &Field{Column: tmp[0], Func: funcStr}
	if len(tmp) > 1 {
		f.Alias = tmp[1]
	}
	return f
}

// 处理查询字段
func (t *DBTable) Select(fields utils.M) *DBTable {
	var tmp []string
	var table string
	for column, tName := range fields {
		if tName.(string) == "" {
			table = t.db.FormatColumn(t.table)
		} else {
			table = t.db.FormatColumn(tName.(string))
		}
		field := t.explainField(column)
		if field.Alias == "" {
			if field.Func == "" {
				tmp = append(tmp, fmt.Sprintf("%s.%s", table, t.db.FormatColumn(field.Column)))
			} else {
				tmp = append(tmp, fmt.Sprintf("%s(%s.%s)", field.Func, table, t.db.FormatColumn(field.Column)))
			}
		} else {
			if field.Func == "" {
				tmp = append(tmp, fmt.Sprintf("%s.%s AS `%s`", table, t.db.FormatColumn(field.Column), field.Alias))
			} else {
				tmp = append(tmp, fmt.Sprintf("%s(%s.%s) AS `%s`", field.Func, table, t.db.FormatColumn(field.Column), field.Alias))
			}
		}
	}
	t.fieldStr = strings.Join(tmp, ",")
	return t
}

// 设置where条件
func (t *DBTable) Where(fields utils.M, table string) *DBTable {
	t.where = fields
	if table == "" {
		table = t.table
	}

	whereStr, val := t.db.ProcessWhere(fields, "AND", table)
	t.whereStr += whereStr
	t.values = append(t.values, val...)
	return t
}

// 设置多个join条件
func (t *DBTable) Join(fields [][]string) *DBTable {
	for _, joinStr := range fields {
		joinTo := strings.Split(joinStr[0], ".")
		joinFrom := strings.Split(joinStr[1], ".")
		t.JoinOne(&Join{
			TableTo:    joinTo[0],
			ColumnTo:   joinTo[1],
			TableFrom:  joinFrom[0],
			ColumnFrom: joinFrom[1],
			Key:        joinStr[2],
		})
	}
	return t
}

// 设置单个join条件
func (t *DBTable) JoinOne(join *Join) *DBTable {
	field := t.explainField(join.TableTo)
	if field.Alias != "" {
		t.joinStr += fmt.Sprintf(" %s JOIN %s %s ON %s.%s=%s.%s", join.Key, field.Column, field.Alias, field.Alias, join.ColumnTo, join.TableFrom, join.ColumnFrom)
	} else {
		t.joinStr += fmt.Sprintf(" %s JOIN %s ON %s.%s=%s.%s", join.Key, join.TableTo, join.TableTo, join.ColumnTo, join.TableFrom, join.ColumnFrom)
	}
	return t
}

// 设置order排序
func (t *DBTable) Order(orders utils.M) *DBTable {
	var tmp []string
	for c, v := range orders {
		field := t.explainField(c)
		if field.Alias != "" {
			tmp = append(tmp, fmt.Sprintf("%s.%s %s", t.db.FormatColumn(field.Alias), t.db.FormatColumn(field.Column), v))
		} else {
			tmp = append(tmp, fmt.Sprintf("%s.%s %s", t.db.FormatColumn(t.table), t.db.FormatColumn(field.Column), v))

		}
	}
	t.orderStr = "ORDER BY " + strings.Join(tmp, ",")
	return t
}

// 设置分页
func (t *DBTable) Limit(pageSize int, page int) *DBTable {
	if page == 0 {
		t.limitStr = fmt.Sprintf("LIMIT %d", pageSize)
	} else {
		currentNum := 0
		if page > 1 {
			currentNum = (page - 1) * pageSize
		}
		t.limitStr = fmt.Sprintf("LIMIT %d,%d", currentNum, pageSize)
	}
	return t
}

// 新增
func (t *DBTable) Insert(data interface{}) (int, bool) {
	return t.db.Insert(t.table, data)
}

// 删除
func (t *DBTable) Delete() bool {
	defer t.Clear()
	_, err := t.db.Delete(t.where, t.table)
	return err == nil
}

// 更新
func (t *DBTable) Update(data utils.M) bool {
	defer t.Clear()
	err := t.db.Update(data, t.where, t.table)
	return err == nil
}

// 查询操作
func (t *DBTable) Query() *DBTable {
	var whereStr string
	if t.whereStr != "" {
		whereStr = "WHERE " + t.whereStr
	}
	t.sqlStr = fmt.Sprintf("SELECT %s FROM %s %s %s %s %s",
		t.fieldStr,
		t.db.FormatColumn(t.table),
		t.joinStr,
		whereStr,
		t.orderStr,
		t.limitStr,
	)
	return t
}

// 获取查询记录条数
func (t *DBTable) Rows() int {
	var whereStr string
	if t.whereStr != "" {
		whereStr = "WHERE " + t.whereStr
	}
	sqlStr := fmt.Sprintf("SELECT count(*) FROM %s %s %s", t.db.FormatColumn(t.table), t.joinStr, whereStr)

	row := t.db.QueryRow(sqlStr, t.values...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// 返回查询结果
func (t *DBTable) Result() ([]utils.M, error) {
	defer t.Clear()
	return t.db.Query(t.sqlStr, t.values...)
}

// 清除查询条件
func (t *DBTable) Clear() {
	t.whereStr = ""
	t.joinStr = ""
	t.groupStr = ""
	t.fieldStr = ""
	t.orderStr = ""
	t.limitStr = ""
	t.sqlStr = ""
	t.where = nil
	t.values = []interface{}{}
}
