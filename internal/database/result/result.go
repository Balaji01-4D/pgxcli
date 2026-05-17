package result

type Type int

const (
	ResultTypeQuery Type = iota
	ResultTypeSpecial
	ResultTypeMultiQuery
	ResultTypeError
)

// Result marks values returned by SQL execution paths.
type Result interface {
	Type() Type
}
