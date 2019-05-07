package tau

// OptionalIDFields -
type OptionalIDFields struct {
	Email string `json:"email,omitempty"`
}

// TangleID is attached to the tangle and used for identification
type TangleID struct {
	OptionalIDFields
	PublicKey string `json:"pbk"`
	Signature string `json:"sgn"`
}
