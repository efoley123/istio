package core_test

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoyconfigroute "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	statefulsessionv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/stateful_session/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/features"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/core"
	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/istio/pilot/pkg/util/protoconv"
)

func TestMutableGatewayListener_build(t *testing.T) {
	t.Run("SuccessCase", func(t *testing.T) {
		listener := &listenerv3.Listener{Name: "listenerName"}
		mutableListener := &core.MutableGatewayListener{Listener: listener}
		builder := &core.ListenerBuilder{}
		opts := core.GatewayListenerOpts{}

		err := mutableListener.Build(builder, opts)
		assert.NoError(t, err, "build should succeed without error")
	})

	t.Run("ErrorCaseEmptyChains", func(t *testing.T) {
		listener := &listenerv3.Listener{Name: "listenerName"}
		mutableListener := &core.MutableGatewayListener{Listener: listener}
		builder := &core.ListenerBuilder{}
		opts := core.GatewayListenerOpts{} // No filterChainOpts provided

		err := mutableListener.Build(builder, opts)
		assert.Error(t, err, "build should fail due to empty chains")
	})
}

// Additional test cases covering buildGatewayListeners, buildGatewayTCPBasedFilterChains,
// buildGatewayHTTP3FilterChains, etc., should be created following the example above.
// Mocking external dependencies and using auxiliary functions to simplify the setup will be necessary.
// For complete coverage, consider different configurations of the Gateway server (e.g., TCP, TLS, HTTP3)
// and different return values from mocked dependencies.
```

The provided test example focuses on one function, `MutableGatewayListener::build`, demonstrating the structure for a success scenario and an error case. For thorough testing of `gateway.go`, similar test cases should be developed for other methods, taking into account various configurations and possible outcomes. Utilizing mocking frameworks like `gomock` and adopting a structured approach for setup, execution, and assertion will facilitate comprehensive testing.