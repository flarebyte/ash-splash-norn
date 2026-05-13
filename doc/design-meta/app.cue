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
      title: "01 Registry Schema"
      notes: ["norn.cue.design-registry-schema", "norn.cue.design-registry-example"]
    }, {
      title: "02 Key Schema And Validation"
      notes: ["norn.cue.config-key-schema", "norn.validation-source"]
    }, {
      title: "03 CUE Config Samples"
      notes: ["norn.cue.config-key"]
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
    name: "norn.cue.design-registry-schema"
    title: "CUE Design Registry Schema"
    filepath: "examples/input/design-registry.schema.cue"
    labels: ["cue", "schema", "registry", "capabilities"]
  },
  {
    name: "norn.cue.design-registry-example"
    title: "CUE Design Registry Example"
    filepath: "examples/input/design-registry.example.cue"
    labels: ["cue", "example", "registry", "capabilities"]
  },
  {
    name: "norn.cue.config-key-schema"
    title: "CUE Config Key Schema"
    filepath: "examples/input/config-key.schema.cue"
    labels: ["cue", "schema", "config", "keys"]
  },
  {
    name: "norn.validation-source"
    title: "Validation Source Of Truth"
    markdown: """
Validation commands are authored in `examples/input/config-key.cue` under the `validations` section.
This CUE input is the canonical source used to compile snake-knot-picker command documents.
"""
    labels: ["validation", "cue", "source-of-truth"]
  },
  {
    name: "norn.cue.config-key"
    title: "Key-oriented Config CUE Example"
    filepath: "examples/input/config-key.cue"
    labels: ["cue", "example", "config"]
  },
]
