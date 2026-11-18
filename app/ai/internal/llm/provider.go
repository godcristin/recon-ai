package llm

import (
	"os"

	"recon-ai/app/ai/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

func NewClientFromConf(c *conf.Bootstrap, logger log.Logger) *Client {
	envName := c.Llm.ApiKeyEnv
	if envName == "" {
		envName = "DEEPSEEK_API_KEY"
	}

	client := NewClient(c.Llm.BaseUrl, os.Getenv(envName), c.Llm.Model, c.Llm.MaxTokens, c.Llm.Timeout.AsDuration())
	if !client.Enabled() {
		log.NewHelper(logger).Warnw("msg", "llm client disabled",
			"reason", "missing api key or base url", "env", envName)
	}
	return client
}
