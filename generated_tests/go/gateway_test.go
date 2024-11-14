// pilot/pkg/networking/core/gateway_test.go

package core

import (
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/types/known/anypb"

	meshconfig "istio.io/api/mesh/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	istionetworking "istio.io/istio/pilot/pkg/networking"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/host"
	"istio.io/istio/pkg/config/protocol"
	"istio.io/istio/pkg/util/sets"
)

// Mock dependencies
type MockPushContext struct {
	mockCtrl *gomock.Controller
}

func NewMockPushContext(ctrl *gomock.Controller) *MockPushContext {
	return &MockPushContext{mockCtrl: ctrl}
}

// TestMutableGatewayListener_build_Success tests the build method of MutableGatewayListener for successful scenarios.
func TestMutableGatewayListener_build_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	mockBuilder := &ListenerBuilder{
		// Initialize necessary fields
	}
	opts := gatewayListenerOpts{
		// Initialize with valid filterChainOpts
	}
	mutableListener := &MutableGatewayListener{
		Listener: &listener.Listener{
			Name:         "test_listener",
			FilterChains: []*listener.FilterChain{},
		},
	}

	// Prepare filterChainOpts
	filterChainOpts := &filterChainOpts{
		// Initialize with valid options
	}
	opts.filterChainOpts = []*filterChainOpts{filterChainOpts}

	// Create ListenerBuilder mock methods if necessary
	// ...

	// Execute
	err := mutableListener.build(mockBuilder, opts)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedFilterChains := len(opts.filterChainOpts)
	if len(mutableListener.Listener.FilterChains) != expectedFilterChains {
		t.Errorf("Expected %d filter chains, got %d", expectedFilterChains, len(mutableListener.Listener.FilterChains))
	}
}

// TestMutableGatewayListener_build_NoFilterChains tests the build method with no filterChainOpts.
func TestMutableGatewayListener_build_NoFilterChains(t *testing.T) {
	mutableListener := &MutableGatewayListener{
		Listener: &listener.Listener{
			Name:         "test_listener",
			FilterChains: []*listener.FilterChain{},
		},
	}
	opts := gatewayListenerOpts{
		filterChainOpts: []*filterChainOpts{},
	}

	err := mutableListener.build(&ListenerBuilder{}, opts)
	if err == nil {
		t.Errorf("Expected error due to no filter chains, got nil")
	}
}

// TestBuildGatewayListeners_NoGateways tests buildGatewayListeners when there are no gateways.
func TestBuildGatewayListeners_NoGateways(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := &ListenerBuilder{
		node: &model.Proxy{
			ID:           "test_proxy",
			MergedGateway: nil,
		},
		push: &model.PushContext{},
	}

	configGen := &ConfigGeneratorImpl{}
	result := configGen.buildGatewayListeners(builder)

	if result != builder {
		t.Errorf("Expected builder to be returned unchanged")
	}
}

// TestBuildGatewayListeners_WithGateways tests buildGatewayListeners with valid gateways.
func TestBuildGatewayListeners_WithGateways(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mergedGateway := &model.MergedGateway{
		ServerPorts: []networking.ServerPort{
			{
				Number:   80,
				Protocol: "HTTP",
				Bind:     "0.0.0.0",
			},
			{
				Number:   443,
				Protocol: "HTTPS",
				Bind:     "0.0.0.0",
			},
		},
		MergedServers: map[model.ServerPort]*model.MergedServers{
			// Populate with mock servers
		},
	}

	builder := &ListenerBuilder{
		node: &model.Proxy{
			ID: "test_proxy",
			MergedGateway: mergedGateway,
		},
		push: &model.PushContext{},
	}

	configGen := &ConfigGeneratorImpl{}
	result := configGen.buildGatewayListeners(builder)

	if len(result.gatewayListeners) == 0 {
		t.Errorf("Expected gateway listeners to be built")
	}
}

// TestBuildGatewayTCPBasedFilterChains_HTTP tests buildGatewayTCPBasedFilterChains for HTTP protocol.
func TestBuildGatewayTCPBasedFilterChains_HTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	push := &model.PushContext{}
	proxy := &model.Proxy{}
	mergedGateway := &model.MergedGateway{}
	lb := &ListenerBuilder{
		push:          push,
		node:          proxy,
		mergedGateway: mergedGateway,
	}
	configGen := &ConfigGeneratorImpl{}

	server := &networking.Server{
		Port: &networking.Port{
			Number:   80,
			Protocol: "HTTP",
		},
		Hosts: []string{"example.com"},
	}

	opts := &gatewayListenerOpts{}

	serversForPort := &model.MergedServers{
		Servers: []*networking.Server{server},
	}

	configGen.buildGatewayTCPBasedFilterChains(lb, protocol.HTTP, model.ServerPort{Port: 80}, opts, serversForPort, &meshconfig.ProxyConfig{}, mergedGateway, map[uint32]map[string]string{})

	if len(opts.filterChainOpts) != 1 {
		t.Errorf("Expected 1 filter chain opts, got %d", len(opts.filterChainOpts))
	}
}

// TestBuildGatewayTCPBasedFilterChains_TLSTermination tests buildGatewayTCPBasedFilterChains for TLS termination.
func TestBuildGatewayTCPBasedFilterChains_TLSTermination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	push := &model.PushContext{}
	proxy := &model.Proxy{}
	mergedGateway := &model.MergedGateway{}
	lb := &ListenerBuilder{
		push:          push,
		node:          proxy,
		mergedGateway: mergedGateway,
	}
	configGen := &ConfigGeneratorImpl{}

	server := &networking.Server{
		Port: &networking.Port{
			Number:   443,
			Protocol: "HTTPS",
		},
		Hosts: []string{"secure.example.com"},
		Tls: &networking.ServerTLSSettings{
			Mode: networking.ServerTLSSettings_SIMPLE,
		},
	}

	opts := &gatewayListenerOpts{}

	serversForPort := &model.MergedServers{
		Servers: []*networking.Server{server},
	}

	configGen.buildGatewayTCPBasedFilterChains(lb, protocol.HTTPS, model.ServerPort{Port: 443}, opts, serversForPort, &meshconfig.ProxyConfig{}, mergedGateway, map[uint32]map[string]string{})

	if len(opts.filterChainOpts) != 1 {
		t.Errorf("Expected 1 filter chain opts for TLS termination, got %d", len(opts.filterChainOpts))
	}
}

// TestBuildGatewayTCPBasedFilterChains_Passthrough tests buildGatewayTCPBasedFilterChains for passthrough TLS.
func TestBuildGatewayTCPBasedFilterChains_Passthrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	push := &model.PushContext{}
	proxy := &model.Proxy{}
	mergedGateway := &model.MergedGateway{}
	lb := &ListenerBuilder{
		push:          push,
		node:          proxy,
		mergedGateway: mergedGateway,
	}
	configGen := &ConfigGeneratorImpl{}

	server := &networking.Server{
		Port: &networking.Port{
			Number:   443,
			Protocol: "HTTPS",
		},
		Hosts: []string{"passthrough.example.com"},
		Tls: &networking.ServerTLSSettings{
			Mode: networking.ServerTLSSettings_PASSTHROUGH,
		},
	}

	opts := &gatewayListenerOpts{}

	serversForPort := &model.MergedServers{
		Servers: []*networking.Server{server},
	}

	filterChains := configGen.buildGatewayTCPBasedFilterChains(lb, protocol.HTTPS, model.ServerPort{Port: 443}, opts, serversForPort, &meshconfig.ProxyConfig{}, mergedGateway, map[uint32]map[string]string{})
	if len(filterChains) == 0 {
		t.Errorf("Expected filter chains for passthrough TLS, got none")
	}
}

// TestBuildGatewayHTTP3FilterChains_Success tests buildGatewayHTTP3FilterChains for successful HTTP/3 configuration.
func TestBuildGatewayHTTP3FilterChains_Success(t *testing.T) {
	configGen := &ConfigGeneratorImpl{}
	builder := &ListenerBuilder{
		push: &model.PushContext{},
	}
	serversForPort := &model.MergedServers{
		Servers: []*networking.Server{
			{
				Port: &networking.Port{
					Number:   443,
					Protocol: "HTTPS",
				},
				Hosts: []string{"h3.example.com"},
			},
		},
	}
	mergedGateway := &model.MergedGateway{}
	proxyConfig := &meshconfig.ProxyConfig{}
	opts := &gatewayListenerOpts{}

	configGen.buildGatewayHTTP3FilterChains(builder, serversForPort, mergedGateway, proxyConfig, opts)

	if len(opts.filterChainOpts) == 0 {
		t.Errorf("Expected filter chain opts for HTTP/3, got none")
	}
}

// TestBuildGatewayHTTPRouteConfig_NoGateways tests buildGatewayHTTPRouteConfig when there are no gateways.
func TestBuildGatewayHTTPRouteConfig_NoGateways(t *testing.T) {
	configGen := &ConfigGeneratorImpl{}
	node := &model.Proxy{
		ID:           "test_proxy",
		MergedGateway: nil,
	}
	push := &model.PushContext{}
	routeName := "test_route"

	routeConfig := configGen.buildGatewayHTTPRouteConfig(node, push, routeName)

	if routeConfig.Name != routeName {
		t.Errorf("Expected route name %s, got %s", routeName, routeConfig.Name)
	}
	if len(routeConfig.VirtualHosts) != 0 {
		t.Errorf("Expected no virtual hosts, got %d", len(routeConfig.VirtualHosts))
	}
}

// TestBuildGatewayHTTPRouteConfig_MissingRoute tests buildGatewayHTTPRouteConfig with missing route.
func TestBuildGatewayHTTPRouteConfig_MissingRoute(t *testing.T) {
	configGen := &ConfigGeneratorImpl{}
	node := &model.Proxy{
		ID: "test_proxy",
		MergedGateway: &model.MergedGateway{
			ServersByRouteName: map[string][]*networking.Server{},
		},
	}
	push := &model.PushContext{}
	routeName := "missing_route"

	routeConfig := configGen.buildGatewayHTTPRouteConfig(node, push, routeName)

	if routeConfig.Name != routeName {
		t.Errorf("Expected route name %s, got %s", routeName, routeConfig.Name)
	}
	if len(routeConfig.VirtualHosts) != 0 {
		t.Errorf("Expected no virtual hosts for missing route, got %d", len(routeConfig.VirtualHosts))
	}
}

// TestCollapseDuplicateRoutes tests the collapseDuplicateRoutes function.
func TestCollapseDuplicateRoutes(t *testing.T) {
	input := map[host.Name]*route.VirtualHost{
		"host1.example.com": {
			Name:    "vhost1",
			Domains: []string{"host1.example.com"},
			Routes:  []*route.Route{{}},
		},
		"host2.example.com": {
			Name:    "vhost2",
			Domains: []string{"host2.example.com"},
			Routes:  []*route.Route{{}},
		},
	}

	output := collapseDuplicateRoutes(input)

	if len(output) != 1 {
		t.Errorf("Expected 1 virtual host after collapsing, got %d", len(output))
	}
	for _, vhost := range output {
		if len(vhost.Domains) != 2 {
			t.Errorf("Expected domains to be merged, got %d", len(vhost.Domains))
		}
	}
}

// TestBuildGatewayListenerTLSContext_Passthrough tests buildGatewayListenerTLSContext for passthrough servers.
func TestBuildGatewayListenerTLSContext_Passthrough(t *testing.T) {
	server := &networking.Server{
		Tls: &networking.ServerTLSSettings{
			Mode: networking.ServerTLSSettings_PASSTHROUGH,
		},
	}

	mesh := &meshconfig.MeshConfig{}
	proxy := &model.Proxy{}
	transportProtocol := istionetworking.TransportProtocolTCP

	tlsContext := buildGatewayListenerTLSContext(mesh, server, proxy, transportProtocol)

	if tlsContext != nil {
		t.Errorf("Expected nil TLS context for passthrough mode, got %v", tlsContext)
	}
}

// TestBuildGatewayListenerTLSContext_SimpleTLS tests buildGatewayListenerTLSContext for simple TLS mode.
func TestBuildGatewayListenerTLSContext_SimpleTLS(t *testing.T) {
	server := &networking.Server{
		Tls: &networking.ServerTLSSettings{
			Mode: networking.ServerTLSSettings_SIMPLE,
		},
	}

	mesh := &meshconfig.MeshConfig{}
	proxy := &model.Proxy{}
	transportProtocol := istionetworking.TransportProtocolTCP

	tlsContext := buildGatewayListenerTLSContext(mesh, server, proxy, transportProtocol)

	if tlsContext == nil {
		t.Errorf("Expected TLS context for SIMPLE mode, got nil")
	}
}

// Test_isGatewayMatch tests the isGatewayMatch function.
func Test_isGatewayMatch(t *testing.T) {
	tests := []struct {
		gateway       string
		gatewayNames  []string
		expectedMatch bool
	}{
		{"gateway1", []string{"gateway1", "gateway2"}, true},
		{"gateway3", []string{"gateway1", "gateway2"}, false},
		{"gateway1", []string{}, true},
	}

	for _, tt := range tests {
		match := isGatewayMatch(tt.gateway, tt.gatewayNames)
		if match != tt.expectedMatch {
			t.Errorf("isGatewayMatch(%s, %v) = %v; want %v", tt.gateway, tt.gatewayNames, match, tt.expectedMatch)
		}
	}
}

// Test_isPortMatch tests the isPortMatch function.
func Test_isPortMatch(t *testing.T) {
	server := &networking.Server{
		Port: &networking.Port{
			Number: 80,
		},
	}

	tests := []struct {
		port        uint32
		serverPort  *networking.Server
		expected    bool
	}{
		{80, server, true},
		{0, server, true},
		{443, server, false},
	}

	for _, tt := range tests {
		match := isPortMatch(tt.port, tt.serverPort)
		if match != tt.expected {
			t.Errorf("isPortMatch(%d, %v) = %v; want %v", tt.port, tt.serverPort, match, tt.expected)
		}
	}
}

// Test_pickMatchingGatewayHosts tests pickMatchingGatewayHosts function.
func Test_pickMatchingGatewayHosts(t *testing.T) {
	virtualService := config.Config{
		Spec: &networking.VirtualService{
			Hosts: []string{"*.example.com"},
		},
		Namespace: "default",
	}

	gatewayHosts := sets.NewWithLength[host.Name](1)
	gatewayHosts.Insert("ns/host1.example.com")

	matching := pickMatchingGatewayHosts(gatewayHosts, virtualService)

	if len(matching) != 1 {
		t.Errorf("Expected 1 matching host, got %d", len(matching))
	}
	if matching["*.example.com"] != "ns/host1.example.com" {
		t.Errorf("Expected host mapping to ns/host1.example.com, got %s", matching["*.example.com"])
	}
}

// Test_hashRouteList tests the hashRouteList function.
func Test_hashRouteList(t *testing.T) {
	route1 := &route.Route{}
	route2 := &route.Route{}
	routes := []*route.Route{route1, route2}

	hash1 := hashRouteList(routes)
	hash2 := hashRouteList(routes)

	if hash1 != hash2 {
		t.Errorf("Expected consistent hash, got %d and %d", hash1, hash2)
	}

	// Modify the list and ensure the hash changes
	routes = append(routes, &route.Route{})
	hash3 := hashRouteList(routes)
	if hash1 == hash3 {
		t.Errorf("Expected different hash after modifying the route list")
	}
}

// Test_getListenerName tests the getListenerName function.
func Test_getListenerName(t *testing.T) {
	tests := []struct {
		bind       string
		port       int
		transport  istionetworking.TransportProtocol
		expected   string
	}{
		{"0.0.0.0", 80, istionetworking.TransportProtocolTCP, "0.0.0.0_80"},
		{"127.0.0.1", 443, istionetworking.TransportProtocolQUIC, "udp_127.0.0.1_443"},
		{"", 8080, istionetworking.TransportProtocolTCP, "_8080"},
		{"[::]", 8443, istionetworking.TransportProtocolQUIC, "udp_[::]_8443"},
		{"localhost", 1234, istionetworking.TransportProtocol_TCP, "localhost_1234"},
		{"unknown", 9999, istionetworking.TransportProtocol(-1), "unknown_9999"},
	}

	for _, tt := range tests {
		name := getListenerName(tt.bind, tt.port, tt.transport)
		if name != tt.expected {
			t.Errorf("getListenerName(%s, %d, %s) = %s; want %s",
				tt.bind, tt.port, tt.transport.String(), name, tt.expected)
		}
	}
}

// Test_collapseDuplicateRoutes_NoCollapse tests collapseDuplicateRoutes without enabling route collapse.
func Test_collapseDuplicateRoutes_NoCollapse(t *testing.T) {
	original := map[host.Name]*route.VirtualHost{
		"host1.example.com": {
			Name:    "vhost1",
			Domains: []string{"host1.example.com"},
			Routes:  []*route.Route{{}},
		},
		"host2.example.com": {
			Name:    "vhost2",
			Domains: []string{"host2.example.com"},
			Routes:  []*route.Route{{}},
		},
	}

	output := collapseDuplicateRoutes(original)

	if len(output) != len(original) {
		t.Errorf("Expected no collapse, got %d vs %d", len(output), len(original))
	}
}

// Test_vhostMergeable tests the vhostMergeable function.
func Test_vhostMergeable(t *testing.T) {
	vhost1 := &route.VirtualHost{
		Name:                       "vhost1",
		IncludeRequestAttemptCount: true,
		RequireTls:                 route.VirtualHost_ALL,
		Routes:                     []*route.Route{{}},
	}
	vhost2 := &route.VirtualHost{
		Name:                       "vhost2",
		IncludeRequestAttemptCount: true,
		RequireTls:                 route.VirtualHost_ALL,
		Routes:                     []*route.Route{{}},
	}
	vhost3 := &route.VirtualHost{
		Name:                       "vhost3",
		IncludeRequestAttemptCount: false,
		RequireTls:                 route.VirtualHost_ALL,
		Routes:                     []*route.Route{{}},
	}
	vhost4 := &route.VirtualHost{
		Name:                       "vhost4",
		IncludeRequestAttemptCount: true,
		RequireTls:                 route.VirtualHost_NONE,
		Routes:                     []*route.Route{{}},
	}
	vhost5 := &route.VirtualHost{
		Name:                       "vhost5",
		IncludeRequestAttemptCount: true,
		RequireTls:                 route.VirtualHost_ALL,
		Routes:                     []*route.Route{{}, {}},
	}

	if !vhostMergeable(vhost1, vhost2) {
		t.Errorf("Expected vhost1 and vhost2 to be mergeable")
	}
	if vhostMergeable(vhost1, vhost3) {
		t.Errorf("Expected vhost1 and vhost3 to be not mergeable")
	}
	if vhostMergeable(vhost1, vhost4) {
		t.Errorf("Expected vhost1 and vhost4 to be not mergeable")
	}
	if vhostMergeable(vhost1, vhost5) {
		t.Errorf("Expected vhost1 and vhost5 to be not mergeable")
	}
}

// Test_routesEqual tests the routesEqual function.
func Test_routesEqual(t *testing.T) {
	route1 := &route.Route{}
	route2 := &route.Route{}
	route3 := &route.Route{}

	tests := []struct {
		a        []*route.Route
		b        []*route.Route
		expected bool
	}{
		{[]*route.Route{route1, route2}, []*route.Route{route1, route2}, true},
		{[]*route.Route{route1, route2}, []*route.Route{route2, route1}, false},
		{[]*route.Route{route1}, []*route.Route{route1, route2}, false},
		{[]*route.Route{route1, route3}, []*route.Route{route1, route2}, false},
		{[]*route.Route{}, []*route.Route{}, true},
	}

	for _, tt := range tests {
		equal := routesEqual(tt.a, tt.b)
		if equal != tt.expected {
			t.Errorf("routesEqual(%v, %v) = %v; want %v", tt.a, tt.b, equal, tt.expected)
		}
	}
}

// Test_buildGatewayConnectionManager_HTTP3 tests buildGatewayConnectionManager with HTTP/3 enabled.
func Test_buildGatewayConnectionManager_HTTP3(t *testing.T) {
	proxyConfig := &meshconfig.ProxyConfig{
		GatewayTopology: &meshconfig.GatewayTopology{
			NumTrustedProxies:          1,
			ForwardClientCertDetails:   meshconfig.ForwardClientCertDetails_SANITIZE_SET,
		},
	}
	node := &model.Proxy{}
	push := &model.PushContext{
		MeshNetworks: nil,
		Mesh:         &meshconfig.MeshConfig{},
	}
	http3Enabled := true

	connManager := buildGatewayConnectionManager(proxyConfig, node, http3Enabled, push)

	if connManager.CodecType != hcm.HttpConnectionManager_HTTP3 {
		t.Errorf("Expected CodecType HTTP3, got %v", connManager.CodecType)
	}
	if connManager.Http3ProtocolOptions == nil {
		t.Errorf("Expected Http3ProtocolOptions to be set")
	}
}

// Test_buildGatewayListener_Success tests buildGatewayListener function.
func Test_buildGatewayListener_Success(t *testing.T) {
	opts := gatewayListenerOpts{
		port: 80,
	}
	transport := istionetworking.TransportProtocolTCP

	listener := buildGatewayListener(opts, transport)

	if listener == nil {
		t.Errorf("Expected listener to be built")
	}
	if listener.Name != "0.0.0.0_80" && listener.Name != "_80" {
		t.Errorf("Unexpected listener name: %s", listener.Name)
	}
}

// Additional tests for edge and error cases can be added similarly.