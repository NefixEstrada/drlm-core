// SPDX-License-Identifier: AGPL-3.0-only

package types

// Type is an authentication type
type Type int

const (
	// Unknown is an unknown authentication type
	Unknown Type = iota
	// Local is the local authentication type. It authenticates agains the local DB
	Local
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
