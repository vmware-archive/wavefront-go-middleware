package echo

import (
	"github.com/wavefronthq/wavefront-sdk-go/senders"
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

//GetWavefrontDirectSender constructs and returns a direct wavefront sender
func getWavefrontDirectSender(server string, token string, flushIntervalSeconds int) (wavefront.Sender, error) {
	var wavefrontSender senders.Sender

	directCfg := &wavefront.DirectConfiguration{
		Server:               server,
		Token:                token,
		BatchSize:            40000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: flushIntervalSeconds,
	}

	wavefrontSender, err := wavefront.NewDirectSender(directCfg)
	return wavefrontSender, err
}

//GetWavefrontProxySender constructs and returns a proxy wavefront sender
func getWavefrontProxySender(proxyHost string, flushIntervalSeconds int) (wavefront.Sender, error) {
	var wavefrontSender senders.Sender

	//Creating proxy sender
	proxyCfg := &wavefront.ProxyConfiguration{
		Host:                 proxyHost,
		MetricsPort:          2878,
		DistributionPort:     2878,
		TracingPort:          30000,
		FlushIntervalSeconds: flushIntervalSeconds,
	}

	wavefrontSender, err := wavefront.NewProxySender(proxyCfg)
	return wavefrontSender, err
}
