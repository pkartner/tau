package tau

import (
	"fmt"
)

// ErrEmptySeed is returned when a seed is empty
var ErrEmptySeed = fmt.Errorf("seed cannot be empty")

// ErrEmptyAuthorization -
var ErrEmptyAuthorization = fmt.Errorf("the authorization header cannot be empty")

// ErrInvalidIOTAAPI -
var ErrInvalidIOTAAPI = fmt.Errorf("invalid iota api")

// ErrCallingTangle -
var ErrCallingTangle = fmt.Errorf("error calling tangle")

// ErrNilRequest -
var ErrNilRequest = fmt.Errorf("nil request")
