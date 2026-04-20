package util

import (
	"fmt"
	"net"
	"strings"
)

type IPNetAlias struct {
	net.IPNet
}

func ParseCIDROptional(value string) (*IPNetAlias, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	_, ipNet, err := net.ParseCIDR(value)
	if err != nil {
		return nil, fmt.Errorf("parse cidr %q: %w", value, err)
	}
	return &IPNetAlias{IPNet: *ipNet}, nil
}

func (n *IPNetAlias) ContainsString(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return false
	}
	return n.Contains(ip)
}
