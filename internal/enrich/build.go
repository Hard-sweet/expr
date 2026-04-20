package enrich

import (
	"sort"
	"strings"

	"expr/internal/cli"
	"expr/internal/model"
	"expr/internal/util"
)

func BuildServices(cfg cli.Config, discovery model.DiscoveryResult) []model.ServiceRecord {
	var services []model.ServiceRecord
	for _, entry := range discovery.Entries {
		port := entry.Port
		if port <= 0 && !cfg.PortRange.IsFull() {
			continue
		}
		if port > 0 && !cfg.PortRange.Contains(port) {
			continue
		}

		ip := pickFirstIP(entry.IPv4, cfg.CIDR)
		if cfg.CIDR != nil && ip == "" {
			continue
		}
		if ip == "" && len(entry.IPv4) > 0 {
			ip = entry.IPv4[0]
		}

		serviceName := normalizeServiceName(entry.Service)
		record := model.ServiceRecord{
			Name:        entry.Instance,
			IP:          ip,
			IPv6:        append([]string(nil), entry.IPv6...),
			Port:        port,
			Proto:       "tcp",
			Service:     serviceName,
			ServiceFQDN: entry.Service,
			Hostname:    entry.HostName,
			TTL:         entry.TTL,
			SourceText:  append([]string(nil), entry.Text...),
			Banner: model.Banner{
				Service: serviceName,
				Fields:  copyMap(entry.TextKV),
				Raw:     append([]string(nil), entry.Text...),
			},
		}
		if path, ok := entry.TextKV["path"]; ok {
			record.Path = path
		}
		services = append(services, record)
	}

	sort.Slice(services, func(i, j int) bool {
		if services[i].Port != services[j].Port {
			return services[i].Port < services[j].Port
		}
		return services[i].Service < services[j].Service
	})
	return services
}

func MergeAssets(services []model.ServiceRecord) []model.Asset {
	byHost := make(map[string]*model.Asset)
	order := make([]string, 0)

	for _, service := range services {
		key := service.IP
		if key == "" {
			key = service.Hostname
		}
		if key == "" {
			key = service.Name
		}
		if key == "" {
			continue
		}
		asset, ok := byHost[key]
		if !ok {
			asset = &model.Asset{
				PrimaryIP: service.IP,
				Hostname:  service.Hostname,
			}
			byHost[key] = asset
			order = append(order, key)
		}
		if asset.PrimaryIP == "" {
			asset.PrimaryIP = service.IP
		}
		if asset.Hostname == "" {
			asset.Hostname = service.Hostname
		}
		asset.Services = append(asset.Services, service)
	}

	assets := make([]model.Asset, 0, len(order))
	for _, key := range order {
		assets = append(assets, *byHost[key])
	}
	return assets
}

func BuildScanResult(cfg cli.Config, discovery model.DiscoveryResult, services []model.ServiceRecord, assets []model.Asset) model.ScanResult {
	return model.ScanResult{
		Scope: model.Scope{
			CIDR:      cfg.CIDRText,
			Ports:     cfg.PortRange.String(),
			Interface: discovery.Interface,
			Timeout:   cfg.Timeout.String(),
		},
		Services:     services,
		Assets:       assets,
		Answers:      discovery.Answers,
		ServiceTypes: discovery.ServiceTypes,
	}
}

func normalizeServiceName(service string) string {
	trimmed := strings.TrimSuffix(service, ".local")
	trimmed = strings.TrimPrefix(trimmed, "_")
	trimmed = strings.TrimSuffix(trimmed, "._tcp")
	trimmed = strings.TrimSuffix(trimmed, "._udp")
	return trimmed
}

func pickFirstIP(values []string, cidr *util.IPNetAlias) string {
	if len(values) == 0 {
		return ""
	}
	if cidr == nil {
		return values[0]
	}
	for _, value := range values {
		if cidr.ContainsString(value) {
			return value
		}
	}
	return ""
}

func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
