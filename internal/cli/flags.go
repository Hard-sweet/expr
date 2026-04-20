package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"expr/internal/util"
)

type Config struct {
	CIDR        *util.IPNetAlias
	CIDRText    string
	PortRange   util.PortRange
	Interface   string
	Timeout     time.Duration
	JSON        bool
	ActiveProbe bool
}

func Parse(args []string) (Config, error) {
	var cfg Config

	fs := flag.NewFlagSet("mdnsmap", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})

	var (
		cidrText  string
		portsText string
		iface     string
		timeout   time.Duration
		jsonOut   bool
		active    bool
	)

	fs.StringVar(&cidrText, "cidr", "", "CIDR filter, for example 192.168.1.0/24")
	fs.StringVar(&portsText, "ports", "1-65535", "port range filter, for example 1-1024 or 80,443,445,5000")
	fs.StringVar(&iface, "iface", "", "network interface name for mDNS discovery")
	fs.DurationVar(&timeout, "timeout", 5*time.Second, "discovery timeout")
	fs.BoolVar(&jsonOut, "json", false, "emit JSON instead of text")
	fs.BoolVar(&active, "active-probe", true, "run active probes for known services")

	if err := fs.Parse(args); err != nil {
		return Config{}, usageError(err)
	}

	ipNet, err := util.ParseCIDROptional(cidrText)
	if err != nil {
		return Config{}, fmt.Errorf("invalid -cidr: %w", err)
	}
	pr, err := util.ParsePortRange(portsText)
	if err != nil {
		return Config{}, fmt.Errorf("invalid -ports: %w", err)
	}

	cfg = Config{
		CIDR:        ipNet,
		CIDRText:    cidrText,
		PortRange:   pr,
		Interface:   strings.TrimSpace(iface),
		Timeout:     timeout,
		JSON:        jsonOut,
		ActiveProbe: active,
	}
	return cfg, nil
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func usageError(err error) error {
	var b strings.Builder
	b.WriteString(err.Error())
	b.WriteString("\n\nUsage:\n")
	b.WriteString("  ")
	b.WriteString(os.Args[0])
	b.WriteString(" -cidr 192.168.1.0/24 -ports 1-1024,5000\n")
	return fmt.Errorf("%s", b.String())
}
