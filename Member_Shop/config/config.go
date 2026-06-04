package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerConfig
	DBConfig
	JWTConfig
	WechatConfig
	SMSConfig
	JushuitanConfig
}

type ServerConfig struct {
	Environment      string
	Port             string
	LogDir           string
	CORSAllowOrigins []string
}

type DBConfig struct {
	Driver      string
	Host        string
	Port        string
	Username    string
	Password    string
	DBName      string
	Charset     string
	RdsPassword string
	RdsHost     string
	RdsNum      int
}

type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type WechatConfig struct {
	AppID     string
	AppSecret string
	LoginURL  string
}

type SMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	Endpoint        string
}

type JushuitanConfig struct {
	Stage                         string
	AppKeyTest                    string
	AppSecretTest                 string
	AppKeyProd                    string
	AppSecretProd                 string
	AuthCodeProd                  string
	AccessTokenTest               string
	ShopID                        string
	GetTokenURLTest               string
	GetTokenURLProd               string
	OpenAPIURLTest                string
	OpenAPIURLProd                string
	ShopQueryURLTest              string
	ShopQueryURLProd              string
	OrderUploadURLTest            string
	OrderUploadURLProd            string
	AfterSaleUploadURLTest        string
	AfterSaleUploadURLProd        string
	AfterSaleReceivedQueryURLTest string
	AfterSaleReceivedQueryURLProd string
	LogisticQueryURLTest          string
	LogisticQueryURLProd          string
	InventoryQueryURLTest         string
	InventoryQueryURLProd         string
	SkuMapQueryURLTest            string
	SkuMapQueryURLProd            string
	SkuQueryURLTest               string
	SkuQueryURLProd               string
}

func LoadConfig() Config {
	dotEnvPath := os.Getenv("APP_DOTENV_PATH")
	if dotEnvPath == "" {
		dotEnvPath = ".env"
	}
	loadDotEnv(dotEnvPath)

	return Config{
		ServerConfig: ServerConfig{
			Environment:      getEnv("APP_ENV", "development"),
			Port:             getEnv("APP_PORT", "3088"),
			LogDir:           getEnv("LOG_DIR", "./logs"),
			CORSAllowOrigins: splitCSV(getEnv("CORS_ALLOW_ORIGINS", "*")),
		},
		DBConfig: DBConfig{
			Driver:      getEnv("DB_ENGINE", "mysql"),
			Host:        getEnv("DB_HOST", "127.0.0.1"),
			Port:        getEnv("DB_PORT", "3306"),
			Username:    getEnv("DB_USER", "root"),
			Password:    getEnv("DB_PASSWORD", ""),
			DBName:      getEnv("DB_NAME", "wechat_member"),
			Charset:     getEnv("DB_CHARSET", "utf8mb4"),
			RdsPassword: getEnv("REDIS_PASSWORD", ""),
			RdsHost:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			RdsNum:      getEnvInt("REDIS_DB", 1),
		},
		JWTConfig: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET", "dev-only-change-me"),
			AccessTokenTTL:  time.Duration(getEnvInt("JWT_ACCESS_TOKEN_TTL_HOURS", 1)),
			RefreshTokenTTL: time.Duration(getEnvInt("JWT_REFRESH_TOKEN_TTL_HOURS", 24)),
		},
		WechatConfig: WechatConfig{
			AppID:     getEnv("WECHAT_APP_ID", ""),
			AppSecret: getEnv("WECHAT_APP_SECRET", ""),
			LoginURL:  getEnv("WECHAT_LOGIN_URL", ""),
		},
		SMSConfig: SMSConfig{
			AccessKeyID:     getEnv("ALIYUN_SMS_ACCESS_KEY_ID", ""),
			AccessKeySecret: getEnv("ALIYUN_SMS_ACCESS_KEY_SECRET", ""),
			SignName:        getEnv("ALIYUN_SMS_SIGN_NAME", ""),
			TemplateCode:    getEnv("ALIYUN_SMS_TEMPLATE_CODE", ""),
			Endpoint:        getEnv("ALIYUN_SMS_ENDPOINT", "dysmsapi.aliyuncs.com"),
		},
		JushuitanConfig: JushuitanConfig{
			Stage:                         getEnv("JST_ENV_STAGE", "develop"),
			AppKeyTest:                    getEnv("JST_APP_KEY_TEST", ""),
			AppSecretTest:                 getEnv("JST_APP_SECRET_TEST", ""),
			AppKeyProd:                    getEnv("JST_APP_KEY_PROD", ""),
			AppSecretProd:                 getEnv("JST_APP_SECRET_PROD", ""),
			AuthCodeProd:                  getEnv("JST_AUTH_CODE_PROD", ""),
			AccessTokenTest:               getEnv("JST_ACCESS_TOKEN_TEST", ""),
			ShopID:                        getEnv("JST_SHOP_ID", ""),
			GetTokenURLTest:               getEnv("JST_GET_TOKEN_URL_TEST", ""),
			GetTokenURLProd:               getEnv("JST_GET_TOKEN_URL_PROD", ""),
			OpenAPIURLTest:                getEnv("JST_OPEN_API_URL_TEST", ""),
			OpenAPIURLProd:                getEnv("JST_OPEN_API_URL_PROD", ""),
			ShopQueryURLTest:              getEnv("JST_SHOP_QUERY_URL_TEST", ""),
			ShopQueryURLProd:              getEnv("JST_SHOP_QUERY_URL_PROD", ""),
			OrderUploadURLTest:            getEnv("JST_ORDER_UPLOAD_URL_TEST", ""),
			OrderUploadURLProd:            getEnv("JST_ORDER_UPLOAD_URL_PROD", ""),
			AfterSaleUploadURLTest:        getEnv("JST_AFTERSALE_UPLOAD_URL_TEST", ""),
			AfterSaleUploadURLProd:        getEnv("JST_AFTERSALE_UPLOAD_URL_PROD", ""),
			AfterSaleReceivedQueryURLTest: getEnv("JST_AFTERSALE_RECEIVED_QUERY_URL_TEST", ""),
			AfterSaleReceivedQueryURLProd: getEnv("JST_AFTERSALE_RECEIVED_QUERY_URL_PROD", ""),
			LogisticQueryURLTest:          getEnv("JST_LOGISTIC_QUERY_URL_TEST", ""),
			LogisticQueryURLProd:          getEnv("JST_LOGISTIC_QUERY_URL_PROD", ""),
			InventoryQueryURLTest:         getEnv("JST_INVENTORY_QUERY_URL_TEST", ""),
			InventoryQueryURLProd:         getEnv("JST_INVENTORY_QUERY_URL_PROD", ""),
			SkuMapQueryURLTest:            getEnv("JST_SKUMAP_QUERY_URL_TEST", ""),
			SkuMapQueryURLProd:            getEnv("JST_SKUMAP_QUERY_URL_PROD", ""),
			SkuQueryURLTest:               getEnv("JST_SKU_QUERY_URL_TEST", ""),
			SkuQueryURLProd:               getEnv("JST_SKU_QUERY_URL_PROD", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return []string{"*"}
	}
	return result
}

func loadDotEnv(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
}
