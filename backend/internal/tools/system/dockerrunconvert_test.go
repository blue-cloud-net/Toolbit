package system

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDockerRunConvertToCompose 断言 docker run 命令转 Compose 配置。
func TestDockerRunConvertToCompose(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "docker run -d --name web -p 8080:80 -e FOO=bar -v ./data:/data nginx:alpine",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if data.Direction != DockerConvertRunToCompose {
		t.Errorf("方向 = %s，期望 %s", data.Direction, DockerConvertRunToCompose)
	}
	if data.Image != "nginx:alpine" {
		t.Errorf("镜像 = %s，期望 nginx:alpine", data.Image)
	}
	if data.ServiceName != "nginx" {
		t.Errorf("服务名 = %s，期望 nginx", data.ServiceName)
	}
	for _, want := range []string{
		"image: \"nginx:alpine\"",
		"container_name: web",
		"- \"8080:80\"",
		"FOO: bar",
		"- \"./data:/data\"",
	} {
		if !strings.Contains(data.Output, want) {
			t.Errorf("配置应含 %s：\n%s", want, data.Output)
		}
	}
	assertComposeParses(t, data.Output)
}

// TestDockerRunConvertWithoutPrefix 断言省略 docker run 前缀也能解析。
func TestDockerRunConvertWithoutPrefix(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "--name web nginx:alpine",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if data.Image != "nginx:alpine" {
		t.Errorf("镜像 = %s，期望 nginx:alpine", data.Image)
	}
	if !strings.Contains(data.Output, "container_name: web") {
		t.Errorf("配置应含容器名：\n%s", data.Output)
	}
}

// TestDockerRunConvertShortFlags 断言组合短参与紧贴写法。
func TestDockerRunConvertShortFlags(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "docker run -dit -p8080:80 -eMODE=fast redis:7",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	for _, want := range []string{"tty: true", "stdin_open: true", "- \"8080:80\"", "MODE: fast"} {
		if !strings.Contains(data.Output, want) {
			t.Errorf("配置应含 %s：\n%s", want, data.Output)
		}
	}
}

// TestDockerRunConvertQuotedValue 断言带引号与空格的取值被正确切分。
func TestDockerRunConvertQuotedValue(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: `docker run -e "MSG=hello world" nginx:alpine sh -c "echo hi"`,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if !strings.Contains(data.Output, "MSG: \"hello world\"") {
		t.Errorf("环境变量应保留空格：\n%s", data.Output)
	}
	if !strings.Contains(data.Output, "- \"echo hi\"") {
		t.Errorf("命令应被正确切分：\n%s", data.Output)
	}
}

// TestDockerRunConvertServiceName 断言显式服务名生效。
func TestDockerRunConvertServiceName(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input:       "docker run nginx:alpine",
		ServiceName: "front-end",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if data.ServiceName != "front-end" {
		t.Errorf("服务名 = %s，期望 front-end", data.ServiceName)
	}
	if !strings.Contains(data.Output, "  front-end:") {
		t.Errorf("配置应含服务名：\n%s", data.Output)
	}
}

// TestDockerRunConvertNetworkMode 断言 host 网络映射为 network_mode。
func TestDockerRunConvertNetworkMode(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "docker run --network host nginx:alpine",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if !strings.Contains(data.Output, "network_mode: host") {
		t.Errorf("配置应含 network_mode：\n%s", data.Output)
	}
}

// TestDockerRunConvertComposeToRun 断言 Compose 配置转 docker run 命令。
func TestDockerRunConvertComposeToRun(t *testing.T) {
	input := `services:
  web:
    image: nginx:alpine
    container_name: web
    ports:
      - "8080:80"
    environment:
      TZ: Asia/Shanghai
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
`
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     input,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if len(data.Services) != 2 {
		t.Fatalf("服务数 = %d，期望 2", len(data.Services))
	}
	if data.Services[0].Name != "web" || data.Services[1].Name != "db" {
		t.Errorf("服务顺序应保持：%s / %s", data.Services[0].Name, data.Services[1].Name)
	}
	web := data.Services[0].RunCommand
	for _, want := range []string{"docker run", "--name web", "-p 8080:80", "-e TZ=Asia/Shanghai", "nginx:alpine"} {
		if !strings.Contains(web, want) {
			t.Errorf("命令应含 %s：%s", want, web)
		}
	}
	if !strings.Contains(data.Services[1].RunCommand, "-e POSTGRES_PASSWORD=secret") {
		t.Errorf("命令应含环境变量：%s", data.Services[1].RunCommand)
	}
}

// TestDockerRunConvertComposeLongPorts 断言长格式端口被转成短格式。
func TestDockerRunConvertComposeLongPorts(t *testing.T) {
	input := `services:
  web:
    image: nginx:alpine
    ports:
      - target: 80
        published: "8080"
        protocol: tcp
`
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     input,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if !strings.Contains(data.Services[0].RunCommand, "-p 8080:80/tcp") {
		t.Errorf("长格式端口应被转换：%s", data.Services[0].RunCommand)
	}
}

// TestDockerRunConvertComposeCommand 断言命令列表转成参数。
func TestDockerRunConvertComposeCommand(t *testing.T) {
	input := `services:
  web:
    image: nginx:alpine
    command: ["sh", "-c", "echo hi"]
`
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     input,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	command := data.Services[0].RunCommand
	if !strings.HasSuffix(command, "nginx:alpine sh -c 'echo hi'") {
		t.Errorf("命令应保留参数边界：%s", command)
	}
}

// TestDockerRunConvertComposeUnsafeValue 断言含 shell 元字符的取值被引号包裹。
func TestDockerRunConvertComposeUnsafeValue(t *testing.T) {
	input := `services:
  web:
    image: nginx:alpine
    environment:
      CMD: "$(rm -rf /)"
`
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     input,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	command := data.Services[0].RunCommand
	if !strings.Contains(command, "'CMD=$(rm -rf /)'") {
		t.Errorf("含元字符的取值必须被引号包裹：%s", command)
	}
}

// TestDockerRunConvertLenientUnsupported 断言宽松模式列出未支持的项。
func TestDockerRunConvertLenientUnsupported(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "docker run --env-file secrets.env --rm nginx:alpine",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if len(data.Unsupported) != 2 {
		t.Fatalf("未支持项 = %v，期望 2 个", data.Unsupported)
	}
	if data.Unsupported[0] != "--env-file" || data.Unsupported[1] != "--rm" {
		t.Errorf("未支持项 = %v", data.Unsupported)
	}
	if data.Image != "nginx:alpine" {
		t.Errorf("镜像 = %s，期望 nginx:alpine（取值不能被当成镜像）", data.Image)
	}
}

// TestDockerRunConvertUnsupportedHidesValue 断言未支持项不回显取值。
func TestDockerRunConvertUnsupportedHidesValue(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "docker run --label-file /run/secrets/labels nginx:alpine",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	for _, item := range data.Unsupported {
		if strings.Contains(item, "secrets") {
			t.Errorf("未支持项不得回显取值：%v", data.Unsupported)
		}
	}
}

// TestDockerRunConvertStrictRejects 断言严格模式遇到未支持项时报错。
func TestDockerRunConvertStrictRejects(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input:  "docker run --rm nginx:alpine",
		Strict: true,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestDockerRunConvertComposeUnsupportedKey 断言 Compose 未支持键被列出。
func TestDockerRunConvertComposeUnsupportedKey(t *testing.T) {
	input := `services:
  web:
    image: nginx:alpine
    depends_on:
      - db
    healthcheck:
      test: ["CMD", "true"]
`
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     input,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if len(data.Unsupported) != 2 {
		t.Fatalf("未支持项 = %v，期望 2 个", data.Unsupported)
	}
	if data.Unsupported[0] != "web.depends_on" || data.Unsupported[1] != "web.healthcheck" {
		t.Errorf("未支持项 = %v", data.Unsupported)
	}
}

// TestDockerRunConvertAutoDirectionFromCommand 断言方向与输入不符时按输入纠正：
// 指定 composeToRun 却贴了 docker run 命令，应按命令转 Compose，而不是报 YAML 错。
func TestDockerRunConvertAutoDirectionFromCommand(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     "docker run -d --name web -p 8080:80 nginx:alpine",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if data.Direction != DockerConvertRunToCompose {
		t.Errorf("方向 = %s，期望 %s", data.Direction, DockerConvertRunToCompose)
	}
	if data.Image != "nginx:alpine" {
		t.Errorf("镜像 = %s，期望 nginx:alpine", data.Image)
	}
	assertDirectionAdjusted(t, result.Warnings, DockerConvertRunToCompose)
}

// TestDockerRunConvertAutoDirectionFromCompose 断言贴了 Compose 配置时按配置纠正方向。
func TestDockerRunConvertAutoDirectionFromCompose(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertRunToCompose,
		Input:     "services:\n  web:\n    image: nginx:alpine\n",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if data.Direction != DockerConvertComposeToRun {
		t.Errorf("方向 = %s，期望 %s", data.Direction, DockerConvertComposeToRun)
	}
	if len(data.Services) != 1 || data.Services[0].Name != "web" {
		t.Errorf("服务列表 = %+v，期望单个 web 服务", data.Services)
	}
	assertDirectionAdjusted(t, result.Warnings, DockerConvertComposeToRun)
}

// TestDockerRunConvertDefaultDirectionDetectsCompose 断言未指定方向时按输入判断。
func TestDockerRunConvertDefaultDirectionDetectsCompose(t *testing.T) {
	result, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "services:\n  web:\n    image: nginx:alpine\n",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data := result.Data.(DockerRunConvertResult)
	if data.Direction != DockerConvertComposeToRun {
		t.Errorf("方向 = %s，期望 %s", data.Direction, DockerConvertComposeToRun)
	}
}

// TestDockerRunConvertAmbiguousInputKeepsDirection 断言形态不明确的输入保留指定方向。
func TestDockerRunConvertAmbiguousInputKeepsDirection(t *testing.T) {
	// 只有「是映射但没有 services」，形态不明确：保留用户指定的方向。
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     "volumes:\n  data:\n",
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// assertDirectionAdjusted 断言警告里带有方向纠正提示。
func assertDirectionAdjusted(t *testing.T, warnings []toolerr.Warning, want string) {
	t.Helper()
	for _, warning := range warnings {
		if warning.Code != toolerr.WarnDirectionAdjusted {
			continue
		}
		if got, _ := warning.Params["direction"].(string); got != want {
			t.Errorf("警告方向 = %s，期望 %s", got, want)
		}
		return
	}
	t.Errorf("缺少 %s 警告，实际警告：%+v", toolerr.WarnDirectionAdjusted, warnings)
}

// TestDockerRunConvertInvalidDirection 断言非法方向返回 UNSUPPORTED_MODE。
func TestDockerRunConvertInvalidDirection(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: "sideways",
		Input:     "docker run nginx:alpine",
	})
	assertCode(t, err, toolerr.CodeUnsupportedMode)
}

// TestDockerRunConvertMissingInput 断言空输入返回 MISSING_FIELD。
func TestDockerRunConvertMissingInput(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestDockerRunConvertMissingImage 断言没有镜像时返回 MISSING_FIELD。
func TestDockerRunConvertMissingImage(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: "docker run --name web",
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestDockerRunConvertInvalidYAML 断言 YAML 语法错误返回 INVALID_YAML。
func TestDockerRunConvertInvalidYAML(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     "services:\n  web:\n   image: x\n  bad",
	})
	assertCode(t, err, toolerr.CodeInvalidYAML)
}

// TestDockerRunConvertMissingServices 断言缺少 services 返回 MISSING_FIELD。
func TestDockerRunConvertMissingServices(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Direction: DockerConvertComposeToRun,
		Input:     "volumes:\n  data:\n",
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestDockerRunConvertInvalidServiceName 断言非法服务名返回 INVALID_FIELD。
func TestDockerRunConvertInvalidServiceName(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input:       "docker run nginx:alpine",
		ServiceName: "bad name",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestDockerRunConvertTooLarge 断言超长输入返回 TOO_LARGE。
func TestDockerRunConvertTooLarge(t *testing.T) {
	_, err := DockerRunConvert(t.Context(), &DockerRunConvertRequest{
		Input: strings.Repeat("a", dockerConvertInputLimit+1),
	})
	assertCode(t, err, toolerr.CodeTooLarge)
}

// TestSplitShellWords 断言 shell 词法切分。
func TestSplitShellWords(t *testing.T) {
	words, err := splitShellWords(`a "b c" 'd e' f\ g`)
	if err != nil {
		t.Fatalf("切分失败：%v", err)
	}
	want := []string{"a", "b c", "d e", "f g"}
	if len(words) != len(want) {
		t.Fatalf("切分结果 = %v，期望 %v", words, want)
	}
	for i, item := range want {
		if words[i] != item {
			t.Errorf("第 %d 个词 = %q，期望 %q", i, words[i], item)
		}
	}
}

// TestSplitShellWordsUnbalanced 断言引号不配对时报错。
func TestSplitShellWordsUnbalanced(t *testing.T) {
	if _, err := splitShellWords(`nginx "unclosed`); err == nil {
		t.Fatal("期望引号不配对报错")
	}
}

// assertComposeParses 断言生成的 Compose 是合法 YAML。
func assertComposeParses(t *testing.T, content string) {
	t.Helper()
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		t.Fatalf("生成的配置不是合法 YAML：%v\n%s", err, content)
	}
}
