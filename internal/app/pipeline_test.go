package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"expr/internal/cli"
	"expr/internal/enrich"
	"expr/internal/fingerprint"
	"expr/internal/model"
	"expr/internal/output"
	"expr/internal/util"
)

func TestFixtureFullOutput(t *testing.T) {
	t.Parallel()

	cfg := mustConfig(t, "192.168.1.0/24", "1-65535", false)
	discovery := loadDiscoveryFixture(t, "qnap_discovery.json")
	result := runFixturePipeline(t, cfg, discovery)

	assertGoldenText(t, "qnap_full.txt", result)
	assertGoldenJSON(t, "qnap_full.json", result)

	qdiscover := findService(t, result.Services, "qdiscover", 5000)
	if qdiscover.Banner.Depth < 6 {
		t.Fatalf("expected qdiscover banner depth >= 6, got %d", qdiscover.Banner.Depth)
	}
	if got := qdiscover.Banner.Fields["displayModel"]; got != "TS-464C" {
		t.Fatalf("expected displayModel TS-464C, got %q", got)
	}
}

func TestFixtureFilteredOutput(t *testing.T) {
	t.Parallel()

	cfg := mustConfig(t, "192.168.1.0/24", "445,5000,548", false)
	discovery := loadDiscoveryFixture(t, "qnap_discovery.json")
	result := runFixturePipeline(t, cfg, discovery)

	assertGoldenText(t, "qnap_filtered.txt", result)
	assertGoldenJSON(t, "qnap_filtered.json", result)

	if len(result.Services) != 4 {
		t.Fatalf("expected 4 services after filtering, got %d", len(result.Services))
	}
}

func runFixturePipeline(t *testing.T, cfg cli.Config, discovery model.DiscoveryResult) model.ScanResult {
	t.Helper()

	services := enrich.BuildServices(cfg, discovery)
	engine := fingerprint.NewEngine(cfg.ActiveProbe, cfg.Timeout)
	for i := range services {
		engine.Enrich(context.Background(), &services[i])
	}

	assets := enrich.MergeAssets(services)
	return enrich.BuildScanResult(cfg, discovery, services, assets)
}

func mustConfig(t *testing.T, cidrText string, portsText string, activeProbe bool) cli.Config {
	t.Helper()

	cidr, err := util.ParseCIDROptional(cidrText)
	if err != nil {
		t.Fatalf("parse cidr: %v", err)
	}
	ports, err := util.ParsePortRange(portsText)
	if err != nil {
		t.Fatalf("parse ports: %v", err)
	}

	return cli.Config{
		CIDR:        cidr,
		CIDRText:    cidrText,
		PortRange:   ports,
		Timeout:     3 * time.Second,
		ActiveProbe: activeProbe,
	}
}

func loadDiscoveryFixture(t *testing.T, name string) model.DiscoveryResult {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "samples", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	var discovery model.DiscoveryResult
	if err := json.Unmarshal(data, &discovery); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return discovery
}

func assertGoldenText(t *testing.T, goldenName string, result model.ScanResult) {
	t.Helper()

	var buf bytes.Buffer
	if err := output.WriteText(&buf, result); err != nil {
		t.Fatalf("write text: %v", err)
	}
	assertGolden(t, goldenName, buf.Bytes())
}

func assertGoldenJSON(t *testing.T, goldenName string, result model.ScanResult) {
	t.Helper()

	var buf bytes.Buffer
	if err := output.WriteJSON(&buf, result); err != nil {
		t.Fatalf("write json: %v", err)
	}

	path := filepath.Join("..", "..", "testdata", "golden", goldenName)
	wantData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenName, err)
	}

	var want any
	if err := json.Unmarshal(wantData, &want); err != nil {
		t.Fatalf("unmarshal golden %s: %v", goldenName, err)
	}

	var got any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output %s: %v", goldenName, err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("json golden mismatch for %s\n\nwant:\n%s\n\ngot:\n%s", goldenName, wantData, buf.Bytes())
	}
}

func assertGolden(t *testing.T, goldenName string, got []byte) {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "golden", goldenName)
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenName, err)
	}
	if normalizeNewlines(string(got)) != normalizeNewlines(string(want)) {
		t.Fatalf("golden mismatch for %s\n\nwant:\n%s\n\ngot:\n%s", goldenName, want, got)
	}
}

func findService(t *testing.T, services []model.ServiceRecord, name string, port int) model.ServiceRecord {
	t.Helper()

	for _, service := range services {
		if service.Service == name && service.Port == port {
			return service
		}
	}
	t.Fatalf("service %s/%d not found", name, port)
	return model.ServiceRecord{}
}

func normalizeNewlines(value string) string {
	return strings.ReplaceAll(value, "\r\n", "\n")
}
