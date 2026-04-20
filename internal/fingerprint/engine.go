package fingerprint

import (
	"context"
	"sort"
	"strings"
	"time"

	"expr/internal/model"
	"expr/internal/probe"
)

type Engine struct {
	activeProbe bool
	timeout     time.Duration
}

func NewEngine(activeProbe bool, timeout time.Duration) Engine {
	return Engine{activeProbe: activeProbe, timeout: timeout}
}

func (e Engine) Enrich(ctx context.Context, service *model.ServiceRecord) {
	fields := service.Banner.Fields
	if fields == nil {
		fields = make(map[string]string)
		service.Banner.Fields = fields
	}

	switch service.Service {
	case "qdiscover":
		service.Banner.Product = detectQNAP(fields)
		service.Banner.Depth = scoreFields(fields, "accessType", "accessPort", "model", "displayModel", "fwVer", "fwBuildNum")
	case "device-info":
		service.Banner.Product = coalesce(fields["model"], "device-info")
		service.Banner.Depth = scoreFields(fields, "model")
	case "http":
		service.Banner.Product = "http"
		service.Banner.Depth = scoreFields(fields, "path")
	case "smb":
		service.Banner.Product = "smb"
		service.Banner.Depth = 1
	case "afpovertcp":
		service.Banner.Product = "afp"
		service.Banner.Depth = 1
	case "workstation":
		service.Banner.Product = "workstation"
		service.Banner.Depth = 1
	default:
		service.Banner.Product = service.Service
	}

	if e.activeProbe {
		e.activeProbeHTTP(ctx, service)
	}

	service.Banner.Summary = summarizeFields(fields)
}

func (e Engine) activeProbeHTTP(ctx context.Context, service *model.ServiceRecord) {
	if service.Service != "http" || service.IP == "" || service.Port <= 0 {
		return
	}

	path := service.Path
	if path == "" {
		path = "/"
	}
	result, err := probe.HTTP(ctx, e.timeout, service.IP, service.Port, path)
	if err != nil {
		return
	}

	fields := service.Banner.Fields
	if result.Scheme != "" {
		fields["scheme"] = result.Scheme
	}
	if result.Server != "" {
		fields["server"] = result.Server
	}
	if result.Title != "" {
		fields["title"] = result.Title
	}
	if result.Status != "" {
		fields["status"] = result.Status
	}
	if result.Location != "" {
		fields["location"] = result.Location
	}
	service.Banner.Depth += scoreFields(fields, "scheme", "server", "title", "status")
}

func detectQNAP(fields map[string]string) string {
	if value := coalesce(fields["displayModel"], fields["model"]); value != "" {
		return value
	}
	return "qnap"
}

func scoreFields(fields map[string]string, keys ...string) int {
	score := 0
	for _, key := range keys {
		if strings.TrimSpace(fields[key]) != "" {
			score++
		}
	}
	if score == 0 {
		return 1
	}
	return score
}

func summarizeFields(fields map[string]string) string {
	if len(fields) == 0 {
		return ""
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+fields[key])
	}
	return strings.Join(parts, ",")
}

func coalesce(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
