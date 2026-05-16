package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

func readConfigInputSource(path string, stage string) (string, []diag.Entry) {
	st, err := os.Stat(path)
	if err != nil {
		return "", []diag.Entry{{Stage: stage, ID: "CFG-0001", Severity: diag.SeverityError, Message: fmt.Sprintf("config path not found: %v", err), Path: path}}
	}
	if !st.IsDir() {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", []diag.Entry{{Stage: stage, ID: "CFG-0002", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to read config input: %v", err), Path: path}}
		}
		return string(b), nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", []diag.Entry{{Stage: stage, ID: "CFG-0003", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to read config directory: %v", err), Path: path}}
	}
	files := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".cue") {
			files = append(files, filepath.Join(path, e.Name()))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return "", []diag.Entry{{Stage: stage, ID: "CFG-0004", Severity: diag.SeverityError, Message: "config directory contains no .cue files", Path: path}}
	}

	pkgs := map[string]struct{}{}
	parts := make([]string, 0, len(files))
	for _, fp := range files {
		b, err := os.ReadFile(fp)
		if err != nil {
			return "", []diag.Entry{{Stage: stage, ID: "CFG-0005", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to read config file: %v", err), Path: fp}}
		}
		src := string(b)
		pkg := extractPackage(src)
		if pkg != "" {
			pkgs[pkg] = struct{}{}
		}
		parts = append(parts, stripPackageDecl(src))
	}
	if len(pkgs) > 1 {
		pkgList := make([]string, 0, len(pkgs))
		for p := range pkgs {
			pkgList = append(pkgList, p)
		}
		sort.Strings(pkgList)
		return "", []diag.Entry{{Stage: stage, ID: "CFG-0006", Severity: diag.SeverityError, Message: fmt.Sprintf("config directory must use a single package, found: %v", pkgList), Path: path}}
	}
	pkg := ""
	for p := range pkgs {
		pkg = p
	}
	var merged strings.Builder
	if pkg != "" {
		merged.WriteString("package ")
		merged.WriteString(pkg)
		merged.WriteString("\n\n")
	}
	for i, part := range parts {
		if i > 0 {
			merged.WriteString("\n\n")
		}
		merged.WriteString(part)
	}
	return merged.String(), nil
}
