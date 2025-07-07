package services

func IsValidLuhn(number string) bool {
	sum := 0
	alt := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alt = !alt
	}
	return sum%10 == 0
}
