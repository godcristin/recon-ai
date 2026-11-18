package llm

import (
	"os"

	"recon-ai/app/ai/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

// NewClientFromConf 从配置构造客户端。
// API Key 只从环境变量读取，绝不写入配置文件或代码。
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
