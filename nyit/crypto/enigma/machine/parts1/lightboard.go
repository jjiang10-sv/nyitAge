package parts1

type Lightboard interface{
	Light(key Signal) rune
}

type lightboardImpl struct{}

var DefaultLightboard Lightboard = &lightboardImpl{}

func (*lightboardImpl)Light(key Signal) rune {
	return signalToRune(key)
}