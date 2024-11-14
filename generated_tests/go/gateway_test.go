package core

import (
	"errors"
	"testing"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/util"

	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name          string
		listenerOpts  gatewayListenerOpts
		wantErr       bool
		expectedError error
	}{
		{
			name: "Success",
			listenerOpts: gatewayListenerOpts{
				filterChainOpts: []*filterChainOpts{
					{},
				},
			},
			wantErr: false,
		},
		{
			name: "No Filter Chains Error",
			listenerOpts: gatewayListenerOpts{
				filterChainOpts: []*filterChainOpts{},
			},
			wantErr:       true,
			expectedError: errors.New("must have more than 0 chains in listener \"\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &MutableGatewayListener{
				Listener: &listener.Listener{Name: "test-listener"},
			}
			builder := &ListenerBuilder{}
			err := ml.build(builder, tt.listenerOpts)
			if tt.wantErr {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildGatewayListeners(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := NewMockListenerBuilderInterface(ctrl)
	mockBuilder.EXPECT().node().Return(&model.Proxy{}).AnyTimes()
	mockBuilder.EXPECT().push().Return(&model.PushContext{}).AnyTimes()

	configGen := &ConfigGeneratorImpl{}

	t.Run("No gateways", func(t *testing.T) {
		mockBuilder.EXPECT().node().Return(&model.Proxy{MergedGateway: nil})
		builder := configGen.buildGatewayListeners(mockBuilder)
		assert.NotNil(t, builder)
	})

	t.Run("Has gateways", func(t *testing.T) {
		gw := &model.MergedGateway{
			Servers: map[string][]*model.Server{},
		}
		mockBuilder.EXPECT().node().Return(&model.Proxy{MergedGateway: gw})
		mockBuilder.EXPECT().push().Return(&model.PushContext{})
		mockBuilder.EXPECT().gatewayListeners().Return([]*listener.Listener{}).AnyTimes()

		builder := configGen.buildGatewayListeners(mockBuilder)
		assert.NotNil(t, builder)
	})
}

func TestBuildGatewayHTTPRouteConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := NewMockListenerBuilderInterface(ctrl)
	mockBuilder.EXPECT().node().Return(&model.Proxy{}).AnyTimes()
	mockBuilder.EXPECT().push().Return(&model.PushContext{}).AnyTimes()

	configGen := &ConfigGeneratorImpl{}

	t.Run("No merged gateways", func(t *testing.T) {
		mockBuilder.EXPECT().node().Return(&model.Proxy{MergedGateway: nil})
		routeConfig := configGen.buildGatewayHTTPRouteConfig(mockBuilder.node(), mockBuilder.push(), "test-route")
		assert.NotNil(t, routeConfig)
		assert.Empty(t, routeConfig.VirtualHosts)
	})

	t.Run("Has merged gateways", func(t *testing.T) {
		gw := &model.MergedGateway{
			Servers: map[string][]*model.Server{},
		}
		mockBuilder.EXPECT().node().Return(&model.Proxy{MergedGateway: gw})
		mockBuilder.EXPECT().push().Return(&model.PushContext{})
		routeConfig := configGen.buildGatewayHTTPRouteConfig(mockBuilder.node(), mockBuilder.push(), "test-route")
		assert.NotNil(t, routeConfig)
	})
}

func TestBuildNameToServiceMapForHTTPRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPushContext := NewMockPushContext(ctrl)
	mockProxy := NewMockProxy(ctrl)

	vs := config.Config{
		Meta: config.Meta{
			Name:      "vs-1",
			Namespace: "default",
		},
		Spec: &networking.VirtualService{
			Hosts: []string{"example.com"},
		},
	}

	nameToServiceMap := buildNameToServiceMapForHTTPRoutes(mockProxy, mockPushContext, vs)
	assert.NotEmpty(t, nameToServiceMap)
}