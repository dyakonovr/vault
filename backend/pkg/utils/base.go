package utils

func Contains[T comparable](slice []T, el T) bool {
	for _, v := range slice {
		if v == el {
			return true
		}
	}

	return false
}