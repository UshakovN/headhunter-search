package utils

import (
	"fmt"
	"strconv"
)

func OneOf(f func(s string) bool, s ...string) bool {
	for _, s := range s {
		if f(s) {
			return true
		}
	}
	return false
}

func MustCast[T any](s string) T {
	var (
		iface any
		err   error
	)
	switch typ := any(*new(T)).(type) {
	case string:
		iface = any(s).(T)
	case int:
		iface, err = strconv.Atoi(s)
	case int64:
		iface, err = strconv.ParseInt(s, 10, 64)
	default:
		err = fmt.Errorf("unsupported type %v", typ)
	}
	if err != nil {
		panic(err)
	}
	return iface.(T)
}
