package helpers

import (
	"runtime"
	"strings"
)

func GetCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	strs := strings.Split((runtime.FuncForPC(pc).Name()), "/")
	return strs[len(strs)-1]
}
