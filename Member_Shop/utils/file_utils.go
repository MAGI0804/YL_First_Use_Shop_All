package utils

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gin-gonic/gin"
)

// SaveUploadedFile 保存上传的文件到指定目录（仅生成路径，不实际保存）
// 参数:
// - c: Gin上下文
// - file: 上传的文件
// - directory: 保存目录
// - prefix: 文件名前缀
// 返回值:
// - 相对路径
// - 错误信息
func SaveUploadedFile(c *gin.Context, file interface{}, directory string, prefix string) (string, error) {
	// 使用反射获取文件名
	fileValue := reflect.ValueOf(file)
	if fileValue.Kind() != reflect.Ptr || fileValue.IsNil() {
		return "", fmt.Errorf("无效的文件参数")
	}

	// 尝试获取Filename字段
	filenameField := fileValue.Elem().FieldByName("Filename")
	if !filenameField.IsValid() || filenameField.Kind() != reflect.String {
		return "", fmt.Errorf("无法获取文件名")
	}

	filename := filenameField.String()
	if filename == "" {
		return "", fmt.Errorf("文件名为空")
	}

	// 生成唯一文件名
	uniqueFilename := GenerateUniqueFilename(filename)
	if prefix != "" {
		uniqueFilename = prefix + uniqueFilename
	}

	// 定义保存路径 - 只保存相对路径，不包含media前缀
	savePath := filepath.Join(directory, uniqueFilename)

	// 确保目录存在，使用 0755 权限
	fullDir := MediaPath(directory)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 确保目录权限正确
	if err := os.Chmod(fullDir, 0755); err != nil {
		log.Printf("设置目录权限失败: %v", err)
	}

	// 使用gin的上下文保存文件
	// 注意：这是一个简化的实现，假设调用者已经通过c.FormFile获取了文件
	// 实际保存文件的逻辑应该在控制器中完成
	return savePath, nil
}

// SaveFileWithPerms 保存文件并设置正确的权限
// 参数:
// - c: Gin上下文
// - file: 上传的文件
// - directory: 保存目录（相对路径，不包含media前缀）
// - prefix: 文件名前缀
// 返回值:
// - 相对路径
// - 完整路径
// - 错误信息
func SaveFileWithPerms(c *gin.Context, file *multipart.FileHeader, directory string, prefix string) (string, string, error) {
	// 生成保存路径
	savePath, err := SaveUploadedFile(c, file, directory, prefix)
	if err != nil {
		return "", "", err
	}

	// 完整路径
	fullPath := MediaPath(savePath)

	// 实际保存文件
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		return "", "", fmt.Errorf("保存文件失败: %v", err)
	}

	// 设置文件权限为 0644（读权限给所有人），确保 nginx 可以访问
	if err := os.Chmod(fullPath, 0644); err != nil {
		log.Printf("设置文件权限失败: %v", err)
	}

	// 确保目录权限正确
	dirPath := filepath.Dir(fullPath)
	if err := os.Chmod(dirPath, 0755); err != nil {
		log.Printf("设置目录权限失败: %v", err)
	}

	return savePath, fullPath, nil
}
