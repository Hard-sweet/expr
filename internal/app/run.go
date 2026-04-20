package app

import (
	"context"
	"fmt"
	"io"

	"expr/internal/cli"
	"expr/internal/discover/mdns"
	"expr/internal/enrich"
	"expr/internal/fingerprint"
	"expr/internal/output"
)

func Run(ctx context.Context, cfg cli.Config, w io.Writer) error {
	discoverer := mdns.NewDiscoverer(cfg.Interface, cfg.Timeout)
	result, err := discoverer.Discover(ctx)
	if err != nil {
		return fmt.Errorf("discover mdns: %w", err)
	}

	services := enrich.BuildServices(cfg, result)
	engine := fingerprint.NewEngine(cfg.ActiveProbe, cfg.Timeout)
	for i := range services {
		engine.Enrich(ctx, &services[i])
	}

	assets := enrich.MergeAssets(services)
	scan := enrich.BuildScanResult(cfg, result, services, assets)

	if cfg.JSON {
		return output.WriteJSON(w, scan)
	}
	return output.WriteText(w, scan)
}
