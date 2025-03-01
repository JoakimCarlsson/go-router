package metadata

import "strconv"

// StatusCodeToString converts an HTTP status code to a string representation
func StatusCodeToString(code int) string {
	return strconv.Itoa(code)
}

// StatusCodeFromString converts a string representation of an HTTP status code to an integer
func StatusCodeFromString(code string) (int, error) {
	return strconv.Atoi(code)
}
