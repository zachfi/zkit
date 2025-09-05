package tracing

import (
	"flag"

	"github.com/grafana/dskit/grpcclient"

	"github.com/zachfi/zkit/pkg/util"
)

type Config struct {
	OtelEndpoint     string            `yaml:"otel_endpoint"`
	OrgID            string            `yaml:"org_id"`
	GRPCClientConfig grpcclient.Config `yaml:"grpc_client_config"`
}

func (c *Config) RegisterFlagsAndApplyDefaults(prefix string, f *flag.FlagSet) {
	f.StringVar(&c.OtelEndpoint, util.PrefixConfig(prefix, "otel.endpoint"), "", "otel endpoint, eg: tempo:4317")
	f.StringVar(&c.OrgID, util.PrefixConfig(prefix, "org.id"), "", "org ID to use when sending traces")

	c.GRPCClientConfig.RegisterFlagsWithPrefix("otel.client", f)
}
