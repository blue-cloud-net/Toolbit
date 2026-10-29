package system

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDockerComposeGen 断言生成 Compose 配置。
func TestDockerComposeGen(t *testing.T) {
	result, err := DockerComposeGen(t.Context(), &DockerComposeRequest{
		ProjectName: "demo",
		Services: []ComposeService{
			{Name: "web", Image: "nginx:alpine", Ports: []string{"8080:80"}},
		},
	})
	data := result.Data.(DockerComposeResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if data.ServiceCount != 1 {
		t.Errorf("服务数 = %d，期望 1", data.ServiceCount)
	}
	if !strings.Contains(data.Compose, "nginx:alpine") {
		t.Errorf("配置应含镜像名：%s", data.Compose)
	}
}

// TestDockerComposeGenMissingProject 断言缺少项目名返回 MISSING_FIELD。
func TestDockerComposeGenMissingProject(t *testing.T) {
	_, err := DockerComposeGen(t.Context(), &DockerComposeRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}
