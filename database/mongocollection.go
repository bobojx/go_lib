package database

import (
	"go_lib/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

// 查询结果
type QueryResult struct {
	List  interface{} `json:"list"`
	Count int         `json:"count"`
}

type Collection struct {
	db      *MongoDB
	tabName string
}

// 实例化集合
func NewCollection(mdb *MongoDB, tabName string) *Collection {
	return &Collection{db: mdb, tabName: tabName}
}

// 新增
func (c *Collection) Insert(rows ...interface{}) error {
	t := c.db.Table(c.tabName)
	err := t.Insert(rows...)
	if err != nil {
		return err
	}
	return nil
}

// 删除
func (c *Collection) Delete(where bson.M) error {
	t := c.db.Table(c.tabName)
	_, err := t.RemoveAll(where)
	if err != nil {
		return err
	}
	return nil
}

// 更新
func (c *Collection) Update(where bson.M, update bson.M) error {
	t := c.db.Table(c.tabName)
	err := t.Update(where, update)
	if err != nil {
		return err
	}
	return nil
}

// 更新所有条件的数据
func (c *Collection) UpdateAll(where bson.M, update bson.M) (*mgo.ChangeInfo, error) {
	t := c.db.Table(c.tabName)
	return t.UpdateAll(where, update)
}

// 更新查询条件的数据，如果未找到则新增要更新的数据
func (c *Collection) Upset(where bson.M, update bson.M) error {
	t := c.db.Table(c.tabName)
	_, err := t.Upsert(where, update)
	if err != nil {
		return err
	}
	return nil
}

// 查找一条数据
func (c *Collection) Find(where bson.M, row interface{}) error {
	t := c.db.Table(c.tabName)
	err := t.Find(where).One(&row)
	if err != nil {
		return err
	}
	return nil
}

// 查询条数
func (c *Collection) Count(where bson.M) int {
	t := c.db.Table(c.tabName)
	count, err := t.Find(where).Count()
	if err != nil {
		return 0
	}
	return count
}

// 分页查询
// where 查询条件
// page 当前页数
// pageNum 每页显示条数
// sortList 排序
// structType 返回数据格式类型
// format 格式化回调方法
func (c *Collection) Query(where bson.M, page int, pageNum int, sortList []string, structType interface{}, format func(interface{})) (*QueryResult, error) {
	t := c.db.Table(c.tabName)
	var list []interface{}
	var err error
	var count int
	var resList *mgo.Iter

	if sortList == nil {
		sortList = []string{"-_id"}
	}
	if pageNum != 0 {
		resList = t.Find(where).Sort(sortList...).Skip((page - 1) * pageNum).Limit(pageNum).Iter()
	} else {
		resList = t.Find(where).Sort(sortList...).Iter()
	}
	count, _ = t.Find(where).Count()

	result := c.getStructType(structType)
	for resList.Next(result) {
		if format != nil {
			format(result)
		}
		list = append(list, result)
		// 重置数据
		result = c.getStructType(structType)
	}

	res := &QueryResult{
		List:  list,
		Count: count,
	}
	return res, err
}

// 聚合操作
func (c *Collection) Pipe(pipeLine ...bson.M) (bson.M, error) {
	t := c.db.Table(c.tabName)
	p := t.Pipe(pipeLine)
	result := bson.M{}
	err := p.One(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 获取结构体的类型
func (c *Collection) getStructType(i interface{}) interface{} {
	if i == nil {
		return utils.M{}
	}

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	}
	return reflect.New(t).Interface()
}

// 删除当前集合
func (c *Collection) Drop() error {
	return c.db.Table(c.tabName).DropCollection()
}

// 关闭连接
func (c *Collection) Close() {
	c.db.Close()
}
