package flyb

source: "ash-splash-norn-design-meta"
name:   "ash-splash-norn-design-meta"
modules: ["core"]

reports: [{
  title:       "Distributed Config CLI Design"
  filepath:    "../design/distributed-config-cli.md"
  description: "Design source for a Go CLI that manages distributed configurations across repositories."
  sections: [{
    title:       "01 Overview"
    description: "Scope, goals, and output targets."
    sections: [{
      title: "01 Intent"
      notes: ["norn.intent", "norn.outputs"]
    }]
  }, {
    title:       "02 Examples"
    description: "Concrete examples collected under doc/design-meta/examples."
    sections: [{
      title: "01 CLI Model"
      notes: ["norn.ts.cli", "norn.ts.app"]
    }, {
      title: "02 Generator Capabilities"
      notes: ["norn.ts.generator-capabilities"]
    }, {
      title: "03 Schema And Validation Model"
      notes: ["norn.ts.i18n-key-schema", "norn.ts.validation"]
    }, {
      title: "04 CUE Config Samples"
      notes: ["norn.cue.distributed", "norn.cue.config-key"]
    }]
  }]
}]

notes: [
  {
    name: "norn.intent"
    title: "Distributed Configuration Management"
    markdown: """
This design models a Go CLI that reads CUE configuration sources and distributes generated artifacts to multiple codebases.
"""
    labels: ["overview", "cli"]
  },
  {
    name: "norn.outputs"
    title: "Generated Output Targets"
    markdown: """
Supported generated outputs are JSON, YAML, Go code, and Dart code.
For i18n keys in Flutter/Dart contexts, the preferred target is `*.arb.json`.
Target compatibility is determined by the CLI capabilities, not by extra user-provided per-kind maps.
"""
    labels: ["overview", "outputs"]
  },
  {
    name: "norn.ts.cli"
    title: "TypeScript CLI Domain Model"
    filepath: "examples/distributed-config-cli.ts"
    labels: ["typescript", "example", "cli"]
  },
  {
    name: "norn.ts.app"
    title: "Application Composition Model"
    filepath: "examples/app.ts"
    labels: ["typescript", "example", "application"]
  },
  {
    name: "norn.ts.generator-capabilities"
    title: "Generator Capabilities Matrix"
    filepath: "examples/generator-capabilities.ts"
    labels: ["typescript", "example", "capabilities", "targets"]
  },
  {
    name: "norn.ts.i18n-key-schema"
    title: "I18n Key Schema Hierarchy Model"
    filepath: "examples/i18n-key-schema-model.ts"
    labels: ["typescript", "example", "i18n", "schema"]
  },
  {
    name: "norn.ts.validation"
    title: "Validation Commands By Key Path"
    filepath: "examples/validation.ts"
    labels: ["typescript", "example", "validation", "cli"]
  },
  {
    name: "norn.cue.distributed"
    title: "Distributed Config CUE Example"
    filepath: "examples/distributed-config-example.cue"
    labels: ["cue", "example", "distribution"]
  },
  {
    name: "norn.cue.config-key"
    title: "Key-oriented Config CUE Example"
    filepath: "examples/config-key.cue"
    labels: ["cue", "example", "config"]
  },
]
