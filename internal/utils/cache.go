package utils

import "strings"

func GetCacheKey(params ...string) string {
	return strings.Join(params, "***")
}
