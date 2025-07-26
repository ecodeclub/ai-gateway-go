package ioc

import (
	"github.com/gotomicro/ego/core/econf"
)

func initSecretKey() string {
	type SecretKeyConfig struct {
		SecretKey string `json:"secret_key"`
	}
	var cfg SecretKeyConfig

	err := econf.UnmarshalKey("secret", &cfg)
	if err != nil {
		panic(err)
	}
	return cfg.SecretKey
}
