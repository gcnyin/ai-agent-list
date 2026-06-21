package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// AgentCLI 表示一个 AI Agent CLI 工具
type AgentCLI struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	PkgManager  string `json:"pkg_manager"` // npm / pip / cargo / go / path
	PkgName     string `json:"pkg_name"`    // 包名（如 @anthropic-ai/claude-code）
	Path        string `json:"path"`
	Description string `json:"description"`
	Category    string `json:"category"` // agent / supplementary
}

func main() {
	outputJSON := flag.Bool("json", false, "以 JSON 格式输出")
	flag.Bool("help", false, "显示帮助")
	flag.Parse()

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("用法: ai-agent-list [--json]")
		fmt.Println()
		fmt.Println("列出本机安装的所有 AI Agent CLI 工具。")
		fmt.Println()
		fmt.Println("选项:")
		fmt.Println("  --json    以 JSON 格式输出")
		fmt.Println("  --help    显示帮助")
		return
	}

	agents := detectAll()

	if *outputJSON {
		output := struct {
			Agents        []AgentCLI `json:"agents"`
			Supplementary []AgentCLI `json:"supplementary"`
		}{
			Agents:        filterByCategory(agents, "agent"),
			Supplementary: filterByCategory(agents, "supplementary"),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
		return
	}

	// 文本输出
	fmt.Println()
	fmt.Println("  === AI Agent CLI 工具列表 ===\n")

	agentCLIs := filterByCategory(agents, "agent")
	supplementary := filterByCategory(agents, "supplementary")

	printSection("AI Coding Agent CLI", agentCLIs)
	printSection("AI 辅助工具", supplementary)

	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("  共检测到 %d 个 AI Agent CLI\n\n", len(agentCLIs))
}



// ============================================================
// 检测
// ============================================================

func detectAll() []AgentCLI {
	var all []AgentCLI
	all = append(all, detectNPMGlobal()...)
	all = append(all, detectPip()...)
	all = append(all, detectCargo()...)
	all = append(all, detectGoInstall()...)
	all = append(all, detectInPATH()...)
	all = append(all, detectExtraTools()...)
	return deduplicate(all)
}

// ---------- npm ----------

var npmAgentPackages = map[string]string{
	"@anthropic-ai/claude-code":      "claude",
	"@openai/codex":                  "codex",
	"@earendil-works/pi-coding-agent": "pi",
	"reasonix":                       "reasonix",
	"@gooseai/goose":                 "goose",
	"@continuedev/continue":          "continue",
	"@amazon-q/amazon-q-cli":         "amazon-q",
	"gptme":                          "gptme",
	"opencoder":                      "opencoder",
	"opencode":                       "opencode",
	"@anthropic-ai/claude-agent-sdk": "claude-agent",
	"@gemini-cli/gemini":             "gemini",
	"@tabby-ml/tabby":                "tabby",
}

func detectNPMGlobal() []AgentCLI {
	var result []AgentCLI
	npmRoot := runCmd("npm", "root", "-g")
	if npmRoot == "" {
		return result
	}

	for pkg, cmd := range npmAgentPackages {
		pkgDir := filepath.Join(npmRoot, pkg)
		if !dirExists(pkgDir) {
			continue
		}
		pj := filepath.Join(pkgDir, "package.json")
		ver, desc := parseNPMPackageJSON(pj)

		binPath := which(cmd)
		if binPath == "" {
			binPath = pkgDir
		}

		result = append(result, AgentCLI{
			Name:        cmd,
			Version:     ver,
			PkgManager:  "npm",
			PkgName:     pkg,
			Path:        binPath,
			Description: desc,
			Category:    "agent",
		})
	}
	return result
}

func parseNPMPackageJSON(path string) (ver, desc string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown", ""
	}
	var pkg struct {
		Version     string `json:"version"`
		Description string `json:"description"`
		Name        string `json:"name"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "unknown", ""
	}
	ver = pkg.Version
	if ver == "" {
		ver = "unknown"
	}
	desc = pkg.Description
	// fallback: local package.json 没有描述时，用 npm info 获取
	if desc == "" && pkg.Name != "" {
		desc = strings.TrimSpace(runCmd("npm", "info", pkg.Name, "description"))
	}
	return
}

// ---------- pip ----------

var pipAgentPackages = map[string]string{
	"aider-chat":       "aider",
	"open-interpreter": "interpreter",
	"gptme":            "gptme",
}

func detectPip() []AgentCLI {
	var result []AgentCLI
	output := runCmd("pip", "list", "--format=json")
	if output == "" {
		return result
	}

	var pkgs []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal([]byte(output), &pkgs); err != nil {
		return result
	}

	pipMap := map[string]bool{}
	for _, p := range pkgs {
		pipMap[strings.ToLower(p.Name)] = true
	}

	for pkg, cmd := range pipAgentPackages {
		if pipMap[strings.ToLower(pkg)] {
			ver := ""
			for _, p := range pkgs {
				if strings.EqualFold(p.Name, pkg) {
					ver = p.Version
					break
				}
			}
			result = append(result, AgentCLI{
				Name:       cmd,
				Version:    ver,
				PkgManager: "pip",
				PkgName:    pkg,
				Path:       which(cmd),
				Category:   "agent",
			})
		}
	}
	return result
}

// ---------- cargo ----------

var cargoAgentPkgs = map[string]string{
	"aider":              "aider",
	"goose":              "goose",
	"gemini-cli":         "gemini",
	"tabby":              "tabby",
	"open-interpreter":   "interpreter",
}

func detectCargo() []AgentCLI {
	var result []AgentCLI
	output := runCmd("cargo", "install", "--list")
	if output == "" {
		return result
	}

	re := regexp.MustCompile(`^(\S+)\s+v?([\d.]+[^\s:]*)`)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		ver := m[2]
		if cmd, ok := cargoAgentPkgs[name]; ok {
			result = append(result, AgentCLI{
				Name:       cmd,
				Version:    ver,
				PkgManager: "cargo",
				PkgName:    name,
				Path:       which(cmd),
				Category:   "agent",
			})
		}
	}
	return result
}

// ---------- go install ----------

func detectGoInstall() []AgentCLI {
	var result []AgentCLI
	gobin := os.Getenv("GOBIN")
	if gobin == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		gobin = filepath.Join(gopath, "bin")
	}

	// rough check
	for _, name := range []string{"gemini", "aide", "mods", "gogpt"} {
		p := filepath.Join(gobin, name)
		if fileExists(p) {
			result = append(result, AgentCLI{
				Name:       name,
				Version:    tryVersion(p),
				PkgManager: "go",
				Path:       p,
				Category:   "agent",
			})
		}
	}
	return result
}

// ---------- PATH 扫描 ----------

var pathAgentCmds = []string{
	"claude", "codex", "pi", "reasonix", "goose", "continue",
	"amazon-q", "aider", "gptme", "opencoder", "opencode", "gemini",
	"tabby", "interpreter", "claude-agent", "warp",
	"cursor-agent", "devin", "qwen", "open-interpreter",
	"cody", "cody-cli", "deepseek",
}

func detectInPATH() []AgentCLI {
	var result []AgentCLI
	seen := map[string]bool{}

	for _, cmd := range pathAgentCmds {
		p := which(cmd)
		if p == "" {
			continue
		}
		if seen[cmd] {
			continue
		}
		seen[cmd] = true

		result = append(result, AgentCLI{
			Name:       cmd,
			Version:    tryVersion(p),
			PkgManager: "path",
			Path:       p,
			Category:   "agent",
		})
	}
	return result
}

// ---------- extra supplementary tools ----------

var extraTools = []struct {
	cmd  string
	desc string
}{
	{"agentsview", "本地网页端 AI agent 会话记录与分析"},
	{"context-mode", "MCP 插件，节省 98% context window"},
	{"ollama", "本地 LLM 运行器"},
}

func detectExtraTools() []AgentCLI {
	var result []AgentCLI
	for _, t := range extraTools {
		p := which(t.cmd)
		if p == "" {
			continue
		}
		result = append(result, AgentCLI{
			Name:       t.cmd,
			Version:    tryVersion(p),
			PkgManager: "—",
			Path:       p,
			Description: t.desc,
			Category:   "supplementary",
		})
	}
	return result
}

// ============================================================
// 工具函数
// ============================================================

func runCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func which(cmd string) string {
	// Windows 特殊处理
	exe := cmd
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(cmd, ".exe") && !strings.HasSuffix(cmd, ".cmd") {
			for _, ext := range []string{".exe", ".cmd", ".bat"} {
				p, _ := exec.LookPath(cmd + ext)
				if p != "" {
					return p
				}
			}
		}
	}
	p, _ := exec.LookPath(exe)
	return p
}

func tryVersion(cmdPath string) string {
	flags := []string{"--version", "-v", "version"}
	for _, f := range flags {
		ver := runCmd(cmdPath, f)
		if ver != "" {
			return firstLine(ver)
		}
	}
	return "unknown"
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	// 截断过长
	if len(s) > 80 {
		s = s[:80] + "..."
	}
	return s
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func deduplicate(agents []AgentCLI) []AgentCLI {
	seen := map[string]bool{}
	var result []AgentCLI
	for _, a := range agents {
		key := strings.ToLower(a.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, a)
	}
	return result
}

func filterByCategory(all []AgentCLI, cat string) []AgentCLI {
	var result []AgentCLI
	for _, a := range all {
		if a.Category == cat {
			result = append(result, a)
		}
	}
	return result
}

// ============================================================
// 输出
// ============================================================

func printSection(title string, agents []AgentCLI) {
	if len(agents) == 0 {
		return
	}
	fmt.Printf("  [%s]\n\n", title)
	for _, a := range agents {
		fmt.Printf("  %-14s v%s\n", a.Name, a.Version)

		if a.PkgName != "" {
			fmt.Printf("    pkg:   %s (%s)\n", a.PkgName, a.PkgManager)
		} else if a.PkgManager != "path" && a.PkgManager != "—" {
			fmt.Printf("    pkg:   %s\n", a.PkgManager)
		}
		fmt.Printf("    path:  %s\n", a.Path)
		if a.Description != "" {
			desc := a.Description
			if len(desc) > 72 {
				desc = desc[:72] + "..."
			}
			fmt.Printf("    desc:  %s\n", desc)
		}
		fmt.Println()
	}
}
