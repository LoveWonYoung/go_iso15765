package log_recorder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// NowString 返回当前时间格式为 "20060102 1504" 的字符串
func NowString() string {
	return time.Now().Format("20060102_1504")
}

// MakeDir 创建以日期命名的目录（如：2025_04_25）
func MakeDir() (string, error) {
	now := time.Now()
	dirName := fmt.Sprintf("%d_%02d_%02d", now.Year(), now.Month(), now.Day())
	fullPath := filepath.Join(".", dirName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return "", fmt.Errorf("创建文件夹失败: %w", err)
		}
		fmt.Println("文件夹已创建:", fullPath)
	}

	return fullPath, nil
}

// NowTimeStringSub 返回当前时间字符串（格式: "1504" -> 小时分钟）
func NowTimeStringSub() string {
	return time.Now().Format("1504")
}

// RecorderAsNameInit 初始化日志记录器，name为日志文件前缀名
func RecorderAsNameInit(name string) error {
	log.SetPrefix("")
	log.SetFlags(log.Lmicroseconds)

	dir, err := MakeDir()
	if err != nil {
		return err
	}

	logPath := filepath.Join(dir, fmt.Sprintf("%s.log", name))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	log.SetOutput(f)
	return nil
}
