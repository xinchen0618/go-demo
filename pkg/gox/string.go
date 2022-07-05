package gox

import (
	"strings"
	"unicode/utf8"
)

// AddSlashes 使用反斜线引用字符串
//  @param s string
//  @return string
func AddSlashes(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`)
	return r.Replace(s)
}

// Substr 返回字符串的子串
//  中英文皆适用
//  @param s string
//  @param offset int
// 	  若offset为非负数, 返回的子串将从字符串的offset位置开始, 从0开始计算.
//	  若offset为负数，返回的子串将从字符串结尾处向前数第offset个字符开始.
//	  若offset大于字符串长度，返回空字符串.
//  @param length int 长度
//    若length大于字符串长度, 返回的子串将从offset位置开始直到字符串结尾.
//  @return string
func Substr(s string, offset, length int) string {
	// 字符串长度
	runeCount := utf8.RuneCountInString(s)
	if 0 == runeCount {
		return ""
	}

	// 开始位置
	if offset >= runeCount {
		return ""
	}
	if offset < 0 { // 从尾部向前数
		offset = runeCount + offset
		if offset < 0 {
			offset = 0
		}
	}

	// 结束位置
	end := offset + length
	if end > runeCount || end < 0 {
		end = runeCount
	}

	return string([]rune(s)[offset:end])
}
