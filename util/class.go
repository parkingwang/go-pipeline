package util

import (
	"reflect"
	"strings"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func SimpleClassName(obj interface{}) string {
	return SimpleTypeName(reflect.TypeOf(obj))
}

func SimpleTypeName(typed reflect.Type) string {
	name := typed.String()
	if dotIdx := strings.LastIndex(name, "."); dotIdx >= 0 {
		name = name[dotIdx+1:]
	}
	return name
}
