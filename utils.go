package jira

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2018 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// paramsToQuery convert params to query string
func paramsToQuery(params interface{}) string {
	var result string

	t := reflect.TypeOf(params)
	v := reflect.ValueOf(params)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		switch value.Type().String() {
		case "string":
			if value.String() != "" {
				result += field.Tag.Get("query") + "=" + esc(value.String()) + "&"
			}

		case "int":
			if value.Int() != 0 {
				result += field.Tag.Get("query") + "=" + fmt.Sprintf("%d", value.Int()) + "&"
			}

		case "bool":
			if value.Bool() {
				result += field.Tag.Get("query") + "=1&"
			}

		case "time.Time":
			d := value.Interface().(time.Time)
			if !d.IsZero() {
				result += field.Tag.Get("query") + "=" + fmt.Sprintf("%d-%02d-%02d", d.Year(), d.Month(), d.Day()) + "&"
			}

		case "[]string":
			if value.Len() > 0 {
				result += formatSlice(field.Tag.Get("query"), value) + "&"
			}
		}
	}

	if result == "" {
		return ""
	}

	return result[:len(result)-1]
}

// formatSlice format slice
func formatSlice(tag string, s reflect.Value) string {
	var result string

	name, unwrap := parseSliceTag(tag)

	if !unwrap {
		result += name + "="
	}

	for i := 0; i < s.Len(); i++ {
		v := s.Index(i)

		if unwrap {
			result += name + "=" + esc(v.String()) + "&"
		} else {
			result += esc(v.String()) + ","
		}
	}

	return result[:len(result)-1]
}

// parseSliceTag parse slice tag and return tag name and unwrap flag
func parseSliceTag(tag string) (string, bool) {
	if !strings.Contains(tag, ",unwrap") {
		return tag, false
	}

	return tag[:strings.Index(tag, ",")], true
}

// esc escapes the string so it can be safely placed inside a URL query
func esc(s string) string {
	return url.QueryEscape(s)
}

// readField read field from given data
func readField(data string, index int, separator rune) string {
	if data == "" || index < 0 {
		return ""
	}

	curIndex, startPointer := -1, -1

	for i, r := range data {
		if r == separator {
			if curIndex == index {
				return data[startPointer:i]
			}

			startPointer = i + 1
			curIndex++
			continue
		}

		if startPointer == -1 {
			startPointer = i
			curIndex++
		}
	}

	if index > curIndex {
		return ""
	}

	return data[startPointer:]
}