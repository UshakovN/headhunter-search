package utils

func ForEach[T any](f func(T), s ...T) {
	for _, s := range s {
		f(s)
	}
}
