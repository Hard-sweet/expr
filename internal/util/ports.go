package util

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type PortRange struct {
	allowed map[int]struct{}
	min     int
	max     int
	label   string
	full    bool
}

func ParsePortRange(value string) (PortRange, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "1-65535"
	}
	allowed := make(map[int]struct{})
	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			startText, endText, _ := strings.Cut(part, "-")
			start, err := strconv.Atoi(strings.TrimSpace(startText))
			if err != nil {
				return PortRange{}, err
			}
			end, err := strconv.Atoi(strings.TrimSpace(endText))
			if err != nil {
				return PortRange{}, err
			}
			if start <= 0 || end <= 0 || start > end || end > 65535 {
				return PortRange{}, fmt.Errorf("invalid port span %q", part)
			}
			for p := start; p <= end; p++ {
				allowed[p] = struct{}{}
			}
			continue
		}
		port, err := strconv.Atoi(part)
		if err != nil {
			return PortRange{}, err
		}
		if port <= 0 || port > 65535 {
			return PortRange{}, fmt.Errorf("invalid port %d", port)
		}
		allowed[port] = struct{}{}
	}
	if len(allowed) == 0 {
		return PortRange{}, fmt.Errorf("empty port range")
	}

	values := make([]int, 0, len(allowed))
	for port := range allowed {
		values = append(values, port)
	}
	sort.Ints(values)
	return PortRange{
		allowed: allowed,
		min:     values[0],
		max:     values[len(values)-1],
		label:   value,
		full:    len(allowed) == 65535,
	}, nil
}

func (r PortRange) Contains(port int) bool {
	if len(r.allowed) == 0 {
		return true
	}
	_, ok := r.allowed[port]
	return ok
}

func (r PortRange) String() string {
	if r.label != "" {
		return r.label
	}
	return fmt.Sprintf("%d-%d", r.min, r.max)
}

func (r PortRange) IsFull() bool {
	return r.full
}
