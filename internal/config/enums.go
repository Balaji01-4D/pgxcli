package config

type OnErrorAction string

const (
	OnErrorResume OnErrorAction = "RESUME"
	OnErrorStop   OnErrorAction = "STOP"
)

func (a OnErrorAction) IsValid() bool {
	switch a {
	case OnErrorResume, OnErrorStop:
		return true
	default:
		return false
	}
}
