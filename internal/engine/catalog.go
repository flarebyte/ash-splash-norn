package engine

import (
	"fmt"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

type ListRow struct {
	SchemaID        string   `json:"schemaId"`
	Target          string   `json:"target"`
	SupportsNode    []string `json:"supportsNodeKinds"`
	ArtifactPattern string   `json:"artifactPattern"`
}

type registryCatalog struct {
	DesignRegistry struct {
		KeySchemaRegistry     map[string]struct{} `cue:"keySchemaRegistry"`
		GeneratorCapabilities []struct {
			KeySchema         string   `cue:"keySchema"`
			Target            string   `cue:"target"`
			SupportsNodeKinds []string `cue:"supportsNodeKinds"`
			ArtifactPattern   string   `cue:"artifactPattern"`
		} `cue:"generatorCapabilities"`
	} `cue:"designRegistry"`
}

type keyExplainRegistry struct {
	DesignRegistry struct {
		KeySchemaRegistry map[string]struct {
			RootLabels   []string `cue:"rootLabels"`
			NodesByLabel map[string]struct {
				Label       string   `cue:"label"`
				Kind        string   `cue:"kind"`
				ChildLabels []string `cue:"childLabels"`
			} `cue:"nodesByLabel"`
		} `cue:"keySchemaRegistry"`
	} `cue:"designRegistry"`
}

type ExplainStep struct {
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

type ExplainTrace struct {
	SchemaID string        `json:"schemaId"`
	Labels   []string      `json:"labels"`
	Steps    []ExplainStep `json:"steps"`
	Key      string        `json:"key"`
}

func BuildCatalog(in app.Inputs) ([]ListRow, []diag.Entry) {
	src, entries := mergeCueSources(in.RegistrySchemaPath, in.RegistryPath, "list")
	if len(entries) > 0 {
		return nil, entries
	}
	ctx := cuecontext.New()
	v := ctx.CompileString(src, cue.Filename("list.cue"))
	if err := v.Err(); err != nil {
		return nil, []diag.Entry{{Stage: "list", ID: "LST-0001", Severity: diag.SeverityError, Message: fmt.Sprintf("registry compile failed: %v", sanitizeCueErr(err)), Path: in.RegistryPath}}
	}
	var reg registryCatalog
	if err := v.Decode(&reg); err != nil {
		return nil, []diag.Entry{{Stage: "list", ID: "LST-0002", Severity: diag.SeverityError, Message: fmt.Sprintf("registry decode failed: %v", sanitizeCueErr(err)), Path: in.RegistryPath}}
	}
	rows := make([]ListRow, 0, len(reg.DesignRegistry.GeneratorCapabilities))
	for _, c := range reg.DesignRegistry.GeneratorCapabilities {
		rows = append(rows, ListRow{
			SchemaID: c.KeySchema, Target: c.Target, SupportsNode: append([]string(nil), c.SupportsNodeKinds...), ArtifactPattern: c.ArtifactPattern,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].SchemaID != rows[j].SchemaID {
			return rows[i].SchemaID < rows[j].SchemaID
		}
		if rows[i].Target != rows[j].Target {
			return rows[i].Target < rows[j].Target
		}
		return rows[i].ArtifactPattern < rows[j].ArtifactPattern
	})
	return rows, nil
}

func ExplainKey(in app.Inputs, schemaID string, labels []string) (string, []diag.Entry) {
	trace, entries := ExplainKeyTrace(in, schemaID, labels)
	if len(entries) > 0 {
		return "", entries
	}
	return trace.Key, nil
}

func ExplainKeyTrace(in app.Inputs, schemaID string, labels []string) (ExplainTrace, []diag.Entry) {
	empty := ExplainTrace{}
	if len(labels) == 0 {
		return empty, []diag.Entry{{Stage: "explain-key", ID: "KEY-0001", Severity: diag.SeverityError, Message: "label path is required"}}
	}
	src, entries := mergeCueSources(in.RegistrySchemaPath, in.RegistryPath, "explain-key")
	if len(entries) > 0 {
		return empty, entries
	}
	ctx := cuecontext.New()
	v := ctx.CompileString(src, cue.Filename("explain-key.cue"))
	if err := v.Err(); err != nil {
		return empty, []diag.Entry{{Stage: "explain-key", ID: "KEY-0002", Severity: diag.SeverityError, Message: fmt.Sprintf("registry compile failed: %v", sanitizeCueErr(err)), Path: in.RegistryPath}}
	}
	var reg keyExplainRegistry
	if err := v.Decode(&reg); err != nil {
		return empty, []diag.Entry{{Stage: "explain-key", ID: "KEY-0003", Severity: diag.SeverityError, Message: fmt.Sprintf("registry decode failed: %v", sanitizeCueErr(err)), Path: in.RegistryPath}}
	}
	if schemaID == "" {
		for k := range reg.DesignRegistry.KeySchemaRegistry {
			schemaID = k
			break
		}
	}
	ks, ok := reg.DesignRegistry.KeySchemaRegistry[schemaID]
	if !ok {
		return empty, []diag.Entry{{Stage: "explain-key", ID: "KEY-0004", Severity: diag.SeverityError, Message: fmt.Sprintf("schema not found: %s", schemaID)}}
	}
	steps := make([]ExplainStep, 0, len(labels))
	for _, l := range labels {
		n, ok := ks.NodesByLabel[l]
		if !ok {
			return empty, []diag.Entry{{Stage: "explain-key", ID: "KEY-0005", Severity: diag.SeverityError, Message: fmt.Sprintf("unknown label in path: %s", l)}}
		}
		steps = append(steps, ExplainStep{Label: n.Label, Kind: n.Kind})
	}
	key := toPathCamel(labels)
	return ExplainTrace{
		SchemaID: schemaID,
		Labels:   append([]string(nil), labels...),
		Steps:    steps,
		Key:      key,
	}, nil
}
