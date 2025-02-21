package controlplane

import (
	"time"

	managercfg "github.com/kong/kubernetes-ingress-controller/v3/pkg/manager/config"
	"github.com/samber/mo"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	"github.com/kong/gateway-operator/pkg/vars"
)

func WithRestConfig(restCfg *rest.Config) managercfg.Opt {
	return func(c *managercfg.Config) {
		c.APIServerHost = restCfg.Host
		c.APIServerCertData = restCfg.CertData
		c.APIServerKeyData = restCfg.KeyData
		c.APIServerCAData = restCfg.CAData
	}
}

func WithKongAdminService(s types.NamespacedName) managercfg.Opt {
	return func(c *managercfg.Config) {
		c.KongAdminSvc = mo.Some(s)
	}
}

func WithKongAdminServicePortName(portName string) managercfg.Opt {
	return func(c *managercfg.Config) {
		c.KongAdminSvcPortNames = []string{portName}
	}
}

func WithKongAdminInitializationRetryDelay(delay time.Duration) managercfg.Opt {
	return func(c *managercfg.Config) {
		c.KongAdminInitializationRetryDelay = delay
	}
}

func WithGatewayToReconcile(gateway types.NamespacedName) managercfg.Opt {
	return func(c *managercfg.Config) {
		c.GatewayToReconcile = mo.Some(gateway)
	}
}

func WithGatewayAPIControllerName() managercfg.Opt {
	return func(c *managercfg.Config) {
		c.GatewayAPIControllerName = vars.ControllerName()
	}
}

func WithKongAdminAPIConfig(cfg managercfg.AdminAPIClientConfig) managercfg.Opt {
	return func(c *managercfg.Config) {
		c.KongAdminAPIConfig = cfg
	}
}

func WithDisabledLeaderElection() managercfg.Opt {
	return func(c *managercfg.Config) {
		c.LeaderElectionForce = "disabled"
	}
}

func WithPublishService(service types.NamespacedName) managercfg.Opt {
	return func(c *managercfg.Config) {
		c.PublishService = mo.Some(service)
	}
}

func WithMetricsServerOff() managercfg.Opt {
	return func(c *managercfg.Config) {
		c.MetricsAddr = "0" // 0 disables metrics server
	}
}
