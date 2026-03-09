package validation

func IsValidLimit(limit int) bool {
	if limit < 1 {
		return false
	}
	return true
}

func IsValidPrice(price float64) bool {
	if price < 1 {
		return false
	}
	return true
}