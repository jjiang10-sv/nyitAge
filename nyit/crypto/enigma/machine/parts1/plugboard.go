package parts1

type Plugboard interface {
	Translate(input Signal) Signal
}

type plugboardImpl struct {
	id    string
	plugs map[Signal]Signal
}

func createPlugboardImpl(id string, plugs string) *plugboardImpl {
	runes := []rune(plugs)
	plugsMap := map[Signal]Signal{}
	for i := 0; i+1 < len(runes); i += 2 {
		a, b := runeToSignal(runes[i]), runeToSignal(runes[i+1])
		if a == -1 || b == -1 {
			continue
		}
		plugsMap[a], plugsMap[b] = b, a
	}
	return &plugboardImpl{id: id, plugs: plugsMap}
}
func (board *plugboardImpl) Translate(input Signal) Signal {
	if b, ok := board.plugs[input]; ok {
		return b
	}
	return input
}

var DefaultPlugBoard Plugboard = &plugboardImpl{}
