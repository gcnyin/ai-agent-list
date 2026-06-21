# ai-agent-list

列出本机安装的所有 AI Agent CLI 工具。

纯标准库 Go，零外部依赖，编译出来就是单文件，跨平台可用。

## 快速开始

```bash
# 编译
go build -o ai-agent-list .

# 运行
./ai-agent-list
```

## 用法

```
./ai-agent-list           # 文本输出（默认）
./ai-agent-list --json    # JSON 格式输出
./ai-agent-list --help    # 帮助
```

## 检测来源

| 来源 | CLI | 说明 |
|------|-----|------|
| `npm -g` | `npm root -g` | 扫描全局安装的 npm 包，解析 `package.json` |
| `pip` | `pip list --format=json` | 扫描 pip 用户级安装 |
| `cargo` | `cargo install --list` | 扫描 cargo 安装的 Rust CLI |
| `go install` | `$GOBIN` / `$GOPATH/bin` | 扫描 go install 的二进制 |
| `$PATH` | `exec.LookPath` | 扫描 PATH 中的命令 |

## 已知 AI Agent CLI

程序内置了以下工具的检测：

### Coding Agent

- `claude` — [@anthropic-ai/claude-code](https://www.npmjs.com/package/@anthropic-ai/claude-code)
- `codex` — [@openai/codex](https://www.npmjs.com/package/@openai/codex)
- `pi` — [@earendil-works/pi-coding-agent](https://www.npmjs.com/package/@earendil-works/pi-coding-agent)
- `reasonix` — [reasonix](https://www.npmjs.com/package/reasonix)
- `goose` — @gooseai/goose
- `continue` — @continuedev/continue
- `amazon-q` — @amazon-q/amazon-q-cli
- `aider` — aider (pip/cargo)
- `gptme` — gptme (npm/pip)
- `opencoder` — opencoder (npm)
- `opencode` — @opencode-ai/cli
- `gemini` — @gemini-cli/gemini
- `tabby` — @tabby-ml/tabby
- `interpreter` — open-interpreter (pip/cargo)
- `claude-agent` — @anthropic-ai/claude-agent-sdk
- `cursor-agent` / `devin` / `qwen` / `deepseek` / `cody` / `warp` — PATH 扫描

### 辅助工具

- `agentsview` — AI agent 会话记录与分析
- `context-mode` — MCP context window 优化插件
- `ollama` — 本地 LLM 运行器

## JSON 输出示例

```json
{
  "agents": [
    {
      "name": "claude",
      "version": "2.1.185",
      "pkg_manager": "npm",
      "pkg_name": "@anthropic-ai/claude-code",
      "path": "/usr/local/bin/claude",
      "description": "Use Claude, Anthropic's AI assistant, right from your terminal.",
      "category": "agent"
    }
  ],
  "supplementary": []
}
```

## 交叉编译

```bash
GOOS=linux   GOARCH=amd64 go build -o ai-agent-list-linux-amd64   .
GOOS=darwin  GOARCH=arm64 go build -o ai-agent-list-darwin-arm64  .
GOOS=windows GOARCH=amd64 go build -o ai-agent-list.exe           .
```
