package tau

import (
	"fmt"
)

// ErrEmptySeed is returned when a seed is empty
var ErrEmptySeed = fmt.Errorf("seed cannot be empty")

// ErrEmptyAuthorization -
var ErrEmptyAuthorization = fmt.Errorf("the authorization header cannot be empty")

// ErrNotVerified -
var ErrNotVerified = fmt.Errorf("the request cannot be verified")

