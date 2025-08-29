package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

func convertOrdinalToString(n uint) string {
	if n == 0 {
		return ""
	}

	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return strconv.Itoa(int(n)) + "th"
	}
}

func getFileLine(skip int) string {
	if _, file, line, ok := runtime.Caller(int(skip) + 1); ok && line > 0 {
		return fmt.Sprintf("%s:%d", file, line)
	} else {
		return ""
	}
}

// AddPrefixPerLine ...
func addPrefixPerLine(text string, prefix string) string {
	sb := strings.Builder{}

	first := true
	array := strings.Split(text, "\n")
	for idx, v := range array {
		if first {
			first = false
		} else {
			sb.WriteByte('\n')
		}

		if v != "" || idx == 0 || idx != len(array)-1 {
			sb.WriteString(prefix)
			sb.WriteString(v)
		}
	}

	return sb.String()
}

func isNil(val any) (ret bool) {
	defer func() {
		if e := recover(); e != nil {
			ret = false
		}
	}()

	if val == nil {
		return true
	}

	return reflect.ValueOf(val).IsNil()
}
