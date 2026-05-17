// purpose: Loads and compiles registry/config CUE documents for lint workflows.
// responsibilities: Merge sources, decode lint models, and map config-input loader diagnostics into lint IDs.
// architecture notes: Input-loading error remapping is centralized so lint callers get stable ID contracts.

package engine

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

func loadLintDocs(in app.Inputs) (lintRegistry, lintConfig, []diag.Entry) {
	var reg lintRegistry
	var cfg lintConfig

	regMerged, entries := mergeCueSources(in.RegistrySchemaPath, in.RegistryPath, "lint")
	if len(entries) > 0 {
		return reg, cfg, entries
	}
	regVal, regErr := compileCue("lint-registry", regMerged)
	if regErr != nil {
		return reg, cfg, []diag.Entry{*regErr}
	}
	if err := regVal.Decode(&reg); err != nil {
		return reg, cfg, []diag.Entry{{Stage: "lint", ID: "LNT-0101", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to decode registry: %v", sanitizeCueErr(err)), Path: in.RegistryPath}}
	}

	cfgInputBody, cfgEntries := readConfigInputSource(in.ConfigPath, "lint")
	if len(cfgEntries) > 0 {
		out := make([]diag.Entry, 0, len(cfgEntries))
		for _, e := range cfgEntries {
			switch e.ID {
			case "CFG-0001", "CFG-0002", "CFG-0003", "CFG-0005":
				e.ID = "LNT-0103"
			case "CFG-0004":
				e.ID = "LNT-0105"
			case "CFG-0006":
				e.ID = "LNT-0106"
			default:
				e.ID = "LNT-0103"
			}
			e.Stage = "lint"
			out = append(out, e)
		}
		return reg, cfg, out
	}
	cfgVal, cfgErr := compileCue("lint-config", cfgInputBody)
	if cfgErr != nil {
		return reg, cfg, []diag.Entry{*cfgErr}
	}
	if err := cfgVal.Decode(&cfg); err != nil {
		return reg, cfg, []diag.Entry{{Stage: "lint", ID: "LNT-0104", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to decode config: %v", sanitizeCueErr(err)), Path: in.ConfigPath}}
	}
	return reg, cfg, nil
}

func compileCue(stage, src string) (cue.Value, *diag.Entry) {
	ctx := cuecontext.New()
	v := ctx.CompileString(src, cue.Filename(stage+".cue"))
	if err := v.Err(); err != nil {
		return cue.Value{}, &diag.Entry{Stage: "lint", ID: "LNT-0100", Severity: diag.SeverityError, Message: fmt.Sprintf("cue compile failed: %v", sanitizeCueErr(err))}
	}
	return v, nil
}
