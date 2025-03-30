package parts1

import (
	"fmt"
	"math"
)

type Rotor interface {
	ID() string
	Move(step Signal)
	Window() rune
	SetWindow(window rune)
	Ring() rune
	SetRing(ring rune)
	IsNotched() bool
	Scramble(input Signal) Signal
	Reverse(input Signal) Signal
}

type rotorImpl struct {
	id       string
	position Signal
	ring     Signal
	sequence []Signal
	notches  []Signal
}

func GetRotor(id string) (Rotor, error) {
	switch id {
	case "I":
		return CreateRotor("I", "EKMFLGDQVZNTOWYHXUSPAIBRCJ", "Q"), nil
	case "II":
		return CreateRotor("II", "AJDKSIRUXBLHWTMCQGZNPYFVOE", "E"), nil
	case "III":
		return CreateRotor("III", "BDFHJLCPRTXVZNYEIWGAKMUSQO", "V"), nil
	case "IV":
		return CreateRotor("IV", "ESOVPZJAYQUIRHXLNFTGKDCMWB", "J"), nil
	case "V":
		return CreateRotor("V", "VZBRGITYUPSDNHLXAWMJQOFECK", "Z"), nil
	case "VI":
		return CreateRotor("VI", "JPGVOUMFYQBENHZRDKASXLICTW", "ZM"), nil
	case "VII":
		return CreateRotor("VII", "NZJHGRCXMYSWBOUFAIVLPEKQDT", "ZM"), nil
	case "VIII":
		return CreateRotor("VIII", "FKQHTLXOCBJSPDZRAMEWNIUYGV", "ZM"), nil
	default:
		return nil, fmt.Errorf("not able to get the rotor with ID %s ", id)
	}
}

func CreateRotor(rotorID, sequence, notches string) Rotor {
	sequenceRunes, notchesRunes := []rune(sequence), []rune(notches)
	sequenceLen, notchesLen := len(sequenceRunes), len(notchesRunes)
	sequenceInt, notchesInt := make([]Signal, sequenceLen), make([]Signal, notchesLen)
	for i := 0; i < sequenceLen; i++ {
		sequenceInt[i] = runeToSignal(sequenceRunes[i])
	}
	for i := 0; i < notchesLen; i++ {
		notchesInt[i] = runeToSignal(notchesRunes[i])
	}
	return &rotorImpl{id: rotorID, position: 1, ring: 1, sequence: sequenceInt, notches: notchesInt}
}

func (r *rotorImpl) ID() string {
	return r.id
}

func (r *rotorImpl) Move(step Signal) {
	newPos := r.position - 1 + step
	r.position = Signal(math.Mod(float64(newPos+26), 26)) + 1
}

func (r *rotorImpl) Window() rune {
	return signalToRune(r.position)
}

func (r *rotorImpl) SetWindow(window rune) {
	r.position = fixAlpha(runeToSignal(window))
}

func (r *rotorImpl) Ring() rune {
	return signalToRune(r.ring)
}

func (r *rotorImpl) SetRing(ring rune) {
	r.ring = fixAlpha(runeToSignal(ring))
}

func (r *rotorImpl) IsNotched() bool {
	for _, notch := range r.notches {
		if notch == r.position {
			return true
		}
	}
	return false
}

func (r *rotorImpl) Scramble(input Signal) Signal {
	// Implement the Scramble logic here
	from := fixAlpha(input - r.ring + r.position)
	from = r.sequence[from-1]
	from = fixAlpha(from + r.ring - r.position)
	return from // Placeholder
}

func (r *rotorImpl) Reverse(input Signal) Signal {
	// Implement the Reverse logic here
	from := fixAlpha(input - r.ring + r.position)
	for _, v := range r.sequence {
		if v == from {
			from += 1
			break
		}
	}
	from = fixAlpha(from + r.ring - r.position)
	return from // Placeholder
}
