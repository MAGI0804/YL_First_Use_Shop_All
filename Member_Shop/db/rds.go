package db

import (
	"Member_shop/config"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisCtx = context.Background()

var Rds *redis.Client

// InitRedis 初始化令牌和验证码共享的Redis客户端
func InitRedis(appConfig config.Config) error {
	Rds = redis.NewClient(&redis.Options{
		Addr:        appConfig.DBConfig.RdsHost,
		Password:    appConfig.DBConfig.RdsPassword,
		DB:          appConfig.DBConfig.RdsNum,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	})

	_, err := Rds.Ping(RedisCtx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	fmt.Println("Redis已初始化")
	return nil
}

// GenerateCaptcha 使用crypto/rand创建6位数字验证码
func GenerateCaptcha() string {
	var builder strings.Builder
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			builder.WriteByte('0')
			continue
		}
		builder.WriteString(n.String())
	}
	return builder.String()
}

// SaveCaptcha 将手机验证码存储在Redis中，有效期3分钟
func SaveCaptcha(phone, captcha string) error {
	if Rds == nil {
		return fmt.Errorf("Redis未初始化")
	}
	key := captchaKey(phone)
	if err := Rds.Set(RedisCtx, key, captcha, 3*time.Minute).Err(); err != nil {
		return fmt.Errorf("保存验证码到Redis失败: %v", err)
	}
	return nil
}

// VerifyCaptcha 检查Redis中的验证码，验证成功后删除
func VerifyCaptcha(phone, captcha string) error {
	if Rds == nil {
		return fmt.Errorf("Redis未初始化")
	}
	key := captchaKey(phone)
	storedCaptcha, err := Rds.Get(RedisCtx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("验证码已过期或未找到")
	}
	if err != nil {
		return fmt.Errorf("从Redis读取验证码失败: %v", err)
	}
	if storedCaptcha != captcha {
		return fmt.Errorf("验证码不正确")
	}
	_ = Rds.Del(RedisCtx, key).Err()
	return nil
}

// captchaKey 将验证码值与其他Redis键隔离
func captchaKey(phone string) string {
	return fmt.Sprintf("captcha:phone:%s", phone)
}

// SaveTokenRedis 将令牌保存到Redis
func SaveTokenRedis(token, ip string) error {
	if Rds == nil {
		return fmt.Errorf("Redis未初始化")
	}
	if token == "" {
		return fmt.Errorf("令牌为空")
	}
	err := Rds.Set(
		RedisCtx,
		ip,
		token,
		720*time.Hour,
	).Err()
	if err != nil {
		return fmt.Errorf("保存令牌到Redis失败: %v", err)
	}
	return nil
}

// GetTokenRedis 从Redis获取令牌
func GetTokenRedis(ip string) (string, error) {
	if Rds == nil {
		return "", fmt.Errorf("Redis未初始化")
	}
	if ip == "" {
		return "", fmt.Errorf("IP为空")
	}
	token, err := Rds.HGet(RedisCtx, "access_token", ip).Result()
	if err != nil {
		return "", err
	}
	return token, nil
}

// SaveToken 保存访问令牌到Redis的哈希结构
func SaveToken(token, ip string) error {
	if Rds == nil {
		return fmt.Errorf("Redis未初始化")
	}
	if token == "" || ip == "" {
		return fmt.Errorf("令牌或IP为空")
	}
	_, err := Rds.HSet(RedisCtx, "access_token", ip, token).Result()
	if err != nil {
		return fmt.Errorf("保存访问令牌失败: %v", err)
	}
	if err := Rds.Expire(RedisCtx, "access_token", 720*time.Hour).Err(); err != nil {
		return fmt.Errorf("设置访问令牌过期时间失败: %v", err)
	}
	return nil
}
