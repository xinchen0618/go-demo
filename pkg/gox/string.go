package gox

import "unicode/utf8"

// AddSlashes 使用反斜线引用字符串
//  @param str string
//  @return string
func AddSlashes(str string) string {
	var tmpRune []rune
	strRune := []rune(str)
	for _, ch := range strRune {
		switch ch {
		case []rune{'\\'}[0], []rune{'"'}[0], []rune{'\''}[0]:
			tmpRune = append(tmpRune, []rune{'\\'}[0])
			tmpRune = append(tmpRune, ch)
		default:
			tmpRune = append(tmpRune, ch)
		}
	}
	return string(tmpRune)
}

// Substr 返回字符串的子串
//  中英文皆适用
//  @param s string
//  @param offset int
// 	  若offset为非负数, 返回的子串将从字符串的offset位置开始, 从0开始计算.
//	  若offset为负数，返回的子串将从字符串结尾处向前数第offset个字符开始.
//	  若offset大于字符串长度，返回空字符串.
//  @param length int 若length大于字符串长度, 取最大可取长度
//  @return string
func Substr(s string, offset, length int) string {
	counts := utf8.RuneCountInString(s)
	if 0 == counts {
		return ""
	}
	if offset >= length {
		return ""
	}

	if offset < 0 {
		offset = counts + offset
		if offset < 0 {
			offset = 0
		}
	}

	end := offset + length
	if end > counts {
		end = counts
	}

	return string([]rune(s)[offset:end])
}
