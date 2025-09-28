package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

var (
	dryRun = flag.Bool("dry-run", true, "仅打印要删除的条目，不修改文件")
	quiet  = flag.Bool("q", false, "静默模式，只输出错误")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// 1. 找到 go.mod
	modPath := filepath.Join(".", "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return err
	}
	f, err := modfile.Parse(modPath, data, nil)
	if err != nil {
		return err
	}

	// 2. 收集主模块所有依赖包路径
	deps, err := listDeps()
	if err != nil {
		return err
	}

	// 3. 扫描 replace
	var toDel []string
	for _, r := range f.Replace {
		p := r.Old.Path
		if !deps[p] {
			toDel = append(toDel, p)
		}
	}
	if len(toDel) == 0 {
		if !*quiet {
			fmt.Println("=> 所有 replace 包都在使用，无需清理")
		}
		return nil
	}

	// 4. 输出/删除
	for _, p := range toDel {
		if *dryRun {
			fmt.Printf("DRY-RUN:  go mod edit -droprequire %s  &&  go mod edit -dropreplace %s\n", p, p)
			continue
		}
		if err := drop(p); err != nil {
			return err
		}
	}

	// 5. tidy
	if !*dryRun {
		return tidy()
	}
	return nil
}

// listDeps 返回主模块（含测试）依赖的所有包路径集合
func listDeps() (map[string]bool, error) {
	m := make(map[string]bool)
	cmd := exec.Command("go", "list", "-e", "-deps", "./...")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	for _, line := range bytes.Split(out, []byte("\n")) {
		p := strings.TrimSpace(string(line))
		if p != "" {
			m[p] = true
		}
	}
	return m, nil
}

func drop(path string) error {
	for _, argv := range [][]string{
		{"go", "mod", "edit", "-droprequire", path},
		{"go", "mod", "edit", "-dropreplace", path},
	} {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%v: %w", argv, err)
		}
	}
	return nil
}

func tidy() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
