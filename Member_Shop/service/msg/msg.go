package msg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translation "github.com/go-playground/validator/v10/translations/en"
	zh_translation "github.com/go-playground/validator/v10/translations/zh"
)

type Response struct {
	Code int             `json:"code"`
	Msg  any             `json:"msg"`
	Data *map[string]any `json:"data"`
}

type ErrResponseST struct {
	Code int             `json:"code"`
	Msg  any             `json:"msg"`
	Data *map[string]any `json:"data"`
	Err  any             `json:err`
}

var trans ut.Translator

func initTranslator(language string) error {
	//转换成go-playground的validator
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		//创建翻译器
		zhT := zh.New()
		enT := en.New()

		//创建通用翻译器
		//第一个参数是备用语言，后面的是应当支持的语言
		uni := ut.New(enT, enT, zhT)

		//从通过中获取指定语言翻译器
		trans, ok = uni.GetTranslator(language)
		if !ok {
			return fmt.Errorf("not found translator %s", language)
		}

		//绑定到gin的验证器上，对binding的tag进行翻译
		switch language {
		case "zh":
			err := zh_translation.RegisterDefaultTranslations(validate, trans)
			if err != nil {
				return err
			}
		default:
			err := en_translation.RegisterDefaultTranslations(validate, trans)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func remove(errors map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range errors {
		result[key[strings.Index(key, ".")+1:]] = value
	}
	return result
}

func SuccessResponse(msg string, dataPtr *map[string]any) *Response {
	if dataPtr == nil {
		emptyMap := make(map[string]any)
		dataPtr = &emptyMap
	}
	normalizedData := normalizeMap(*dataPtr)
	return &Response{
		Code: 200,
		Msg:  msg,
		Data: &normalizedData,
	}
}

func SuccessResponseStr(msg string) *Response {

	return &Response{
		Code: 200,
		Msg:  msg,
		Data: &map[string]any{},
	}
}

func ErrResponse(msg string, errors error) *ErrResponseST {
	err := initTranslator("zh")
	if err != nil {
		panic(err)
	}
	B := errors.Error()
	if errors, ok := err.(validator.ValidationErrors); ok {
		B := remove(errors.Translate(trans))
		return &ErrResponseST{
			Code: 201,
			Msg:  msg,
			Data: &map[string]any{},
			Err:  B,
		}
	}
	return &ErrResponseST{
		Code: 201,
		Msg:  msg,
		Data: &map[string]any{},
		Err:  B,
	}
}

func ErrResponseStr(msg string) *ErrResponseST {

	return &ErrResponseST{
		Code: 201,
		Msg:  msg,
		Data: &map[string]any{},
		Err:  "",
	}
}

func normalizeMap(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}

	result := make(map[string]any, len(data))
	for key, value := range data {
		result[key] = normalizeValue(key, value)
	}
	return result
}

func normalizeValue(key string, value any) any {
	if value == nil {
		return emptyValueForKey(key)
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Interface:
		if val.IsNil() {
			return emptyValueForKey(key)
		}
		return normalizeValue(key, val.Elem().Interface())
	case reflect.Map:
		if val.IsNil() {
			return map[string]any{}
		}
		result := make(map[string]any, val.Len())
		for _, mapKey := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", mapKey.Interface())
			result[keyStr] = normalizeValue(keyStr, val.MapIndex(mapKey).Interface())
		}
		return result
	case reflect.Slice, reflect.Array:
		if val.Kind() == reflect.Slice && val.IsNil() {
			return []any{}
		}
		result := make([]any, 0, val.Len())
		for i := 0; i < val.Len(); i++ {
			result = append(result, normalizeValue("", val.Index(i).Interface()))
		}
		return result
	case reflect.Ptr:
		if val.IsNil() {
			return emptyValueForKey(key)
		}
	}

	return value
}

func emptyValueForKey(keyName string) any {
	key := strings.ToLower(keyName)
	switch {
	case key == "data" || key == "info" || key == "detail" || key == "details" || key == "meta":
		return map[string]any{}
	case strings.HasSuffix(key, "_data") || strings.HasSuffix(key, "_info") || strings.HasSuffix(key, "_detail"):
		return map[string]any{}
	case key == "list" || key == "items" || key == "ids" || key == "images" || key == "pics" || key == "tags":
		return []any{}
	case strings.HasSuffix(key, "_list") || strings.HasSuffix(key, "_items") || strings.HasSuffix(key, "_ids"):
		return []any{}
	case strings.HasSuffix(key, "_images") || strings.HasSuffix(key, "_pics") || strings.HasSuffix(key, "_tags"):
		return []any{}
	default:
		return ""
	}
}
