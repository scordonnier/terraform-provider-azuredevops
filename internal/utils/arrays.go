package utils

import "golang.org/x/exp/slices"

func Difference(a, b *[]string) *[]string {
	var c []string
	for _, s := range *a {
		if slices.Contains(*b, s) {
			continue
		}
		c = append(c, s)
	}
	return &c
}
