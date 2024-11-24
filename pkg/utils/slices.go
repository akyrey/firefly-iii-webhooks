package utils

// Filter returns a new slice containing only the elements of the original slice for which the test function returns true.
func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
