package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 目标时间格式
const targetTimeFormat = "2006-01-02 15:04:05"

// ISO时间格式（兼容带时区的格式）
var timeLayouts = []string{
	"2006-01-02T15:04:05.999+08:00",
	"2006-01-02T15:04:05.999Z07:00",
	time.RFC3339,
	"2006-01-02T15:04:05",
}

// 重写ResponseWriter，解决重复写入问题
type rewriteResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	// 标记是否已写入响应，避免重复
	written bool
}

// Write 重写：只写入缓冲区，不调用原生Write（最终统一写入）
func (w *rewriteResponseWriter) Write(b []byte) (int, error) {
	if w.written {
		return len(b), nil
	}
	return w.body.Write(b)
}

// WriteHeader 重写：标记响应头已写入
func (w *rewriteResponseWriter) WriteHeader(statusCode int) {
	if w.written {
		return
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// FormatTimeMiddleware 修复版：解决重复响应+确保时间转换生效
func FormatTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/media/") || strings.HasPrefix(path, "/static/") {
			c.Next()
			return
		}

		// 1. 初始化包装Writer
		writer := &rewriteResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
			written:        false,
		}
		c.Writer = writer

		// 2. 放行执行后续逻辑（业务路由/其他中间件）
		c.Next()

		// 3. 跳过非JSON响应、重定向、文件下载等场景
		contentType := c.Writer.Header().Get("Content-Type")
		if !(contentType == "application/json; charset=utf-8" || contentType == "application/json") {
			// 非JSON响应，直接输出原始内容
			_, _ = writer.ResponseWriter.Write(writer.body.Bytes())
			writer.written = true
			return
		}

		// 4. 读取原始响应体
		originalBody := writer.body.Bytes()
		if len(originalBody) == 0 {
			writer.written = true
			return
		}

		// 5. 解析响应体（容错处理）
		var data interface{}
		if err := json.Unmarshal(originalBody, &data); err != nil {
			fmt.Printf("时间中间件解析响应失败：%v，原始内容：%s\n", err, string(originalBody))
			// 解析失败则输出原始内容
			_, _ = writer.ResponseWriter.Write(originalBody)
			writer.written = true
			return
		}

		// 6. 处理时间字段
		processedData := processTimeFields(data)

		// 7. 重新序列化
		newBody, err := json.Marshal(processedData)
		if err != nil {
			fmt.Printf("时间中间件序列化失败：%v\n", err)
			_, _ = writer.ResponseWriter.Write(originalBody)
			writer.written = true
			return
		}

		// 8. 统一输出处理后的响应（核心：只写一次）
		c.Writer.Header().Set("Content-Length", fmt.Sprint(len(newBody)))
		_, _ = writer.ResponseWriter.Write(newBody)
		writer.written = true
	}
}

// processTimeFields 递归处理所有时间字段（增强兼容性）
func processTimeFields(data interface{}) interface{} {
	return processResponseFields(data, "")
}

func processResponseFields(data interface{}, keyName string) interface{} {
	if data == nil {
		return emptyValueForKey(keyName)
	}

	val := reflect.ValueOf(data)
	// 处理指针类型
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return emptyValueForKey(keyName)
		}
		return processResponseFields(val.Elem().Interface(), keyName)
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		if val.Kind() == reflect.Slice && val.IsNil() {
			return []interface{}{}
		}
		result := make([]interface{}, 0, val.Len())
		for i := 0; i < val.Len(); i++ {
			result = append(result, processResponseFields(val.Index(i).Interface(), ""))
		}
		return result

	case reflect.Map:
		if val.IsNil() {
			return map[string]interface{}{}
		}
		result := make(map[string]interface{}, val.Len())
		for _, key := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			value := val.MapIndex(key).Interface()

			// 重点：处理ISO格式字符串时间
			if strVal, ok := value.(string); ok && strVal != "" {
				for _, layout := range timeLayouts {
					t, err := time.Parse(layout, strVal)
					if err == nil {
						result[keyStr] = t.Format(targetTimeFormat)
						goto NextField // 解析成功则跳过其他格式
					}
				}
				// 不是时间字符串，直接赋值
				result[keyStr] = strVal
			} else {
				// 非字符串类型，递归处理
				result[keyStr] = processResponseFields(value, keyStr)
			}
		NextField:
		}
		return result

	case reflect.Struct:
		// 处理time.Time结构体
		if t, ok := data.(time.Time); ok {
			return t.Format(targetTimeFormat)
		}
		// 处理普通结构体
		result := make(map[string]interface{})
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			if field.PkgPath != "" {
				continue
			}
			fieldVal := val.Field(i).Interface()
			// 解析JSON标签
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				jsonTag = field.Name
			} else if idx := bytes.IndexByte([]byte(jsonTag), ','); idx != -1 {
				jsonTag = jsonTag[:idx]
			}
			result[jsonTag] = processResponseFields(fieldVal, jsonTag)
		}
		return result

	case reflect.Int, reflect.Int64:
		// 处理时间戳
		timestamp := val.Int()
		if timestamp > 0 && timestamp < 4102444800 {
			return time.Unix(timestamp, 0).Format(targetTimeFormat)
		}
		return data

	case reflect.Interface:
		if val.IsNil() {
			return emptyValueForKey(keyName)
		}
		return processResponseFields(val.Elem().Interface(), keyName)

	default:
		return data
	}
}

func emptyValueForKey(keyName string) interface{} {
	key := strings.ToLower(keyName)
	switch {
	case key == "data" || key == "info" || key == "detail" || key == "details" || key == "meta":
		return map[string]interface{}{}
	case strings.HasSuffix(key, "_data") || strings.HasSuffix(key, "_info") || strings.HasSuffix(key, "_detail"):
		return map[string]interface{}{}
	case key == "list" || key == "items" || key == "ids" || key == "images" || key == "pics" || key == "tags":
		return []interface{}{}
	case strings.HasSuffix(key, "_list") || strings.HasSuffix(key, "_items") || strings.HasSuffix(key, "_ids"):
		return []interface{}{}
	case strings.HasSuffix(key, "_images") || strings.HasSuffix(key, "_pics") || strings.HasSuffix(key, "_tags"):
		return []interface{}{}
	default:
		return ""
	}
}
