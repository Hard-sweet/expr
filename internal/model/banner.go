package model

type Banner struct {
	Service string            `json:"service,omitempty"`
	Product string            `json:"product,omitempty"`
	Depth   int               `json:"depth"`
	Summary string            `json:"summary,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
	Raw     []string          `json:"raw,omitempty"`
}
