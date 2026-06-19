package validation

func IsValidLimit(limit int) bool {
	return limit >= 1 
}

func IsValidPrice(price float64) bool {
	return price >= 1  
}
