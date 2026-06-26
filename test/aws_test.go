package test

import (
	"os"
	"testing"

	"nezha_cyber_mcp/internal/utilities"
)

func TestIsRunInAWS_NotEnabled(t *testing.T) {
	t.Setenv("IS_AWS", "false")
	if utilities.IsRunInAWS() {
		t.Error("IS_AWS=false 时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_MissingCredentials(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	if utilities.IsRunInAWS() {
		t.Error("凭证缺失时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_PlaceholderKeyID_MultipleX(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "multiple x")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	if utilities.IsRunInAWS() {
		t.Error("AWS_ACCESS_KEY_ID 为 'multiple x' 占位符时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_PlaceholderKeyID_MultipleXXX(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "multiple xxx")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	if utilities.IsRunInAWS() {
		t.Error("AWS_ACCESS_KEY_ID 为 'multiple xxx' 占位符时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_PlaceholderKeyID_Uppercase(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "Multiple X")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	if utilities.IsRunInAWS() {
		t.Error("AWS_ACCESS_KEY_ID 为大写 'Multiple X' 占位符时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_PlaceholderKeyID_MisspelledVariant(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "muleiplte x")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	if utilities.IsRunInAWS() {
		t.Error("AWS_ACCESS_KEY_ID 为拼写错误变体 'muleiplte x' 时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_PlaceholderSecretKey_RepeatedChar(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "xxx")

	if utilities.IsRunInAWS() {
		t.Error("AWS_SECRET_ACCESS_KEY 为纯重复字符 'xxx' 时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_PlaceholderSecretKey_RepeatedLong(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "aaaa")

	if utilities.IsRunInAWS() {
		t.Error("AWS_SECRET_ACCESS_KEY 为纯重复字符 'aaaa' 时 IsRunInAWS() 应返回 false")
	}
}

func TestIsRunInAWS_ValidCredentials(t *testing.T) {
	t.Setenv("IS_AWS", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	if !utilities.IsRunInAWS() {
		t.Error("凭证合法时 IsRunInAWS() 应返回 true")
	}
}

func TestAWSRegion_Fallback(t *testing.T) {
	os.Unsetenv("AWS_REGION")
	got := utilities.AWSRegion("ap-southeast-1")
	if got != "ap-southeast-1" {
		t.Errorf("AWSRegion fallback = %q, want %q", got, "ap-southeast-1")
	}
}

func TestAWSRegion_FromEnv(t *testing.T) {
	t.Setenv("AWS_REGION", "us-west-2")
	got := utilities.AWSRegion("ap-southeast-1")
	if got != "us-west-2" {
		t.Errorf("AWSRegion from env = %q, want %q", got, "us-west-2")
	}
}
