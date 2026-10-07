package main

func rotateRunes(s string, shift int) string {
	runes := []rune(s)
	lengthOfS := len(runes)
	if lengthOfS == 0 {
		return s
	}

	clearShift := shift % lengthOfS
	var answer string

	if (clearShift >= 0) {
		answer = string(runes[clearShift:]) + string(runes[:clearShift])
	} else if (clearShift < 0) {
		rightShift := lengthOfS + clearShift
		answer = string(runes[rightShift:]) + string(runes[:rightShift])
	}
	return answer
}
