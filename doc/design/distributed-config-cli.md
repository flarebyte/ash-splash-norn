# Distributed Config CLI Design

Design source for a Go CLI that manages distributed configurations across repositories.

## 01 Overview

Scope, goals, and output targets.

### 01 Intent

#### Distributed Configuration Management

This design models a Go CLI that reads CUE configuration sources and distributes generated artifacts to multiple codebases.

#### Generated Output Targets

Supported generated outputs are JSON, YAML, Go code, and Dart code.
For i18n keys in Flutter/Dart contexts, the preferred target is `*.arb.json`.
Target compatibility is determined by the CLI capabilities, not by extra user-provided per-kind maps.

## 02 Examples

Concrete examples collected under doc/design-meta/examples.

### 01 Registry Schema

#### CUE Design Registry Example

```cue
package designregistry

designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field@1": {
      metadata: {
        id: "input-field"
        version: "1.0.0"
        status: "stable"
        features: [
          "nodesByLabel-graph",
          "label-path-key-generation",
          "meta-args-command-spec",
          "snake-knot-picker-flag-schema",
        ]
        compatibleTargets: ["arb.json", "json", "yaml", "go", "dart"]
      }
      supportedLanguages: ["en", "fr"]
      supportedCommandSections: ["validation", "monitoring", "transform"]
      metaArgsValidation: {
        args: {
          validation: {
            commandPath: ["meta"]
            adminOnly: false
            flags: [
              {
                kind: "string"
                name: "status"
                schema: ["schema", "string", "--enum", "draft,stable,experimental", "--required"]
                schemas: []
              },
              {
                kind: "string"
                name: "app"
                schema: ["schema", "string", "--enum", "v1,v2", "--required"]
                schemas: []
              },
            ]
          }
        }
      }
      rootLabels: ["fields"]
      nodesByLabel: {
        fields: {
          label: "fields"
          kind: "branch"
          childLabels: ["textInput"]
        }
        textInput: {
          label: "textInput"
          kind: "branch"
          childLabels: ["label", "tooltip", "placeholder", "value"]
        }
        label: {
          label: "label"
          kind: "i18n"
          mandatory: true
          childLabels: []
        }
        tooltip: {
          label: "tooltip"
          kind: "i18n"
          childLabels: []
        }
        placeholder: {
          label: "placeholder"
          kind: "i18n"
          childLabels: []
        }
        value: {
          label: "value"
          kind: "text"
          childLabels: ["validation"]
        }
        validation: {
          label: "validation"
          kind: "validator"
          childLabels: []
        }
      }
      keyGeneration: {
        delimiter: "."
        from: "label-path"
      }
      generatedKeyExamples: {
        i18n: [
          "fields.textInput.label",
          "fields.textInput.tooltip",
          "fields.textInput.placeholder",
        ]
        text: ["fields.textInput.value"]
        validator: ["fields.textInput.value.validation"]
      }
    }
  }

  generatorCapabilities: [
    {
      keySchema: "input-field@1"
      target: "arb.json"
      supportsNodeKinds: ["i18n"]
      artifactPattern: "lib/l10n/app_<locale>.arb.json"
      notes: "Preferred Flutter/Dart localization target."
    },
    {
      keySchema: "input-field@1"
      target: "json"
      supportsNodeKinds: ["i18n", "text"]
      artifactPattern: "generated/config/<domain>.json"
    },
  ]
}

designRegistryWithConfig: #DesignRegistryWithConfigSpec & {
  keySchemaRegistry: designRegistry.keySchemaRegistry
  generatorCapabilities: designRegistry.generatorCapabilities
  selectedKeySchemaRef: "input-field@1"
  strictKeySet: false
  configKeyIndex: {
    i18nKeys: [
      "fields.textInput.label",
      "fields.textInput.tooltip",
      "fields.textInput.placeholder",
    ]
    textKeys: ["fields.textInput.value"]
    validatorKeys: [
      "fields.textInput.value.validation",
      "fields.tags.validation",
    ]
    commandSections: ["validation", "monitoring"]
  }
}
```

#### CUE Design Registry Schema

```cue
package designregistry

import "list"

#FlagKind: "string" | "number" | "boolean" | "tuple"

#CommandFlagDef: {
  kind:   #FlagKind
  name:   string & !=""
  schema: [...string]
  schema: ["schema", ...string]
  schemas?: [...[...string]]
}

#CommandSpec: {
  commandPath: [...string]
  commandPath: [string, ...string]
  adminOnly:   bool
  flags:       [...#CommandFlagDef]
}

#CommandKind: "validation" | "monitoring" | "transform" | string

#Command: {
  args: [#CommandKind]: #CommandSpec
}

#NodeKind: "branch" | "i18n" | "text" | "validator"

#NodeMaintenance: {
  intent?:   string
  do?:       [...string]
  avoid?:    [...string]
  examples?: [...string]
}

#SchemaNode: {
  label:       string & !=""
  kind:        #NodeKind
  mandatory?:  bool
  childLabels: [...string]
  maintenance?: #NodeMaintenance
}

#KeyGenerationPolicy: {
  delimiter: "."
  from:      "label-path"
}

#KeySchemaMetadata: {
  id:      string & !=""
  version: string & !=""
  status:  "draft" | "stable" | "deprecated"

  features?:          [...string]
  compatibleTargets?: [...#TargetFormat]
  supersedes?:        [...string]

  maintenance?: {
    intent?: string
    do?:     [...string]
    avoid?:  [...string]
  }
}

#KeySchema: {
  metadata: #KeySchemaMetadata

  supportedLanguages:       [...string]
  supportedCommandSections: [...string]

  metaArgsValidation: #Command

  rootLabels:   [...string]
  nodesByLabel: [string]: #SchemaNode

  keyGeneration: #KeyGenerationPolicy

  generatedKeyExamples: {
    i18n:      [...string]
    text:      [...string]
    validator: [...string]
  }
}

#KeySchemaRegistry: [string]: #KeySchema

#TargetFormat: "arb.json" | "json" | "yaml" | "go" | "dart"
#CapabilityNodeKind: "i18n" | "text"

#GeneratorCapability: {
  keySchema:         string & !=""
  target:            #TargetFormat
  supportsNodeKinds: [...#CapabilityNodeKind]
  artifactPattern:   string & !=""
  notes?:            string
  scopeFilter?:      #Command
}

#DesignRegistrySpec: {
  keySchemaRegistry:     #KeySchemaRegistry
  generatorCapabilities: [...#GeneratorCapability]

  // Every generator capability must reference a known key schema id.
  _keySchemaRefChecks: [for c in generatorCapabilities {
    keySchemaRegistry[c.keySchema]
  }]
}

#ConfigKeyIndex: {
  i18nKeys:      [...string]
  textKeys:      [...string]
  validatorKeys: [...string]
  commandSections: [...string]
}

#DesignRegistryWithConfigSpec: {
  keySchemaRegistry:     #KeySchemaRegistry
  generatorCapabilities: [...#GeneratorCapability]
  _keySchemaRefChecks: [for c in generatorCapabilities {
    keySchemaRegistry[c.keySchema]
  }]

  selectedKeySchemaRef: string & != ""
  selectedKeySchema:    keySchemaRegistry[selectedKeySchemaRef]
  configKeyIndex:       #ConfigKeyIndex
  strictKeySet?: bool | *false

  // config keys must be allowed by the selected key schema examples
  _i18nSubsetChecks: [for k in configKeyIndex.i18nKeys if strictKeySet {
    list.Contains(selectedKeySchema.generatedKeyExamples.i18n, k) & true
  }]
  _textSubsetChecks: [for k in configKeyIndex.textKeys if strictKeySet {
    list.Contains(selectedKeySchema.generatedKeyExamples.text, k) & true
  }]
  _validatorSubsetChecks: [for k in configKeyIndex.validatorKeys if strictKeySet {
    list.Contains(selectedKeySchema.generatedKeyExamples.validator, k) & true
  }]

  // and every expected key should be present in config-key
  _i18nCoverageChecks: [for k in selectedKeySchema.generatedKeyExamples.i18n {
    list.Contains(configKeyIndex.i18nKeys, k) & true
  }]
  _textCoverageChecks: [for k in selectedKeySchema.generatedKeyExamples.text {
    list.Contains(configKeyIndex.textKeys, k) & true
  }]
  _validatorCoverageChecks: [for k in selectedKeySchema.generatedKeyExamples.validator {
    list.Contains(configKeyIndex.validatorKeys, k) & true
  }]

  // command sections used in config-key must be declared by selected key schema
  _commandSectionChecks: [for s in configKeyIndex.commandSections {
    list.Contains(selectedKeySchema.supportedCommandSections, s) & true
  }]
}
```

### 02 Key Schema And Validation

#### CUE Config Key Schema

```cue
package configkey

#CommandFlagDef: {
  kind: "string" | "number" | "boolean" | "tuple"
  name: string & !=""
  schema: [...string] & ["schema", string, ...string]
  schemas?: [...[...string]]
}

#CommandSpec: {
  commandPath: [...string] & [string, ...string]
  adminOnly: bool
  flags: [...#CommandFlagDef] & [#CommandFlagDef, ...#CommandFlagDef]
}

#ValidationArgs: {
  validation?: #CommandSpec
  monitoring?: #CommandSpec
  transform?: #CommandSpec
}

#ValidationCommand: {
  args: #ValidationArgs
}

#ValidationEntry: {
  key: string & !=""
  metaArgs: ["meta", "--status", "draft" | "stable" | "experimental", "--app", "v1" | "v2"]
  commands: [...#ValidationCommand] & [#ValidationCommand, ...#ValidationCommand]
}

#I18nTranslation: {
  text: string
  context?: {
    feature?: string
    component?: string
    type?: string
    surface?: string
  }
}

#I18nEntry: {
  key: string & !=""
  description: string
  kind: "i18n"
  metaArgs: ["meta", "--status", "draft" | "stable" | "experimental", "--app", "v1" | "v2"]
  translations: {
    en: #I18nTranslation
    fr: #I18nTranslation
    [string]: #I18nTranslation
  }
}

#TextEntry: {
  key: string & !=""
  description: string
  kind: "text"
  metaArgs: ["meta", "--status", "draft" | "stable" | "experimental", "--app", "v1" | "v2"]
  value: string
}

i18nEntries: [...#I18nEntry]
textEntries: [...#TextEntry]
validations: [...#ValidationEntry]
```

#### Validation Source Of Truth

Validation commands are authored in `examples/input/config-key.cue` under the `validations` section.
This CUE input is the canonical source used to compile snake-knot-picker command documents.

### 03 CUE Config Samples

#### Key-oriented Config CUE Example

```cue
i18nEntries: [
    {
        key: "fields.textInput.label"
        description: "After checkout form"
        kind: "i18n"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        translations: {
            en: {
                text: "Checkout Payment"
                context: {
                    feature: "checkout"
                    component: "payment"
                    type: "error"
                    surface: "snackbar"
                }
            }
            fr: {
                text: "Payment"
            }
        }
    },
    {
        key: "fields.textInput.tooltip"
        description: "Tooltip for text input"
        kind: "i18n"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        translations: {
            en: {
                text: "Enter the value used for checkout."
            }
            fr: {
                text: "Saisissez la valeur utilisee pour le paiement."
            }
        }
    },
    {
        key: "fields.textInput.placeholder"
        description: "Placeholder for text input"
        kind: "i18n"
        metaArgs: ["meta", "--status", "stable", "--app", "v1"]
        translations: {
            en: {
                text: "Type here"
            }
            fr: {
                text: "Saisissez ici"
            }
        }
    },
]

textEntries: [
    {
        key: "fields.textInput.value"
        description: "Default value for text input"
        kind: "text"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        value: "example value"
    },
]

validations: [
    {
        key: "fields.textInput.value.validation"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        commands: [
            {
                args: {
                    validation: {
                        commandPath: ["validate", "text-input", "value"]
                        adminOnly: false
                        flags: [
                            {
                                kind: "string"
                                name: "value"
                                schema: ["schema", "string", "--required", "--min-length", "1", "--max-length", "120"]
                                schemas: []
                            },
                        ]
                    }
                }
            },
            {
                args: {
                    monitoring: {
                        commandPath: ["monitor", "text-input", "value"]
                        adminOnly: false
                        flags: [
                            {
                                kind: "string"
                                name: "warn-missing-translation"
                                schema: ["schema", "string", "--required"]
                                schemas: []
                            },
                        ]
                    }
                }
            },
        ]
    },
    {
        key: "fields.tags.validation"
        metaArgs: ["meta", "--status", "experimental", "--app", "v2"]
        commands: [{
            args: {
                validation: {
                    commandPath: ["validate", "tags"]
                    adminOnly: false
                    flags: [
                        {
                            kind: "string"
                            name: "tags"
                            schema: ["schema", "string", "--alphabetic"]
                            schemas: [
                                ["schema", "repeatable", "--min-length", "1", "--max-length", "20"],
                            ]
                        },
                    ]
                }
            }
        }]
    },
]
```

