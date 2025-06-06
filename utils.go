package jira

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
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

// Supported options
const (
	_OPTION_UNWRAP  = "unwrap"
	_OPTION_RESPECT = "respect"
	_OPTION_REVERSE = "reverse"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// paramsToQuery convert params to query string
func paramsToQuery(params any) string {
	var result string

	t := reflect.TypeOf(params)
	v := reflect.ValueOf(params)

	for i := range t.NumField() {
		field := t.Field(i)
		value := v.Field(i)
		tag := field.Tag.Get("query")

		switch value.Type().String() {
		case "string":
			if value.String() != "" {
				result += tag + "=" + esc(value.String()) + "&"
			} else if hasTagOption(tag, _OPTION_RESPECT) {
				result += getTagName(tag) + "=&"
			}

		case "int":
			if value.Int() != 0 {
				result += tag + "=" + fmt.Sprintf("%d", value.Int()) + "&"
			} else if hasTagOption(tag, _OPTION_RESPECT) {
				result += getTagName(tag) + "=0&"
			}

		case "bool":
			b := value.Bool()
			if hasTagOption(tag, _OPTION_REVERSE) && b {
				result += getTagName(tag) + "=false&"
			} else {
				if b {
					result += getTagName(tag) + "=true&"
				} else if hasTagOption(tag, _OPTION_RESPECT) {
					result += getTagName(tag) + "=false&"
				}
			}

		case "time.Time":
			d := value.Interface().(time.Time)
			if !d.IsZero() {
				result += tag + "=" + fmt.Sprintf("%d-%02d-%02d", d.Year(), d.Month(), d.Day()) + "&"
			}

		case "[]string":
			if value.Len() > 0 {
				result += formatSlice(tag, value) + "&"
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

	name := getTagName(tag)
	unwrap := hasTagOption(tag, _OPTION_UNWRAP)

	if !unwrap {
		result += name + "="
	}

	for i := range s.Len() {
		v := s.Index(i)

		if unwrap {
			result += name + "=" + esc(v.String()) + "&"
		} else {
			result += esc(v.String()) + ","
		}
	}

	return result[:len(result)-1]
}

// getTagOption extract option from tag
func hasTagOption(tag, option string) bool {
	if !strings.Contains(tag, ",") {
		return false
	}

	return tag[strings.Index(tag, ",")+1:] == option
}

// getTagName return tag name
func getTagName(tag string) string {
	if !strings.Contains(tag, ",") {
		return tag
	}

	return tag[:strings.Index(tag, ",")]
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
