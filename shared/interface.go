// Package shared contains shared data between the host and plugins.
package shared

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Handshake is used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"customPlugin": &CustomPlugin{},
}

// Service is the interface that we're exposing as a plugin.
type Service interface {
	Start(args ...interface{})
	Stop()
}

// PluginRPC is an implementation that talks over RPC
type PluginRPC struct{ client *rpc.Client }

// PluginRPCServer is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type PluginRPCServer struct {
	// This is the real implementation
	Impl Service
}

func (s *PluginRPCServer) Start(args interface{}, resp *string) error {
	s.Impl.Start(args)
	*resp = "Plugin started"
	return nil
}

// CustomPlugin This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a GreeterRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return GreeterRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type CustomPlugin struct {
	// CustomPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Service
}

func (p *CustomPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (CustomPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginRPC{client: c}, nil
}
