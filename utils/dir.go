package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

// 判断文件是否存在
func Exist(filename string) bool {
	// 返回文件详情
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// 判断路径是否存在
func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

// 读取目录路径
func ReadDir(path string) (files []string, dirs []string, err error) {
	path = fixDirPath(path)
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}
	for _, d := range dir {
		if d.IsDir() {
			dirs = append(dirs, path+"/"+d.Name()+"/")
		} else {
			files = append(files, path+"/"+d.Name())
		}
	}
	return
}

// 混合目录路径
func fixDirPath(path string) string {
	return strings.TrimRight(path, "/")
}
