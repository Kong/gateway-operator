package isolated

import (
	"errors"
	"io"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kong/kong-operator/ingress-controller/test"
	"github.com/kong/kong-operator/ingress-controller/test/integration/consts"
)

func assertEventuallyNoResponseUDP(t *testing.T, udpGatewayURL string) {
	t.Helper()
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		// For UDP lack of response (a timeout) means that we can't reach a service.
		type timeout interface {
			Timeout() bool
		}
		err := test.EchoResponds(test.ProtocolUDP, udpGatewayURL, "irrelevant")
		var timeoutErr timeout
		if assert.ErrorAs(c, err, &timeoutErr, "unexpected error: %v", err) {
			assert.True(c, timeoutErr.Timeout(), `returned syscall error should be "i/o timeout", but it's: %v`, err)
		}
	}, consts.IngressWait, consts.WaitTick)
}

func assertEventuallyResponseUDP(t *testing.T, udpGatewayURL, expectedMsg string) {
	t.Helper()
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, test.EchoResponds(test.ProtocolUDP, udpGatewayURL, expectedMsg))
	}, consts.IngressWait, consts.WaitTick)
}

func assertEventuallyNoResponseTCP(t *testing.T, tcpGatewayURL string) {
	t.Helper()
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := test.EchoResponds(test.ProtocolTCP, tcpGatewayURL, "irrelevant")
		assert.True(c, errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET), "unexpected error: %v", err)
	}, consts.IngressWait, consts.WaitTick)
}

func assertEventuallyResponseTCP(t *testing.T, tcpGatewayURL, expectedMsg string) {
	t.Helper()
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, test.EchoResponds(test.ProtocolTCP, tcpGatewayURL, expectedMsg))
	}, consts.IngressWait, consts.WaitTick)
}
