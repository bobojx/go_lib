package database

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
