package gox

import (
	"strings"
)

// AddSlashes 使用反斜线引用字符串
func AddSlashes(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
