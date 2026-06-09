package config

import (
	"os"
	"testing"
)

func TestLoadConfigUsesNonSensitiveLocalDefaults(t *testing.T) {
	t.Setenv("APP_DOTENV_PATH", t.TempDir()+"/missing.env")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("WECHAT_APP_ID", "")
	t.Setenv("WECHAT_APP_SECRET", "")

	cfg := LoadConfig()

	if cfg.DBConfig.Host != "127.0.0.1" {
		t.Fatalf("expected local DB host default, got %q", cfg.DBConfig.Host)
	}
	if cfg.DBConfig.Password != "" {
		t.Fatalf("expected empty DB password default")
	}
	if cfg.DBConfig.RdsPassword != "" {
		t.Fatalf("expected empty Redis password default")
	}
	if cfg.JWTConfig.SecretKey != "dev-only-change-me" {
		t.Fatalf("expected non-sensitive JWT default, got %q", cfg.JWTConfig.SecretKey)
	}
	if cfg.WechatConfig.AppID != "" || cfg.WechatConfig.AppSecret != "" {
		t.Fatalf("expected empty WeChat credentials by default")
	}
}

func TestLoadConfigReadsEnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_DOTENV_PATH", t.TempDir()+"/missing.env")
	t.Setenv("APP_PORT", "3099")
	t.Setenv("LOG_DIR", "./tmp-logs")
	t.Setenv("CORS_ALLOW_ORIGINS", "https://admin.example.com,https://ops.example.com")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PASSWORD", "from-env")
	t.Setenv("REDIS_ADDR", "redis.internal:6379")
	t.Setenv("REDIS_DB", "3")
	t.Setenv("JWT_SECRET", "jwt-from-env")
	t.Setenv("WECHAT_APP_ID", "wx-from-env")
	t.Setenv("WECHAT_LOGIN_URL", "https://wechat.example.com/login")
	t.Setenv("WECHAT_ACCESS_TOKEN_URL", "https://wechat.example.com/stable-token")
	t.Setenv("WECHAT_PHONE_NUMBER_URL", "https://wechat.example.com/phone-number")
	t.Setenv("ALIYUN_SMS_TEMPLATE_CODE", "sms-template")
	t.Setenv("JST_ENV_STAGE", "test")
	t.Setenv("JST_APP_KEY_PROD", "jst-key")
	t.Setenv("JST_SHOP_ID", "10001")
	t.Setenv("JST_ORDER_UPLOAD_URL_TEST", "https://jst.example.com/order-upload")
	t.Setenv("JST_INVENTORY_QUERY_URL_PROD", "https://jst.example.com/inventory-query")
	t.Setenv("JST_SKUMAP_QUERY_URL_TEST", "https://jst.example.com/skumap-query")

	cfg := LoadConfig()

	if cfg.ServerConfig.Port != "3099" {
		t.Fatalf("expected APP_PORT override, got %q", cfg.ServerConfig.Port)
	}
	if len(cfg.ServerConfig.CORSAllowOrigins) != 2 {
		t.Fatalf("expected two CORS origins, got %#v", cfg.ServerConfig.CORSAllowOrigins)
	}
	if cfg.DBConfig.Host != "db.internal" || cfg.DBConfig.Password != "from-env" {
		t.Fatalf("expected DB env overrides, got host=%q password=%q", cfg.DBConfig.Host, cfg.DBConfig.Password)
	}
	if cfg.DBConfig.RdsHost != "redis.internal:6379" || cfg.DBConfig.RdsNum != 3 {
		t.Fatalf("expected Redis env overrides, got addr=%q db=%d", cfg.DBConfig.RdsHost, cfg.DBConfig.RdsNum)
	}
	if cfg.JWTConfig.SecretKey != "jwt-from-env" {
		t.Fatalf("expected JWT env override")
	}
	if cfg.WechatConfig.AppID != "wx-from-env" {
		t.Fatalf("expected WeChat env override")
	}
	if cfg.WechatConfig.LoginURL != "https://wechat.example.com/login" {
		t.Fatalf("expected WeChat login URL env override")
	}
	if cfg.WechatConfig.AccessTokenURL != "https://wechat.example.com/stable-token" {
		t.Fatalf("expected WeChat access token URL env override")
	}
	if cfg.WechatConfig.PhoneNumberURL != "https://wechat.example.com/phone-number" {
		t.Fatalf("expected WeChat phone number URL env override")
	}
	if cfg.SMSConfig.TemplateCode != "sms-template" {
		t.Fatalf("expected SMS env override")
	}
	if cfg.JushuitanConfig.AppKeyProd != "jst-key" {
		t.Fatalf("expected Jushuitan env override")
	}
	if cfg.JushuitanConfig.Stage != "test" || cfg.JushuitanConfig.ShopID != "10001" {
		t.Fatalf("expected Jushuitan stage/shop overrides, got stage=%q shop_id=%q", cfg.JushuitanConfig.Stage, cfg.JushuitanConfig.ShopID)
	}
	if cfg.JushuitanConfig.OrderUploadURLTest != "https://jst.example.com/order-upload" {
		t.Fatalf("expected Jushuitan order upload URL env override")
	}
	if cfg.JushuitanConfig.InventoryQueryURLProd != "https://jst.example.com/inventory-query" {
		t.Fatalf("expected Jushuitan inventory query URL env override")
	}
	if cfg.JushuitanConfig.SkuMapQueryURLTest != "https://jst.example.com/skumap-query" {
		t.Fatalf("expected Jushuitan skumap test URL env override")
	}
}

func TestLoadDotEnvFillsMissingValuesOnly(t *testing.T) {
	t.Setenv("DOTENV_FILLED", "")
	t.Setenv("DOTENV_EXISTING", "from-env")

	envFile := t.TempDir() + "/.env"
	if err := os.WriteFile(envFile, []byte("DOTENV_FILLED=from-file\nDOTENV_EXISTING=from-file\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	loadDotEnv(envFile)

	if got := os.Getenv("DOTENV_FILLED"); got != "from-file" {
		t.Fatalf("expected .env value, got %q", got)
	}
	if got := os.Getenv("DOTENV_EXISTING"); got != "from-env" {
		t.Fatalf("expected existing env to win, got %q", got)
	}
}
