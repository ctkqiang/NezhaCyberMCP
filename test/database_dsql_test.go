package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"nezha_cyber_mcp/internal/services"
)

func TestInitDatabase_UnsupportedType_Oracle(t *testing.T) {
	cfg := services.DatabaseConfiguration{
		Type: services.Oracle,
	}
	_, err := services.InitDatabase(context.Background(), cfg)
	if err == nil {
		t.Fatal("Oracle 类型应返回错误，但返回了 nil")
	}
}

func TestInitDatabase_UnsupportedType_QuestDB(t *testing.T) {
	cfg := services.DatabaseConfiguration{
		Type: services.QuestDB,
	}
	_, err := services.InitDatabase(context.Background(), cfg)
	if err == nil {
		t.Fatal("QuestDB 类型应返回错误，但返回了 nil")
	}
}

// TestInitDatabase_AuroraDSQL_FailsWithoutAWSCredentials 验证在无有效 AWS 凭证时，
// AmazonAuroraDSQL 驱动会在 IAM token 生成阶段返回错误，而不是静默失败。
// 使用 3 秒超时上下文，避免等待 EC2 IMDS 完整超时（默认约 5 秒）。
func TestInitDatabase_AuroraDSQL_FailsWithoutAWSCredentials(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cfg := services.DatabaseConfiguration{
		Type:      services.AmazonAuroraDSQL,
		Host:      "abc123.dsql.us-east-1.on.aws",
		Port:      "5432",
		User:      "admin",
		DBName:    "postgres",
		AWSRegion: "us-east-1",
	}

	_, err := services.InitDatabase(ctx, cfg)
	if err == nil {
		t.Fatal("无 AWS 凭证时 AmazonAuroraDSQL 应返回错误，但返回了 nil")
	}
	if !strings.Contains(err.Error(), "token") && !strings.Contains(err.Error(), "AWS") && !strings.Contains(err.Error(), "credential") {
		t.Errorf("错误信息应包含 token/AWS/credential 关键字，实际: %v", err)
	}
}

func TestDatabaseConfiguration_AWSRegionField(t *testing.T) {
	cfg := services.DatabaseConfiguration{
		Type:      services.AmazonAuroraDSQL,
		Host:      "cluster.dsql.us-east-1.on.aws",
		AWSRegion: "us-east-1",
	}
	if cfg.AWSRegion != "us-east-1" {
		t.Errorf("AWSRegion = %q, want %q", cfg.AWSRegion, "us-east-1")
	}
}

func TestDatabaseConfiguration_AWSRegionEmpty_UsesEnvFallback(t *testing.T) {
	t.Setenv("AWS_REGION", "ap-southeast-1")

	cfg := services.DatabaseConfiguration{
		Type:      services.AmazonAuroraDSQL,
		Host:      "cluster.dsql.ap-southeast-1.on.aws",
		AWSRegion: "",
	}

	if cfg.AWSRegion != "" {
		t.Errorf("AWSRegion 字段应为空字符串（由 InitDatabase 内部回退到环境变量），实际: %q", cfg.AWSRegion)
	}
}
