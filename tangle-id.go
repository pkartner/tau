package tau

// OptionalIDFields -
type OptionalIDFields struct {
	Name string `json:"name,omitempty"`
}

// TangleID is attached to the tangle and used for identification
type TangleID struct {
	OptionalIDFields
	PublicKey string `json:"pbk"`
	Signature string `json:"sgn"`
}
