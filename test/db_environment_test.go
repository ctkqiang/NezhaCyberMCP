package test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"nezha_cyber_mcp/internal/services"
	"nezha_cyber_mcp/internal/utilities"
)

// clearCloudRuntimeEnv 清除所有云运行时环境变量，确保 IsLocalMode() 返回 true。
// 必须使用 os.Unsetenv 而非 t.Setenv("KEY","")：
// os.Getenv 对未设置的 key 返回 ""，而 t.Setenv("KEY","") 会将 key 设置为空字符串，
// 两者在 os.Getenv 层面等价，但语义上 Unsetenv 才是"不存在"。
func clearCloudRuntimeEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"_LAMBDA_SERVER_PORT",
		"AWS_LAMBDA_RUNTIME_API",
		"FC_FUNCTION_NAME",
	} {
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })
	}
}

// ---- ResolveDBEnvironment ----

func TestResolveDBEnvironment_LocalMode_ReturnsLocal(t *testing.T) {
	clearCloudRuntimeEnv(t)

	env, err := utilities.ResolveDBEnvironment()
	if err != nil {
		t.Fatalf("本地模式下 ResolveDBEnvironment() 不应返回错误，实际: %v", err)
	}
	if env != utilities.DBEnvLocal {
		t.Errorf("env = %v, want DBEnvLocal", env)
	}
}

func TestResolveDBEnvironment_LocalMode_IgnoresISAWS(t *testing.T) {
	clearCloudRuntimeEnv(t)
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	env, err := utilities.ResolveDBEnvironment()
	if err != nil {
		t.Fatalf("本地模式下即使 IS_AWS=true 也不应返回错误，实际: %v", err)
	}
	if env != utilities.DBEnvLocal {
		t.Errorf("本地模式下应强制返回 DBEnvLocal，实际: %v", env)
	}
}

func TestResolveDBEnvironment_DBEnvironmentString(t *testing.T) {
	if utilities.DBEnvLocal.String() != "local" {
		t.Errorf("DBEnvLocal.String() = %q, want %q", utilities.DBEnvLocal.String(), "local")
	}
	if utilities.DBEnvAWS.String() != "aws" {
		t.Errorf("DBEnvAWS.String() = %q, want %q", utilities.DBEnvAWS.String(), "aws")
	}
}

// ---- ValidateConfiguration ----

func TestValidateConfiguration_LocalMode_PostgreSQL_Allowed(t *testing.T) {
	clearCloudRuntimeEnv(t)

	cfg := services.DatabaseConfiguration{Type: services.PostgreSQL}
	if err := services.ValidateConfiguration(cfg); err != nil {
		t.Errorf("本地模式下 PostgreSQL 应被允许，实际错误: %v", err)
	}
}

func TestValidateConfiguration_LocalMode_MySQL_Allowed(t *testing.T) {
	clearCloudRuntimeEnv(t)

	cfg := services.DatabaseConfiguration{Type: services.MySQL}
	if err := services.ValidateConfiguration(cfg); err != nil {
		t.Errorf("本地模式下 MySQL 应被允许，实际错误: %v", err)
	}
}

func TestValidateConfiguration_LocalMode_AuroraDSQL_Blocked(t *testing.T) {
	clearCloudRuntimeEnv(t)

	cfg := services.DatabaseConfiguration{Type: services.AmazonAuroraDSQL}
	err := services.ValidateConfiguration(cfg)
	if err == nil {
		t.Fatal("本地模式下 AmazonAuroraDSQL 应被阻止，但返回了 nil")
	}
	if !strings.Contains(err.Error(), "本地开发环境") {
		t.Errorf("错误信息应包含 '本地开发环境'，实际: %v", err)
	}
}

// TestValidateConfiguration_LocalMode_AuroraDSQL_BlockedByInitDatabase 验证
// ValidateConfiguration 防护在 InitDatabase 入口处生效，不会进入 IAM token 生成阶段。
// 使用 2 秒超时上下文，防止意外穿透到网络调用。
func TestValidateConfiguration_LocalMode_AuroraDSQL_BlockedByInitDatabase(t *testing.T) {
	clearCloudRuntimeEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := services.DatabaseConfiguration{
		Type:      services.AmazonAuroraDSQL,
		Host:      "cluster.dsql.us-east-1.on.aws",
		Port:      "5432",
		User:      "admin",
		DBName:    "postgres",
		AWSRegion: "us-east-1",
	}

	_, err := services.InitDatabase(ctx, cfg)
	if err == nil {
		t.Fatal("本地模式下 InitDatabase 使用 AmazonAuroraDSQL 应返回错误")
	}
	if !strings.Contains(err.Error(), "本地开发环境") {
		t.Errorf("错误信息应包含 '本地开发环境'，实际: %v", err)
	}
}
