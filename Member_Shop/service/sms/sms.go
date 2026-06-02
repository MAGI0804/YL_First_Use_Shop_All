package message

import (
	"fmt"
	"log"

	"Member_shop/config"
	"Member_shop/db"
	"Member_shop/models"
	"encoding/json"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"gorm.io/gorm"
)

// CreateClient 创建短信客户端，优先从环境变量获取凭证，其次从数据库获取
func CreateClient() (_result *dysmsapi20170525.Client, _err error) {
	appConfig := config.LoadConfig()
	if appConfig.SMSConfig.AccessKeyID != "" && appConfig.SMSConfig.AccessKeySecret != "" {
		config := &openapi.Config{
			AccessKeyId:     tea.String(appConfig.SMSConfig.AccessKeyID),
			AccessKeySecret: tea.String(appConfig.SMSConfig.AccessKeySecret),
		}
		config.Endpoint = tea.String(appConfig.SMSConfig.Endpoint)
		return dysmsapi20170525.NewClient(config)
	}

	// 确保数据库连接已初始化
	if db.DB == nil {
		log.Println("数据库连接未初始化，正在初始化...")
		db.InitDB(appConfig)
		log.Println("数据库连接初始化完成")
	}

	// 从数据库获取 TokenData：ID=2 且 PlatformName=阿里云
	var tokenData models.TokenData
	result := db.DB.Where("id = ? AND platform_name = ?", 2, "阿里云").First(&tokenData)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("未找到符合条件的 TokenData 记录（ID=2，PlatformName=阿里云）")
			return nil, fmt.Errorf("未找到阿里云短信配置记录")
		}
		log.Printf("查询 TokenData 数据库失败: %v", result.Error)
		return nil, fmt.Errorf("查询数据库失败: %v", result.Error)
	}

	// 关键修复：先断言 interface{} 为 string
	verificationStr, ok := tokenData.VerificationInfo.(string)
	if !ok {
		log.Printf("VerificationInfo 类型不是 string，无法解析，实际类型: %T", tokenData.VerificationInfo)
		return nil, fmt.Errorf("配置数据类型错误，期望 string 类型的 JSON")
	}

	// 解析 VerificationInfo JSON 字符串
	type AliyunConfig struct {
		AccessKeyId     string `json:"AccessKeyId"`
		AccessKeySecret string `json:"AccessKeySecret"`
	}
	var aliyunCfg AliyunConfig
	err := json.Unmarshal([]byte(verificationStr), &aliyunCfg)
	if err != nil {
		log.Printf("解析 VerificationInfo JSON 失败: %v", err)
		return nil, fmt.Errorf("解析阿里云配置失败: %v", err)
	}

	// 校验解析后的数据是否为空
	if aliyunCfg.AccessKeyId == "" || aliyunCfg.AccessKeySecret == "" {
		log.Println("阿里云 AccessKey 或 AccessKeySecret 为空")
		return nil, fmt.Errorf("阿里云密钥配置不完整")
	}

	// 使用解析后的值创建阿里云客户端配置
	config := &openapi.Config{
		AccessKeyId:     tea.String(aliyunCfg.AccessKeyId),
		AccessKeySecret: tea.String(aliyunCfg.AccessKeySecret),
	}
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")

	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

// SendSms 发送短信验证码的方法，只需传递手机号和验证码
func SendSms(phoneNumber string, code string) (*string, error) {
	appConfig := config.LoadConfig()
	if appConfig.SMSConfig.SignName == "" || appConfig.SMSConfig.TemplateCode == "" {
		return nil, fmt.Errorf("短信签名或模板未配置")
	}

	client, err := CreateClient()
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %v", err)
	}

	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneNumber),
		SignName:      tea.String(appConfig.SMSConfig.SignName),
		TemplateCode:  tea.String(appConfig.SMSConfig.TemplateCode),
		TemplateParam: tea.String(fmt.Sprintf("{\"code\":\"%s\"}", code)),
	}
	runtime := &util.RuntimeOptions{}

	resp, err := client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		// 处理错误
		var error = &tea.SDKError{}
		if _t, ok := err.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(err.Error())
		}
		return nil, fmt.Errorf("发送短信失败: %s", tea.StringValue(error.Message))
	}

	return util.ToJSONString(resp), nil
}
