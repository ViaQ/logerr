package sink

import (
	"strconv"
)

// Verbosity is a level of verbosity to log between 0 and math.MaxInt32
// However it is recommended to keep the numbers between 0 and 3
type Verbosity int

func (v Verbosity) String() string {
	return strconv.Itoa(int(v))
}

// MarshalJSON marshas JSON
func (v Verbosity) MarshalJSON() ([]byte, error) {
	return []byte(v.String()), nil
}
