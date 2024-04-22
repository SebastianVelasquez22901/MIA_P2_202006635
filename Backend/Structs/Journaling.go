package Structs

type Journaling struct {
	Journaling_start int64
	Operacion        [16]byte
	Path             [64]byte
	Contenido        [64]byte
	Fecha            [16]byte
}

func NewJournal() Journaling {
	var jr Journaling
	return jr
}
