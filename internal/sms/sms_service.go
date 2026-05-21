package sms

import (
	"encoding/json"
	"event-platform/internal/config"
	"event-platform/internal/utils"
	"fmt"
	"math/rand"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi20170525 "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/sirupsen/logrus"
)

type SMSService struct {
	client *dypnsapi20170525.Client
	cfg    *config.Config
}

func NewSMSService(cfg *config.Config) *SMSService {
	client, err := createClient(cfg)
	if err != nil {
		logrus.Errorf("创建阿里云短信客户端失败: %v", err)
		return &SMSService{cfg: cfg}
	}
	return &SMSService{client: client, cfg: cfg}
}

func createClient(cfg *config.Config) (*dypnsapi20170525.Client, error) {
	openapiConfig := &openapi.Config{
		AccessKeyId:     tea.String(cfg.SMS.AccessKeyID),
		AccessKeySecret: tea.String(cfg.SMS.AccessKeySecret),
	}
	openapiConfig.Endpoint = tea.String("dypnsapi.aliyuncs.com")
	client, err := dypnsapi20170525.NewClient(openapiConfig)
	if err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("创建dypnsapi客户端失败: %w", err))
	}
	return client, nil
}

func (s *SMSService) GenerateVerifyCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%04d", r.Intn(10000))
}

func (s *SMSService) SendVerifyCode(phoneNumber, code string) error {
	// 测试环境不发送短信验证码，直接返回成功
	if s.cfg.App.Env != "production" {
		return nil
	}
	if s.client == nil {
		return utils.NewBusinessError(utils.ErrCodeDependencyServiceError, "短信服务客户端未初始化")
	}

	request := &dypnsapi20170525.SendSmsVerifyCodeRequest{
		SignName:      tea.String(s.cfg.SMS.SignName),
		TemplateCode:  tea.String(s.cfg.SMS.TemplateCode),
		TemplateParam: tea.String(fmt.Sprintf(`{"code":"%s","min":"5"}`, code)),
		PhoneNumber:   tea.String(phoneNumber),
	}

	runtime := &util.RuntimeOptions{}
	resp, err := s.client.SendSmsVerifyCodeWithOptions(request, runtime)
	if err != nil {
		var sdkErr *tea.SDKError
		if strings.Contains(err.Error(), "SDKError") {
			_ = sdkErr
			var data interface{}
			d := json.NewDecoder(strings.NewReader(tea.StringValue(err.(*tea.SDKError).Data)))
			_ = d.Decode(&data)
		}
		logrus.Errorf("阿里云短信发送失败: Phone=%s, Error=%v", phoneNumber, err)
		return utils.NewBusinessError(utils.ErrCodeDependencyServiceError, "短信发送失败，请稍后重试")
	}

	body := resp.Body
	if body == nil || tea.StringValue(body.Code) != "OK" {
		errCode := ""
		errMsg := ""
		if body != nil {
			errCode = tea.StringValue(body.Code)
			errMsg = tea.StringValue(body.Message)
		}
		logrus.Errorf("阿里云短信发送失败: Code=%s, Message=%s, Phone=%s", errCode, errMsg, phoneNumber)
		return utils.NewBusinessError(utils.ErrCodeDependencyServiceError, "短信发送失败，请稍后重试")
	}

	logrus.Infof("短信验证码发送成功: Phone=%s", phoneNumber)
	return nil
}
