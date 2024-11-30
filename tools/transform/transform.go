package transform

import "strings"

// TODO(patrik): Add option to escape html?
// TODO(patrik): Add tests
func String(s string) string {
	return strings.TrimSpace(s)
}

func StringPtr(s *string) *string {
	if s == nil {
		return nil
	}

	*s = String(*s)
	return s
}

func StringArrayPtr(arr *[]string) *[]string {
	if arr == nil {
		return nil
	}

	v := *arr
	for i, s := range v {
		v[i] = String(s)
	}

	return arr
}

func DiscardEmptyStringEntries(arr *[]string) *[]string {
	if arr == nil {
		return nil
	}

	var res []string

	v := *arr
	for _, s := range v {
		if s != "" {
			res = append(res, s)
		}
	}

	if len(res) == 0 {
		return nil
	}

	return &res
}
