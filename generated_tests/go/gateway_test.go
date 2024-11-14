package core_test

import (
	"testing"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/core"
	networking "istio.io/api/networking/v1alpha3"
	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/istio/pkg/config/host"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestBuildGatewayListenerTLSContext(t *testing.T) {
	tests := []struct {
		name      string
		server    *networking.Server
		transport core.TransportProtocol
		want      *tls.DownstreamTlsContext
	}{
		{
			name: "simple TLS mode",
			server: &networking.Server{
				Tls: &networking.ServerTLSSettings{
					Mode: networking.ServerTLSSettings_SIMPLE,
				},
			},
			transport: core.TransportProtocolTCP,
			want: &tls.DownstreamTlsContext{
				CommonTlsContext: &tls.CommonTlsContext{},
			},
		},
		{
			name: "nil TLS settings",
			server: &networking.Server{
				Tls: nil,
			},
			transport: core.TransportProtocolTCP,
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := core.BuildListenerTLSContext(tt.server.Tls, &model.Proxy{}, &meshconfig.MeshConfig{}, tt.transport, true)
			if !proto.Equal(got, tt.want) {
				t.Errorf("BuildListenerTLSContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPortMatch(t *testing.T) {
	tests := []struct {
		name      string
		port      uint32
		server    *networking.Server
		wantMatch bool
	}{
		{
			name: "Matching port",
			port: 80,
			server: &networking.Server{
				Port: &networking.Port{
					Number: 80,
				},
			},
			wantMatch: true,
		},
		{
			name: "Non-matching port",
			port: 80,
			server: &networking.Server{
				Port: &networking.Port{
					Number: 8080,
				},
			},
			wantMatch: false,
		},
		{
			name: "Wildcard port match",
			port: 0,
			server: &networking.Server{
				Port: &networking.Port{
					Number: 80,
				},
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := core.IsPortMatch(tt.port, tt.server); got != tt.wantMatch {
				t.Errorf("IsPortMatch() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestIsGatewayMatch(t *testing.T) {
	tests := []struct {
		name          string
		gateway       string
		gatewayNames  []string
		wantMatch     bool
	}{
		{
			name:          "Matching gateway",
			gateway:       "gateway-1",
			gatewayNames:  []string{"gateway-1"},
			wantMatch:     true,
		},
		{
			name:          "Non-matching gateway",
			gateway:       "gateway-1",
			gatewayNames:  []string{"gateway-2"},
			wantMatch:     false,
		},
		{
			name:          "Wildcard gateway match",
			gateway:       "gateway-1",
			gatewayNames:  nil, // nil implies wildcard
			wantMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := core.IsGatewayMatch(tt.gateway, tt.gatewayNames); got != tt.wantMatch {
				t.Errorf("IsGatewayMatch() = %v, wantMatch %v", got, tt.wantMatch)
			}
		})
	}
}

func TestPickMatchingGatewayHosts(t *testing.T) {
	tests := []struct {
		name        string
		serverHosts []string
		virtualService config.Config
		wantMatchingHosts map[string]host.Name
	}{
		{
			name:        "Matching hosts",
			serverHosts: []string{"example.com"},
			virtualService: config.Config{
				ConfigMeta: model.ConfigMeta{},
				Spec: &networking.VirtualService{
					Hosts: []string{"example.com"},
				},
			},
			wantMatchingHosts: map[string]host.Name{
				"example.com": "example.com",
			},
		},
		{
			name:        "Non-matching hosts",
			serverHosts: []string{"example.com"},
			virtualService: config.Config{
				ConfigMeta: model.ConfigMeta{},
				Spec: &networking.VirtualService{
					Hosts: []string{"test.com"},
				},
			},
			wantMatchingHosts: make(map[string]host.Name),
		},
		{
			name:        "Wildcard hosts",
			serverHosts: []string{"*"},
			virtualService: config.Config{
				ConfigMeta: model.ConfigMeta{},
				Spec: &networking.VirtualService{
					Hosts: []string{"example.com"},
				},
			},
			wantMatchingHosts: map[string]host.Name{
				"example.com": "*",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverHosts := util.NewStringSet(tt.serverHosts...)
			got := core.PickMatchingGatewayHosts(serverHosts, tt.virtualService)
			if !reflect.DeepEqual(got, tt.wantMatchingHosts) {
				t.Errorf("PickMatchingGatewayHosts() got = %v, want %v", got, tt.wantMatchingHosts)
			}
		})
	}
}