package parts1

type Reflector interface{
	ID() string
	Reflect(input Signal) Signal
}

func (board *plugboardImpl) ID() string{
	return board.id
}

func (board *plugboardImpl) Reflect(input Signal) Signal{
	return board.Translate(input)
}

var Reflectors = map[string]Reflector{
	"B":      Reflector(createPlugboardImpl("B", "AYBRCUDHEQFSGLIPJXKNMOTZVW")),
	"C":      Reflector(createPlugboardImpl("C", "AFBVCPDJEIGOHYKRLZMXNWTQSU")),
	"B Dünn": Reflector(createPlugboardImpl("B Dünn", "AEBNCKDQFUGYHWIJLOMPRXSZTV")),
	"C Dünn": Reflector(createPlugboardImpl("C Dünn", "ARBDCOEJFNGTHKIVLMPWQZSXUY")),
}