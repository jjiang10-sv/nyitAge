package enigma1

import (
	"fmt"

	"errors"

	"github.com/ibraimgm/enigma/machine/parts1"
)

type Enigma interface {
	Reflector() string
	Fast() string
	Middle() string
	Slow() string
	SetSetting(settings string, settingType SettingType) error
	Window() string
	Ring() string
	Configure(windowSetting, ringSetting string) error
	Encode(input rune) (rune, bool)
	EncodeMessage(input string, blockSize uint) string
}

type SettingType int

const (
	Window SettingType = iota
	Ring
)

type MyEnigma struct {
	// Add fields as necessary for your implementation
	keyboard   parts1.Keyboard
	plugboard  parts1.Plugboard
	fast       parts1.Rotor
	middle     parts1.Rotor
	slow       parts1.Rotor
	reflector  parts1.Reflector
	lightboard parts1.Lightboard
}

func WithDefault() Enigma {
	fast, _ := parts1.GetRotor("I")
	middle, _ := parts1.GetRotor("II")
	slow, _ := parts1.GetRotor("III")
	return AssembleEnigma(
		parts1.DefaultKeyboard,
		parts1.DefaultPlugBoard,
		fast,
		middle,
		slow,
		parts1.Reflectors["B"],
		parts1.DefaultLightboard,
	)
}

// func WithRotors(fast, middle,slow string, )

func WithConfig(windowSetting, ringSetting string) Enigma {
	myEnigma := WithDefault()
	myEnigma.Configure(windowSetting, ringSetting)
	return myEnigma
}

func WithRotors(slow, middle, fast, reflector string) (Enigma, error) {
	var r1, r2, r3 parts1.Rotor
	var err error

	if r1, err = parts1.GetRotor(slow); err != nil {
		return nil, err
	}
	if r2, err = parts1.GetRotor(middle); err != nil {
		return nil, err
	}
	if r3, err = parts1.GetRotor(fast); err != nil {
		return nil, err
	}

	ref, ok := parts1.Reflectors[reflector]
	if !ok {
		return nil, errors.New("unknown reflector: '" + reflector + "'")
	}

	return AssembleEnigma(
		parts1.DefaultKeyboard,
		parts1.DefaultPlugBoard,
		r3,
		r2,
		r1,
		ref,
		parts1.DefaultLightboard,
	), nil
}

func AssembleEnigma(keyboard parts1.Keyboard, plugboard parts1.Plugboard, fast, middle, slow parts1.Rotor, reflector parts1.Reflector, lightboard parts1.Lightboard) Enigma {
	enigmaImp := &MyEnigma{keyboard, plugboard, fast, middle, slow, reflector, lightboard}
	enigmaImp.SetSetting("AAA", Window)
	enigmaImp.SetSetting("AAA", Ring)
	return enigmaImp
}

// Implementing the Enigma interface methods
func (e *MyEnigma) Fast() string {
	// Implementation here
	return e.fast.ID()
}

func (e *MyEnigma) Middle() string {
	// Implementation here
	return e.middle.ID()
}

func (e *MyEnigma) Slow() string {
	// Implementation here
	return e.slow.ID()
}

func (e *MyEnigma) SetSetting(settings string, settingType SettingType) error {
	// Implementation here
	var runes [3]rune
	if settings == "" {
		runes = [3]rune{'A', 'A', 'A'}
	} else {
		if len(settings) != 3 {
			return fmt.Errorf(" settings %s %s need be length of 3", settings, settingType)
		}
		for i, r := range settings {
			if r < 'A' || r > 'Z' {
				return fmt.Errorf("rune %r is outside the range on %s %s ", r, settings, settingType)
			}
			runes[i] = r
		}
	}
	if settingType == Window {
		e.slow.SetWindow(runes[0])
		e.middle.SetWindow(runes[1])
		e.fast.SetWindow(runes[2])
	} else if settingType == Ring {
		e.slow.SetRing(runes[0])
		e.middle.SetRing(runes[1])
		e.fast.SetRing(runes[2])
	}
	return nil
}

func (e *MyEnigma) Reflector() string {
	return e.reflector.ID()
}

func (e *MyEnigma) Window() string {
	// Implementation here
	return string([]rune{e.slow.Window(), e.middle.Window(), e.fast.Window()})
}

func (e *MyEnigma) Ring() string {
	// Implementation here
	return string([]rune{e.slow.Ring(), e.middle.Ring(), e.fast.Ring()})
}

func (e *MyEnigma) Configure(windowSetting, ringSetting string) error {
	// Implementation here
	if err := e.SetSetting(windowSetting, Window); err != nil {
		return err
	}
	return e.SetSetting(ringSetting, Ring)
}

func (e *MyEnigma) Encode(input rune) (rune, bool) {
	// Implementation here
	signal, ok := e.keyboard.InputKey(input)
	if !ok {
		return input, false
	}
	// rotate
	if e.fast.IsNotched() {
		if e.middle.IsNotched() {
			e.slow.Move(1)
		}
		e.middle.Move(1)
	} else {
		if e.middle.IsNotched() {
			e.slow.Move(1)
			e.middle.Move(1)
		}
	}
	e.fast.Move(1)
	signal = e.plugboard.Translate(signal)
	enigmaRotors := [3]parts1.Rotor{e.fast, e.middle, e.slow}
	for _, rotor := range enigmaRotors {
		signal = rotor.Scramble(signal)
	}
	signal = e.reflector.Reflect(signal)
	for _, rotor := range enigmaRotors {
		signal = rotor.Reverse(signal)
	}
	signal = e.plugboard.Translate(signal)
	return e.lightboard.Light(signal), true
}

func (e *MyEnigma) EncodeMessage(input string, blockSize uint) string {
	// Implementation here
	if blockSize == 0 {
		return ""
	}
	// if blockSize < 0 then It never record the blocksize
	var blockTracker uint = 0
	encodedMsg := ""
	for _, s := range input {
		r, ok := e.Encode(rune(s))
		if ok {
			if blockTracker == blockSize {
				encodedMsg += " "
				blockTracker = 0
			}
			encodedMsg += string(r)
			blockTracker++
		}
	}
	return encodedMsg
}
