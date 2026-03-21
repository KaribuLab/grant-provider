package grantprovider

// ListMap devuelve un slice nuevo aplicando fn a cada elemento de slice.
func ListMap[T any, R any](slice []T, fn func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}
