package utilities

import (
	"fmt"
	"os"
	"strings"
)

// IsRunInAWS 检查当前进程是否运行在 AWS 环境中，并验证 AWS 凭证的有效性。
//
// 判断逻辑：
//  1. 读取环境变量 IS_AWS；若不为 "true" 则直接返回 false
//  2. 当 IS_AWS=true 时，验证 AWS_ACCESS_KEY_ID 和 AWS_SECRET_ACCESS_KEY
//     均不为空且不包含无效占位符值
//
// 返回：
//   - true  : IS_AWS=true 且两个凭证均通过校验
//   - false : IS_AWS!=true，或任意凭证校验失败
func IsRunInAWS() bool {
	isAws := awsEnv("IS_AWS", "false") == "true"

	LogProgress("AWSUtil", "IsRunInAWS", fmt.Sprintf("IS_AWS=%v", isAws))

	if !isAws {
		return false
	}

	keyID := awsEnv("AWS_ACCESS_KEY_ID", "")
	secretKey := awsEnv("AWS_SECRET_ACCESS_KEY", "")

	if isInvalidAWSCredential(keyID) {
		err := fmt.Errorf("AWS_ACCESS_KEY_ID 缺失或包含无效占位符值 %q", keyID)
		LogError("AWSUtil", "IsRunInAWS", err, 0, "key=AWS_ACCESS_KEY_ID")
		return false
	}

	if isInvalidAWSCredential(secretKey) {
		err := fmt.Errorf("AWS_SECRET_ACCESS_KEY 缺失或包含无效占位符值 %q", secretKey)
		LogError("AWSUtil", "IsRunInAWS", err, 0, "key=AWS_SECRET_ACCESS_KEY")
		return false
	}

	LogProgress("AWSUtil", "IsRunInAWS", "AWS 凭证校验通过")
	return true
}

// AWSRegion 返回当前配置的 AWS 区域（环境变量 AWS_REGION），
// 若未设置则返回 fallback 默认值。
func AWSRegion(fallback string) string {
	return awsEnv("AWS_REGION", fallback)
}

// isInvalidAWSCredential 判断一个 AWS 凭证值是否为无效占位符。
//
// 无效条件（满足任意一条即判定为无效）：
//  1. 空字符串
//  2. 以 "multiple " 或 "muleiplte " 开头（大小写不敏感）
//     涵盖："multiple x", "multiple xxx", "muleiplte x" 等占位符形式
//  3. 整个字符串由同一个字符重复三次或以上组成
//     涵盖："xxx", "xxxxxxxxxxx" 等纯重复占位符
func isInvalidAWSCredential(v string) bool {
	if v == "" {
		return true
	}

	lower := strings.ToLower(v)
	for _, prefix := range []string{"multiple ", "muleiplte "} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	if len(v) >= 3 {
		allSame := true
		for i := 1; i < len(v); i++ {
			if v[i] != v[0] {
				allSame = false
				break
			}
		}
		if allSame {
			return true
		}
	}

	return false
}

// awsEnv 读取指定环境变量的值；若未设置或为空则返回 fallback。
func awsEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// IsLocalMode reports whether none of the supported cloud runtimes are
// detected, meaning the local development HTTP server should start.
//
// Returns:
//   - bool: true when neither AWS Lambda nor Alibaba Cloud FC runtime
//     environment variables are present.
func IsLocalMode() bool {
	_, lambdaPort := os.LookupEnv("_LAMBDA_SERVER_PORT")
	_, lambdaAPI := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	_, fcFunc := os.LookupEnv("FC_FUNCTION_NAME")
	onAWS := lambdaPort && lambdaAPI
	onAliyun := fcFunc

	return !onAWS && !onAliyun
}
