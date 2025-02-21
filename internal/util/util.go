package util

import "unicode"

func LuhnCheck(orderNumber string) bool {
	sum := 0
	alt := false

	for i := len(orderNumber) - 1; i >= 0; i-- {
		r := rune(orderNumber[i])
		if !unicode.IsDigit(r) {
			return false
		}

		n := int(r - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}

	return sum%10 == 0
}
