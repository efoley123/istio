package core_test

import (
	"testing"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	"istio.io/istio/pilot/pkg/features"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/core"
	"istio.io/istio/pilot/pkg/networking/util"
)

func TestBuildGatewayListeners_NoGateways(t *testing.T) {
	configgen := &core.ConfigGeneratorImpl{}

	builder := &core.ListenerBuilder{
		Node: &model.Proxy{
			MergedGateway: nil, // Simulating no gateways
		},
	}

	result := configgen.BuildGatewayListeners(builder)
	if result != builder {
		t.Errorf("Expected the same builder instance returned when no gateways are present")
	}
}

func TestBuildGatewayListeners_Success(t *testing.T) {
	configgen := &core.ConfigGeneratorImpl{}
	builder := createTestListenerBuilder()

	result := configgen.BuildGatewayListeners(builder)
	if len(result.GatewayListeners) == 0 {
		t.Errorf("Expected non-empty list of gateway listeners")
	}
}

func createTestListenerBuilder() *core.ListenerBuilder {
	gateway := &model.MergedGateway{ // Minimal setup for testing purposes
		ServerPorts: []model.ServerPort{
			{
				Number: 80,
				Protocol: "HTTP",
				Bind: "",
			},
		},
	}

	node := &model.Proxy{
		ID: "test-proxy",
		Metadata: &model.NodeMetadata{
			Labels: map[string]string{
				"some-label": "some-value",
			},
			InstanceIPs: []string{"127.0.0.1"},
		},
		MergedGateway: gateway,
	}

	pushContext := model.NewPushContext()
	pushContext.DefaultConfig = &model.ProxyConfig{}

	return &core.ListenerBuilder{
		Node: node,
		Push: pushContext,
	}
}

func TestMutableGatewayListener_Build_WithTLSContext(t *testing.T) {
	features.EnableProtocolSniffingForOutbound = true
	defer func() { features.EnableProtocolSniffingForOutbound = false }()
	listener := &core.MutableGatewayListener{
		Listener: &listener.Listener{
			Name: "http-listener",
			FilterChains: []*listener.FilterChain{
				{
					Filters: nil,
				},
			},
			TrafficDirection: core.TrafficDirection_OUTBOUND,
		},
	}
	builder := &core.ListenerBuilder{}
	opts := core.GatewayListenerOpts{
		FilterChainOpts: []*core.FilterChainOpts{
			{
				HTTPOpts: &core.HTTPOptions{
					RDS:          "http-routes",
					StatPrefix:   "http",
					RouteConfig:  &route.RouteConfiguration{},
					HTTP3Enabled: false,
				},
				TLSContext: &tls.DownstreamTlsContext{},
			},
		},
		Port: 8080,
	}

	err := listener.Build(builder, opts)
	if err != nil {
		t.Fatalf("Failed to build listener: %v", err)
	}

	if len(listener.Listener.FilterChains) == 0 {
		t.Error("Expected listener to have filter chains, got none")
	}
}

func TestBuildGatewayHTTPRouteConfig_NoGateways(t *testing.T) {
	configgen := &core.ConfigGeneratorImpl{}
	node := &model.Proxy{
		MergedGateway: nil, // Simulating no gateways
	}

	routeName := "http.80"
	pushContext := model.NewPushContext()

	routeConfig := configgen.BuildGatewayHTTPRouteConfig(node, pushContext, routeName)

	if routeConfig.Name != routeName {
		t.Errorf("Expected route configuration name to be %s, got %s", routeName, routeConfig.Name)
	}

	if len(routeConfig.VirtualHosts) != 0 {
		t.Errorf("Expected no virtual hosts, got %d", len(routeConfig.VirtualHosts))
	}
}

func TestBuildGatewayHTTPRouteConfig_Success(t *testing.T) {
	configgen := &core.ConfigGeneratorImpl{}
	node := createTestProxyWithMergedGateway()

	routeName := "http.80"
	pushContext := model.NewPushContext()

	routeConfig := configgen.BuildGatewayHTTPRouteConfig(node, pushContext, routeName)

	if routeConfig.Name != routeName {
		t.Errorf("Expected route configuration name to be %s, got %s", routeName, routeConfig.Name)
	}

	// Add more assertions as necessary to validate successful route configurations
}

func createTestProxyWithMergedGateway() *model.Proxy {
	gateway := &model.MergedGateway{ // Minimal setup for testing purposes
		// Fill in required fields
	}

	return &model.Proxy{
		ID:           "test-proxy-gateway",
		MergedGateway: gateway,
	}
}