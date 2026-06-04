package jushuitan

import (
	"Member_shop/config"
	"fmt"
	"strings"
)

const (
	jushuitanStageTest = "test"
	jushuitanStageProd = "prod"
)

type openAPIEnvironment struct {
	AppKey    string
	AppSecret string
	Stage     string
}

func normalizeJushuitanStage(stage string) string {
	switch strings.ToLower(strings.TrimSpace(stage)) {
	case "", "test", "develop", "development", "dev":
		return jushuitanStageTest
	case "prod", "production", "formal", "release", "main", "master":
		return jushuitanStageProd
	default:
		return jushuitanStageTest
	}
}

func useJushuitanTestEnvironment(cfg config.Config) bool {
	return normalizeJushuitanStage(cfg.JushuitanConfig.Stage) == jushuitanStageTest
}

func activeOpenAPIEnvironment(cfg config.Config) (openAPIEnvironment, error) {
	if useJushuitanTestEnvironment(cfg) {
		if cfg.JushuitanConfig.AppKeyTest == "" || cfg.JushuitanConfig.AppSecretTest == "" {
			return openAPIEnvironment{}, fmt.Errorf("聚水潭测试应用配置未完整设置")
		}
		return openAPIEnvironment{
			AppKey:    cfg.JushuitanConfig.AppKeyTest,
			AppSecret: cfg.JushuitanConfig.AppSecretTest,
			Stage:     jushuitanStageTest,
		}, nil
	}

	if cfg.JushuitanConfig.AppKeyProd == "" || cfg.JushuitanConfig.AppSecretProd == "" {
		return openAPIEnvironment{}, fmt.Errorf("聚水潭正式应用配置未完整设置")
	}
	return openAPIEnvironment{
		AppKey:    cfg.JushuitanConfig.AppKeyProd,
		AppSecret: cfg.JushuitanConfig.AppSecretProd,
		Stage:     jushuitanStageProd,
	}, nil
}

func activeURL(cfg config.Config, testURL, prodURL, testEnvName, prodEnvName string) (string, error) {
	apiURL := strings.TrimSpace(testURL)
	envName := testEnvName
	if !useJushuitanTestEnvironment(cfg) {
		apiURL = strings.TrimSpace(prodURL)
		envName = prodEnvName
	}
	if apiURL == "" {
		return "", fmt.Errorf("%s未配置", envName)
	}
	return apiURL, nil
}
