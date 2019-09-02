package utils

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

type M map[string]interface{}

// 转成json格式
func (m *M) ToJson() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return data
}

// 转换成字符串
func (m *M) ToJsonString() string {
	data := m.ToJson()
	if data == nil {
		return ""
	}
	return string(data)
}

// 解析成json格式
func (m *M) ParseJson(data []byte) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *M) ParseJsonString(data string) error {
	return m.ParseJson([]byte(data))
}

// 下划线转驼峰
func Under2Hump(str string) string {
	strMap := strings.Split(str, "_")
	for i, c := range strMap {
		strMap[i] = UcFirst(c)
	}
	return strings.Join(strMap, "")
}

// 首字母大写
func UcFirst(str string) string {
	first := str[0:1]
	long := str[1:]
	return strings.ToUpper(first) + long
}

const baseString = "ABCDEFGHIGKLMNOPQRSTUVWXYZabcdefghijklmnopqrskuvwxyz1234567890"

// 生成随机字符串
func RandString(num int) string {
	time.Sleep(1)
	rand.Seed(time.Now().UnixNano())
	str := make([]string, 0)
	randLen := len(baseString)
	for i := 0; i < num; i++ {
		str = append(str, string(baseString[rand.Intn(randLen)]))
	}
	return strings.Join(str, "")
}

// 机构转map
// fields指定转换的字段
func Struct2Map(obj interface{}, fields []string) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	// 如果类型为指针，则返回一个类型的元素类型。
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		t = reflect.TypeOf(obj).Elem()
		v = reflect.ValueOf(obj).Elem()
	}

	data := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if fields != nil {
			if flag := StringIndexOf(fields, tag); flag != -1 {
				data[tag] = v.Field(i).Interface()
			}
		} else {
			data[tag] = v.Field(i).Interface()
		}
	}
	return data
}

// map转结构体
// str必须是指针类型
func Map2Struct(obj interface{}, str interface{}) error {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, str)
	return err
}

// 获取下标
func StringIndexOf(arr []string, search string) int {
	for i, v := range arr {
		if search == v {
			return i
		}
	}
	return -1
}

// api接口输出
func ApiOutput(status bool, msg string, data interface{}) *map[string]interface{} {
	return &map[string]interface{}{"status": status, "msg": msg, "data": data}
}
