package bencodeparser

func indexOf(haystack []byte, needle byte) int {
	for i, v := range haystack {
		if v == needle {
			return i
		}
	}
	return -1
}

func isDigit(maybeDigit byte) bool {
	return maybeDigit >= '0' && maybeDigit <= '9'
}
