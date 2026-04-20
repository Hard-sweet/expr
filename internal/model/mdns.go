package model

import "time"

type DiscoveryResult struct {
	ServiceTypes []string     `json:"service_types"`
	Entries      []MDNSEntry  `json:"entries"`
	Answers      DNSAnswers   `json:"answers"`
	ObservedAt   time.Time    `json:"observed_at"`
	Interface    string       `json:"interface,omitempty"`
}

type DNSAnswers struct {
	PTR []string `json:"ptr"`
}

type MDNSEntry struct {
	Instance  string            `json:"instance"`
	Service   string            `json:"service"`
	Domain    string            `json:"domain"`
	HostName  string            `json:"hostname,omitempty"`
	Port      int               `json:"port,omitempty"`
	IPv4      []string          `json:"ipv4,omitempty"`
	IPv6      []string          `json:"ipv6,omitempty"`
	Text      []string          `json:"text,omitempty"`
	TextKV    map[string]string `json:"text_kv,omitempty"`
	TTL       uint32            `json:"ttl,omitempty"`
	Interface string            `json:"interface,omitempty"`
}
