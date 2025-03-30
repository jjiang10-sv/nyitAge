package parts1


type Signal int


func runeToSignal(key rune) Signal {
	if key >= 'A' && key <= 'Z' {
		key -= 'A'
		return Signal(key) + 1
	}
	return -1
}

func signalToRune(key Signal) rune {
	if key >=1 && key <= 26 {
		return rune(key-1)+'A'
	}
	return 0
}

func fixAlpha(key Signal) Signal {
	if key < 1 {
		key += 26
	}
	if key >26 {
		key -= 26
	}
	return key
}