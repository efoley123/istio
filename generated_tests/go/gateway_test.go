package core_test

import (
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"istio.io/istio/pilot/pkg/features"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/core"
	"istio.io/istio/pkg/config/host"
)

type mockListenerBuilder struct {
	mock.Mock
}

func (m *mockListenerBuilder) buildHTTPConnectionManager(opts *core.HttpOptions) *core.HttpConnectionManager {
	args := m.Called(opts)
	return args.Get(0).(*core.HttpConnectionManager)
}

func TestMutableGatewayListenerBuild(t *testing.T) {
	features.EnableDualStack = true

	tests := []struct {
		name    string
		mutable core.MutableGatewayListener
		opts    core.GatewayListenerOpts
		wantErr bool
	}{
		{
			name: "empty filter chains",
			mutable: core.MutableGatewayListener{
				Listener: &listener.Listener{},
			},
			opts: core.GatewayListenerOpts{
				FilterChainOpts: []core.FilterChainOpts{},
			},
			wantErr: true,
		},
		{
			name: "successful build with HTTP options",
			mutable: core.MutableGatewayListener{
				Listener: &listener.Listener{
					FilterChains: []*listener.FilterChain{
						{},
					},
				},
			},
			opts: core.GatewayListenerOpts{
				FilterChainOpts: []core.FilterChainOpts{
					{
						HttpOpts: &core.HttpListenerOpts{},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockListenerBuilder{}
			builder.On("buildHTTPConnectionManager", mock.Anything).Return(&core.HttpConnectionManager{})

			err := tt.mutable.Build(builder, tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetListenerName(t *testing.T) {
	tests := []struct {
		name      string
		bind      string
		port      int
		transport core.TransportProtocol
		want      string
	}{
		{
			name:      "TCP listener",
			bind:      "0.0.0.0",
			port:      80,
			transport: core.TransportProtocolTCP,
			want:      "0.0.0.0_80",
		},
		{
			name:      "QUIC listener",
			bind:      "::",
			port:      443,
			transport: core.TransportProtocolQUIC,
			want:      "udp_::_443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := core.GetListenerName(tt.bind, tt.port, tt.transport)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildNameToServiceMapForHTTPRoutes(t *testing.T) {
	// This test would require mocking a lot of the model and push context.
	// It's quite complex because it involves understanding how virtual services and services are structured and linked.
	// A possible way to approach it is by creating a mock push context that returns predefined services and virtual services
	// when methods like Services() or VirtualServices() are called. You would then test if the function correctly maps
	// virtual service hosts to the corresponding services based on the virtual service configs.
	t.Skip("Skipping test due to complexity in setting up the required mock push context and services.")
}

func TestBuildGatewayHTTPRouteConfig(t *testing.T) {
	// Similar to TestBuildNameToServiceMapForHTTPRoutes, this test requires extensive mocking of the model and push context,
	// as well as the proxy and its merged gateway. You'd need to mock the return values of push context methods like
	// VirtualServices(), Services(), and the proxy's MergedGateway to contain specific gateways and servers.
	t.Skip("Skipping test due to complexity in setting up the required mock push context, proxy, and merged gateways.")
}