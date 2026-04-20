package model

type ServiceRecord struct {
	Name       string   `json:"name,omitempty"`
	IP         string   `json:"ip,omitempty"`
	IPv6       []string `json:"ipv6,omitempty"`
	Port       int      `json:"port,omitempty"`
	Proto      string   `json:"proto,omitempty"`
	Service    string   `json:"service,omitempty"`
	ServiceFQDN string  `json:"service_fqdn,omitempty"`
	Hostname   string   `json:"hostname,omitempty"`
	TTL        uint32   `json:"ttl,omitempty"`
	Path       string   `json:"path,omitempty"`
	Banner     Banner   `json:"banner"`
	SourceText []string `json:"source_text,omitempty"`
}

type Asset struct {
	PrimaryIP string          `json:"primary_ip,omitempty"`
	Hostname  string          `json:"hostname,omitempty"`
	Services  []ServiceRecord `json:"services,omitempty"`
}
