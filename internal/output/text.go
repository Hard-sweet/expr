package output

import (
	"fmt"
	"io"
	"strings"

	"expr/internal/model"
)

func WriteText(w io.Writer, result model.ScanResult) error {
	if _, err := fmt.Fprintln(w, "services:"); err != nil {
		return err
	}
	for _, service := range result.Services {
		if err := writeService(w, service); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, "answers:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "PTR:"); err != nil {
		return err
	}
	for _, answer := range result.Answers.PTR {
		if _, err := fmt.Fprintln(w, answer); err != nil {
			return err
		}
	}
	return nil
}

func writeService(w io.Writer, service model.ServiceRecord) error {
	if service.Port > 0 {
		if _, err := fmt.Fprintf(w, "%d/%s %s:\n", service.Port, service.Proto, service.Service); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "%s:\n", service.Service); err != nil {
			return err
		}
	}

	lines := []string{
		fmt.Sprintf("Name=%s", service.Name),
	}
	if service.IP != "" {
		lines = append(lines, fmt.Sprintf("IPv4=%s", service.IP))
	}
	if len(service.IPv6) > 0 {
		lines = append(lines, fmt.Sprintf("IPv6=%s", strings.Join(service.IPv6, ",")))
	}
	if service.Hostname != "" {
		lines = append(lines, fmt.Sprintf("Hostname=%s", service.Hostname))
	}
	if service.TTL > 0 {
		lines = append(lines, fmt.Sprintf("TTL=%d", service.TTL))
	}
	if service.Path != "" {
		lines = append(lines, fmt.Sprintf("path=%s", service.Path))
	}
	if service.Banner.Summary != "" && service.Banner.Summary != fmt.Sprintf("path=%s", service.Path) {
		lines = append(lines, service.Banner.Summary)
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}
