package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// FlutterFileValidator Flutter文件验证器
type FlutterFileValidator struct{}

// CalculateSHA256 计算文件的SHA256哈希值
func (f *FlutterFileValidator) CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件 %s: %v", filePath, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("读取文件时出错: %v", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// VerifySHA256 验证文件的SHA256哈希值
func (f *FlutterFileValidator) VerifySHA256(filePath string, expectedSHA256 string) (bool, error) {
	actualSHA256, err := f.CalculateSHA256(filePath)
	if err != nil {
		return false, err
	}

	return actualSHA256 == expectedSHA256, nil
}

// ValidateFlutterRelease 验证Flutter发布文件
func (f *FlutterFileValidator) ValidateFlutterRelease(filePath string, expectedSHA256 string) error {
	fmt.Printf("验证Flutter发布文件: %s\n", filePath)
	fmt.Printf("期望的SHA256: %s\n", expectedSHA256)

	actualSHA256, err := f.CalculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("计算SHA256时出错: %v", err)
	}

	fmt.Printf("实际的SHA256: %s\n", actualSHA256)

	if actualSHA256 == expectedSHA256 {
		fmt.Println("✓ 文件验证成功")
		return nil
	} else {
		fmt.Println("✗ 文件验证失败")
		return fmt.Errorf("SHA256校验和不匹配")
	}
}

// 示例使用
func main() {
	validator := &FlutterFileValidator{}

	// 如果有命令行参数
	if len(os.Args) >= 2 {
		filePath := os.Args[1]

		// 计算SHA256
		sha256Hash, err := validator.CalculateSHA256(filePath)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			return
		}

		fmt.Printf("文件 %s 的SHA256: %s\n", filePath, sha256Hash)

		// 如果提供了期望的SHA256值
		if len(os.Args) >= 3 {
			expectedSHA256 := os.Args[2]
			isValid, err := validator.VerifySHA256(filePath, expectedSHA256)
			if err != nil {
				fmt.Printf("验证时出错: %v\n", err)
				return
			}

			if isValid {
				fmt.Println("✓ SHA256验证通过")
			} else {
				fmt.Println("✗ SHA256验证失败")
			}
		}
	} else {
		fmt.Println("用法:")
		fmt.Println("  go run main.go <文件路径>                    # 计算SHA256")
		fmt.Println("  go run main.go <文件路径> <期望的SHA256>     # 验证SHA256")
		// E:\workSpace\mirror\flutter-mirror\cache\flutter_infra_release\flutter\a18df97ca57a249df5d8d68cd0820600223ce262\flutter_gpu.zip
		// https://storage.flutter-io.cn/flutter_infra_release/flutter/0009cc358ff7e2c06d67b239cfa1f054cff93132/flutter_gpu.zip
	}
}
