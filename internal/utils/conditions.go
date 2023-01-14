package utils

func IfThenElse[T any](condition bool, a interface{}, b interface{}) T {
	if condition {
		return a.(T)
	}
	return b.(T)
}
