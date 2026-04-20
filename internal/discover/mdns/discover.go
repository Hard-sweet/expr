package mdns

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"

	"expr/internal/model"
)

type Discoverer struct {
	ifaceName string
	timeout   time.Duration
}

func NewDiscoverer(ifaceName string, timeout time.Duration) Discoverer {
	return Discoverer{ifaceName: ifaceName, timeout: timeout}
}

func (d Discoverer) Discover(ctx context.Context) (model.DiscoveryResult, error) {
	iface, err := lookupInterface(d.ifaceName)
	if err != nil {
		return model.DiscoveryResult{}, err
	}

	serviceTypes, err := d.browseTypes(ctx, iface)
	if err != nil {
		return model.DiscoveryResult{}, err
	}

	result := model.DiscoveryResult{
		ServiceTypes: serviceTypes,
		ObservedAt:   time.Now(),
	}
	if iface != nil {
		result.Interface = iface.Name
	}
	result.Answers.PTR = append(result.Answers.PTR, serviceTypes...)

	var (
		mu      sync.Mutex
		entries = make(map[string]model.MDNSEntry)
		wg      sync.WaitGroup
		errs    = make(chan error, len(serviceTypes))
	)

	for _, serviceType := range serviceTypes {
		wg.Add(1)
		go func(serviceType string) {
			defer wg.Done()
			found, err := d.browseService(ctx, iface, serviceType)
			if err != nil {
				errs <- err
				return
			}
			mu.Lock()
			for _, item := range found {
				key := item.Instance + "|" + item.Service + "|" + item.HostName + "|" + fmt.Sprint(item.Port)
				entries[key] = item
			}
			mu.Unlock()
		}(serviceType)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return model.DiscoveryResult{}, err
		}
	}

	result.Entries = make([]model.MDNSEntry, 0, len(entries))
	for _, item := range entries {
		result.Entries = append(result.Entries, item)
	}
	sort.Slice(result.Entries, func(i, j int) bool {
		if result.Entries[i].Port != result.Entries[j].Port {
			return result.Entries[i].Port < result.Entries[j].Port
		}
		return result.Entries[i].Service < result.Entries[j].Service
	})
	return result, nil
}

func (d Discoverer) browseTypes(ctx context.Context, iface *net.Interface) ([]string, error) {
	entries, err := d.browse(ctx, iface, "_services._dns-sd._udp", "local.")
	if err != nil {
		return nil, err
	}
	types := make(map[string]struct{})
	for _, entry := range entries {
		for _, txt := range entry.Text {
			if strings.HasSuffix(txt, ".local") {
				types[ensureLocalSuffix(txt)] = struct{}{}
			}
		}
		if entry.Instance != "" && strings.HasPrefix(entry.Instance, "_") {
			types[ensureLocalSuffix(entry.Instance)] = struct{}{}
		}
	}
	if len(types) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(types))
	for item := range types {
		out = append(out, item)
	}
	sort.Strings(out)
	return out, nil
}

func (d Discoverer) browseService(ctx context.Context, iface *net.Interface, serviceType string) ([]model.MDNSEntry, error) {
	trimmed := strings.TrimSuffix(serviceType, ".local")
	entries, err := d.browse(ctx, iface, trimmed, "local.")
	if err != nil {
		return nil, err
	}

	out := make([]model.MDNSEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, normalizeEntry(entry, serviceType, iface))
	}
	return out, nil
}

func (d Discoverer) browse(ctx context.Context, iface *net.Interface, service, domain string) ([]*zeroconf.ServiceEntry, error) {
	resolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces(interfacesOrNil(iface)))
	if err != nil {
		return nil, fmt.Errorf("create resolver: %w", err)
	}

	browseCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	entriesCh := make(chan *zeroconf.ServiceEntry)
	var out []*zeroconf.ServiceEntry
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case entry, ok := <-entriesCh:
				if !ok {
					return
				}
				out = append(out, entry)
			case <-browseCtx.Done():
				return
			}
		}
	}()

	if err := resolver.Browse(browseCtx, service, domain, entriesCh); err != nil {
		wg.Wait()
		return nil, fmt.Errorf("browse %s: %w", service, err)
	}

	<-browseCtx.Done()
	wg.Wait()

	if browseCtx.Err() != nil && browseCtx.Err() != context.DeadlineExceeded && browseCtx.Err() != context.Canceled {
		return nil, browseCtx.Err()
	}
	return out, nil
}

func normalizeEntry(entry *zeroconf.ServiceEntry, serviceType string, iface *net.Interface) model.MDNSEntry {
	item := model.MDNSEntry{
		Instance: entry.Instance,
		Service:  serviceType,
		Domain:   entry.Domain,
		HostName: strings.TrimSuffix(entry.HostName, "."),
		Port:     entry.Port,
		Text:     append([]string(nil), entry.Text...),
		TextKV:   parseTextPairs(entry.Text),
		TTL:      uint32(entry.TTL),
	}
	if iface != nil {
		item.Interface = iface.Name
	}
	for _, ip := range entry.AddrIPv4 {
		item.IPv4 = append(item.IPv4, ip.String())
	}
	for _, ip := range entry.AddrIPv6 {
		item.IPv6 = append(item.IPv6, ip.String())
	}
	return item
}

func parseTextPairs(lines []string) map[string]string {
	out := make(map[string]string)
	for _, line := range lines {
		k, v, ok := strings.Cut(line, "=")
		if ok && k != "" {
			out[k] = v
		}
	}
	return out
}

func lookupInterface(name string) (*net.Interface, error) {
	if strings.TrimSpace(name) == "" {
		return nil, nil
	}
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("lookup interface %q: %w", name, err)
	}
	return iface, nil
}

func interfacesOrNil(iface *net.Interface) []net.Interface {
	if iface == nil {
		return nil
	}
	return []net.Interface{*iface}
}

func ensureLocalSuffix(value string) string {
	value = strings.TrimSuffix(value, ".")
	if strings.HasSuffix(value, ".local") {
		return value
	}
	return value + ".local"
}
