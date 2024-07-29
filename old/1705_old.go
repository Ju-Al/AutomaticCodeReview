package aws

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/spiffe/spire/pkg/common/catalog"
	caws "github.com/spiffe/spire/pkg/common/plugin/aws"
	"github.com/spiffe/spire/pkg/common/telemetry"
	"github.com/spiffe/spire/pkg/server/plugin/noderesolver"
	spi "github.com/spiffe/spire/proto/spire/common/plugin"
)

func BuiltIn() catalog.Plugin {
	return builtin(New())
}

func builtin(p *IIDResolverPlugin) catalog.Plugin {
	return catalog.MakePlugin(caws.PluginName,
		noderesolver.PluginServer(p),
	)
}

// IIDResolverPlugin implements node resolution for agents running in aws.
type IIDResolverPlugin struct {
	log hclog.Logger
}

// New creates a new IIDResolverPlugin.
func New() *IIDResolverPlugin {
	return &IIDResolverPlugin{}
}

func (p *IIDResolverPlugin) SetLogger(log hclog.Logger) {
	p.log = log
}

// Configure configures the IIDResolverPlugin
func (p *IIDResolverPlugin) Configure(ctx context.Context, req *spi.ConfigureRequest) (*spi.ConfigureResponse, error) {
	// Parse HCL config payload into config struct
	config := new(caws.SessionConfig)
	if err := hcl.Decode(config, req.Configuration); err != nil {
		return nil, iidError.New("unable to decode configuration: %v", err)
	}

	// Set defaults from the environment
	if config.AccessKeyID == "" {
		config.AccessKeyID = p.hooks.getenv(caws.AccessKeyIDVarName)
	}
	if config.SecretAccessKey == "" {
		config.SecretAccessKey = p.hooks.getenv(caws.SecretAccessKeyVarName)
	}

	switch {
	case config.AccessKeyID != "" && config.SecretAccessKey == "":
		return nil, iidError.New("configuration missing secret access key, but has access key id")
	case config.AccessKeyID == "" && config.SecretAccessKey != "":
		return nil, iidError.New("configuration missing access key id, but has secret access key")
	}

	// set the AWS configuration and reset clients
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
	p.clients = make(map[string]awsClient)
	return &spi.ConfigureResponse{}, nil
}

// GetPluginInfo returns the version and related metadata of the installed plugin.
func (p *IIDResolverPlugin) GetPluginInfo(context.Context, *spi.GetPluginInfoRequest) (*spi.GetPluginInfoResponse, error) {
	return &spi.GetPluginInfoResponse{}, nil
}

// Resolve handles the given resolve request
func (p *IIDResolverPlugin) Resolve(ctx context.Context, req *noderesolver.ResolveRequest) (*noderesolver.ResolveResponse, error) {
	return &noderesolver.ResolveResponse{}, nil
}
