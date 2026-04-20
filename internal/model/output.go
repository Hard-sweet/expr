package model

type ScanResult struct {
	Scope        Scope           `json:"scope"`
	Services     []ServiceRecord `json:"services"`
	Assets       []Asset         `json:"assets"`
	Answers      DNSAnswers      `json:"answers"`
	ServiceTypes []string        `json:"service_types"`
}

type Scope struct {
	CIDR      string `json:"cidr,omitempty"`
	Ports     string `json:"ports"`
	Interface string `json:"interface,omitempty"`
	Timeout   string `json:"timeout"`
}
