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
      title: "03 Output Targets"
      notes: ["norn.outputs.catalog"]
    }, {
      title: "04 CLI Commands"
      notes: ["norn.cli.commands"]
    }, {
      title: "05 Implementation Libraries"
      notes: ["norn.impl.libraries"]
    }, {
      title: "06 Implementation Suggestions"
      notes: ["norn.impl.suggestions"]
    }, {
      title: "07 CUE Config Samples"
      notes: ["norn.cue.config-key"]
    }, {
      title: "08 Output Samples"
      notes: ["norn.out.text-json", "norn.out.validator-json"]
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
    filepath: "examples/model/design-registry.schema.cue"
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
    filepath: "examples/model/config-key.schema.cue"
    labels: ["cue", "schema", "config", "keys"]
  },
  {
    name: "norn.validation-source"
    title: "Validation Source Of Truth"
    markdown: """
Validation commands are authored in `examples/input/config-key.cue` under the `validations` section.
This CUE input is the canonical source used to compile snake-knot-picker command documents.

Mandatory behavior: if a reachable schema node is marked `mandatory: true`,
the corresponding key entry must exist in the matching config section by node kind
(`i18nEntries`, `textEntries`, or `validations`), otherwise lint must raise an error.
"""
    labels: ["validation", "cue", "source-of-truth"]
  },
  {
    name: "norn.outputs.catalog"
    title: "Output Target Catalog"
    filepath: "examples/other/output-targets.csv"
    arguments: ["format-csv=table"]
    labels: ["outputs", "targets", "csv"]
  },
  {
    name: "norn.cli.commands"
    title: "CLI Command Catalog"
    filepath: "examples/other/cli-commands.csv"
    arguments: ["format-csv=table"]
    labels: ["cli", "commands", "csv"]
  },
  {
    name: "norn.impl.suggestions"
    title: "Implementation Suggestions"
    filepath: "examples/other/implementation.csv"
    arguments: ["format-csv=table"]
    labels: ["implementation", "csv", "suggestions"]
  },
  {
    name: "norn.impl.libraries"
    title: "Implementation Libraries"
    filepath: "examples/other/implementation-libraries.csv"
    arguments: ["format-csv=table"]
    labels: ["implementation", "libraries", "csv"]
  },
  {
    name: "norn.cue.config-key"
    title: "Key-oriented Config CUE Example"
    filepath: "examples/input/config-key.cue"
    labels: ["cue", "example", "config"]
  },
  {
    name: "norn.out.text-json"
    title: "Text Node Output Example (JSON)"
    filepath: "examples/output/text-node-output.json"
    labels: ["output", "json", "text"]
  },
  {
    name: "norn.out.validator-json"
    title: "Validator Node Output Example (JSON)"
    filepath: "examples/output/validator-node-output.json"
    labels: ["output", "json", "validator"]
  },
]

argumentRegistry: {
  version: "1"
  arguments: [
    {
      name: "format-csv"
      valueType: "enum"
      scopes: ["note"]
      allowedValues: ["table"]
      defaultValue: "table"
    },
  ]
}
