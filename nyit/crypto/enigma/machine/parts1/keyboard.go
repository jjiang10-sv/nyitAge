package parts1

type Keyboard interface {
	InputKey(key rune) (Signal, bool)
}

type keyboardImpl struct{}

var DefaultKeyboard Keyboard = &keyboardImpl{}

func (*keyboardImpl) InputKey(key rune) (Signal, bool) {
	// transfer the keyboard key to Signal (1-26)
	// convert to uppercase letter
	if key >= 'a' && key <= 'z' {
		key -= 32
	}

	s := runeToSignal(key)
	return s, s != -1

}
