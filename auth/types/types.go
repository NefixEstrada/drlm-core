package types

// Type is an authentication type
type Type int

const (
	// Local is the local authentication type. It authenticates agains the local DB
	Local Type = iota
)

// String returns the authentication type in a string
func (t Type) String() string {
	switch t {
	case Local:
		return "local"

	default:
		return "unknown"
	}
}
