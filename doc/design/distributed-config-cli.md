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
    "input-field": {
      metadata: {
        id: "input-field"
        version: "1.0.0"
      }
      supportedLanguages: ["en", "fr"]
      supportedCommandSections: ["validation", "monitoring", "transform"]
      translationPolicy: {
        requireAllSupportedLanguages: true
      }
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
          childLabels: ["textInput", "tags"]
        }
        textInput: {
          label: "textInput"
          kind: "branch"
          childLabels: ["label", "tooltip", "placeholder", "value"]
        }
        tags: {
          label: "tags"
          kind: "branch"
          childLabels: ["validation"]
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
        from: "label-path-camelCase"
        allowCycles: false
        onGeneratedKeyCollision: "error"
      }
    }
  }

  generatorCapabilities: [
    {
      keySchema: "input-field"
      target: "arb.json"
      supportsNodeKinds: ["i18n"]
      artifactPattern: "lib/l10n/app_<locale>.arb.json"
    },
    {
      keySchema: "input-field"
      target: "json"
      supportsNodeKinds: ["i18n", "text"]
      artifactPattern: "generated/config/<domain>.json"
    },
    {
      keySchema: "input-field"
      target: "cue"
      supportsNodeKinds: ["i18n", "text"]
      artifactPattern: "generated/config/<domain>.cue"
    },
  ]
}
```

#### CUE Design Registry Schema

```cue
package designregistry

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

#CommandKind: string

#Command: {
  // Section names are intentionally open at schema level.
  // Implementations should check membership against supportedCommandSections.
  args: [#CommandKind]: #CommandSpec
}

#NodeKind: "branch" | "i18n" | "text" | "validator"

#SchemaNode: {
  label:       string & !=""
  kind:        #NodeKind
  mandatory?:  bool
  childLabels: [...string]
}

#KeyGenerationPolicy: {
  // Optional for future variants that may need explicit separators.
  delimiter?: "." | "_" | "-"
  from: "label-path-camelCase"
  allowCycles?: false | *false
  onGeneratedKeyCollision?: "error" | *"error"
}

#KeySchemaMetadata: {
  id:      string & !=""
  // Informative schema version (not used as registry lookup key in v1).
  version?: string & !=""
}

#KeySchema: {
  metadata: #KeySchemaMetadata

  supportedLanguages:       [...string]
  supportedCommandSections: [...string]
  translationPolicy?: {
    requireAllSupportedLanguages: bool | *true
  }

  metaArgsValidation: #Command

  rootLabels:   [...string]
  nodesByLabel: [string]: #SchemaNode

  keyGeneration: #KeyGenerationPolicy
}

#KeySchemaRegistry: [string]: #KeySchema

#TargetFormat: "arb.json" | "json" | "yaml" | "go" | "dart" | "cue"
#CapabilityNodeKind: "i18n" | "text"

#GeneratorCapability: {
  keySchema:         string & !=""
  target:            #TargetFormat
  supportsNodeKinds: [...#CapabilityNodeKind]
  artifactPattern:   string & !=""
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
```

### 02 Key Schema And Validation

#### CUE Config Key Schema

```cue
package configkey

#ConfigKey: string & =~"^[a-z][A-Za-z0-9]*$"

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

#ValidationArgs: [string]: #CommandSpec

#ValidationCommand: {
  args: #ValidationArgs
}

#ValidationEntry: {
  key: #ConfigKey
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
  key: #ConfigKey
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
  key: #ConfigKey
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

Mandatory behavior: if a reachable schema node is marked `mandatory: true`,
the corresponding key entry must exist in the matching config section by node kind
(`i18nEntries`, `textEntries`, or `validations`), otherwise lint must raise an error.

### 03 Output Targets

#### Output Target Catalog

| artifact_pattern_example | id | in_registry_example | notes | primary_use | supports_node_kinds | target |
| --- | --- | --- | --- | --- | --- | --- |
| lib/l10n/app_<locale>.arb.json | out-001 | yes | Preferred i18n output for Dart/Flutter | Flutter i18n bundles | i18n | arb.json |
| generated/config/<domain>.json | out-002 | yes | Portable exchange format | General machine-readable config | i18n;text | json |
| generated/config/<domain>.yaml | out-003 | no | Planned support; keep parity with JSON where possible | Human-readable config export | text | yaml |
| internal/generated/<domain>_config.go | out-004 | no | Planned support; intended for Go services/libraries | Generated Go constants/types | text | go |
| lib/generated/<domain>_config.dart | out-005 | no | Planned support for non-i18n runtime config | Generated Dart config model | text | dart |
| generated/config/<domain>.cue | out-006 | yes | Useful for round-trip workflows and downstream CUE composition | Canonical CUE config output | i18n;text | cue |

### 04 CLI Commands

#### CLI Command Catalog

| command | error_policy | id | inputs | outputs | priority | purpose |
| --- | --- | --- | --- | --- | --- | --- |
| validate | hard-fail on schema errors | cmd-001 | design-registry.example.cue; config-key.cue | validation report | high | Validate registry and config inputs against CUE schemas |
| generate | hard-fail on generation errors | cmd-002 | resolved key schema + config entries | files by target format | high | Generate target artifacts from validated config |
| lint | non-zero exit on lint policy error | cmd-003 | design registry + config entries | lint diagnostics | high | Run policy checks across schema and config |
| list | soft-fail for empty lists | cmd-004 | registry input | stdout table/json | medium | List available registries, schemas, and targets |
| lint-schema | non-zero exit on violation | cmd-005 | key schema metadata + policies | lint diagnostics | high | Validate schema governance rules |
| lint-config | non-zero exit on violation | cmd-006 | config-key.cue + selected key schema | lint diagnostics | high | Validate config content quality and completeness |
| lint-sections | non-zero exit on unknown section | cmd-007 | validations[].commands[].args | lint diagnostics | high | Ensure command sections are allowed by supportedCommandSections |
| lint-keys | non-zero exit on mismatch | cmd-008 | nodesByLabel traversal + config entries | missing/unexpected key diagnostics | high | Check generated/expected keys versus config keys |
| lint-translations | non-zero exit when required language missing | cmd-009 | i18nEntries + supportedLanguages | translation coverage diagnostics | high | Ensure i18n entries satisfy supportedLanguages policy |
| lint-patterns | non-zero exit on invalid or missing required token | cmd-010 | generatorCapabilities[].artifactPattern | pattern diagnostics | medium | Validate artifactPattern placeholders against token policy |
| lint-graph | non-zero exit on cycle/collision | cmd-011 | nodesByLabel graph | graph diagnostics | high | Enforce no cycles and no generated key collisions |
| explain-key | soft-fail for unknown labels | cmd-012 | key schema + target label path | human-readable derivation trace | medium | Explain how a canonical key is derived from label-path-camelCase |
| dry-run-preview | non-zero exit on unresolved references | cmd-013 | config + generator capabilities | planned artifact list | medium | Preview outputs without writing files |
| version | always succeeds unless startup fails | cmd-014 | none | stdout version string/json | high | Return CLI version/build metadata |

### 05 Implementation Libraries

#### Implementation Libraries

| id | library | priority | purpose | why |
| --- | --- | --- | --- | --- |
| lib-001 | cobra | high | CLI command tree, flags, help and completion | De-facto standard Go CLI framework; clear command structure and ergonomics |
| lib-002 | snake-knot-picker | high | Schema-driven argv validation and parsing | Provides strict command/flag validation with stable error semantics |
| lib-003 | cuelang.org/go | high | Load, evaluate and validate CUE configs | Official CUE implementation for robust package/file handling and validation |

### 06 Implementation Suggestions

#### Implementation Suggestions

| area | id | priority | rationale | suggestion |
| --- | --- | --- | --- | --- |
| key-generation | impl-001 | high | Required to derive canonical keys from nodesByLabel DAG reliably | Implement deterministic traversal from rootLabels and childLabels with visited-set semantics |
| key-generation | impl-002 | high | Prevents drift between schema intent and generated keys | Use keyGeneration delimiter and from policy as the only key build mechanism |
| registry-resolution | impl-003 | high | Ensures capability runs with an explicit known schema | Resolve generatorCapabilities[*].keySchema against keySchemaRegistry before generation |
| meta-filtering | impl-004 | medium | Enables restrictive generation scope without hardcoding logic | Compile metaArgsValidation command and apply it to entry metaArgs when scopeFilter is present |
| config-validation | impl-005 | high | Fails early on malformed input shape | Validate config-key.cue sections against config-key.schema.cue before generation |
| key-coverage | impl-006 | high | Captures schema/data divergence early | Compare derived expected keys by kind with config-key.cue keys and report missing and unexpected keys |
| section-validation | impl-007 | medium | Keeps command namespaces controlled across versions | Validate command sections in validations against supportedCommandSections |
| target-routing | impl-008 | high | Prevents silent no-op or wrong output mapping | Select generators by target and supportsNodeKinds and keep unknown target as hard error |
| versioning | impl-009 | medium | Keeps registry references simple while leaving room for future migration strategy | Treat metadata.version as informative in v1 and avoid coupling lookup keys to version |
| diagnostics | impl-010 | medium | Improves automation and troubleshooting reliability | Emit diagnostics with stable IDs for schema, config, and generation stages |
| i18n-key-naming | impl-011 | high | Aligns with Flutter AppLocalizations naming conventions | Adopt camelCase for ARB keys and generated Dart API compatibility |
| i18n-key-naming | impl-012 | high | Improves searchability and avoids cross-feature key collisions | Use a stable feature prefix for every key (for example auth, checkout, settings) |
| i18n-key-naming | impl-013 | medium | Clarifies UI role without encoding fragile layout position | Prefer semantic intent suffixes such as Title Label Button Error DialogBody |
| i18n-key-naming | impl-014 | high | Reduces ambiguity and accidental reuse across unrelated contexts | Ban generic shared keys like saveButton in favor of feature-scoped keys |
| i18n-key-naming | impl-015 | high | Keys should survive wording and minor UI updates | Treat key renames as schema migrations and preserve stability across copy tweaks |
| i18n-key-naming | impl-016 | medium | Prevents artificial verbose keys and supports real-world complexity | Avoid strict positional naming templates; keep structure loose but consistent |
| i18n-key-metadata | impl-017 | medium | Keeps keys readable while preserving machine-usable structure | Store strict structural metadata outside the key string (ARB metadata or external config) |
| i18n-key-review | impl-018 | medium | Automates consistency and prevents regressions | Add a naming lint check that enforces prefix and generic-key denylist rules |
| cli-contract | impl-019 | high | Makes CLI capabilities explicit and versioned with the spec | Publish supported CLI commands in registry metadata and fail unknown command invocation |
| linting | impl-020 | high | Converts section governance from convention into an enforceable rule | Fail lint when validation command sections are outside supportedCommandSections |
| artifact-pattern | impl-021 | high | Prevents invalid output path templates and target-specific omissions | Validate artifactPattern placeholders against an allowed token set and required-by-target policy |
| graph-integrity | impl-022 | high | Ensures deterministic key generation from nodesByLabel | Raise errors for graph cycles and generated key collisions during key derivation |
| translation-coverage | impl-023 | high | Keeps localized output complete and consistent | Require each i18n entry to provide all supportedLanguages when translationPolicy demands it |
| spec-scope | impl-024 | high | Reduces user burden and avoids duplicating CLI-owned metadata | Keep design-registry CUE schema focused on required runtime inputs only |
| cli-owned-metadata | impl-025 | high | The CLI should know built-in command inventory and policy defaults | Keep supported CLI commands, lint behavior, and versioning semantics in CLI docs/code and command catalog CSV |
| schema-metadata | impl-026 | medium | Keeps schema refs minimal and avoids forcing informative fields | Use key schema metadata only for stable identity and version (id + version) |
| key-examples | impl-027 | medium | Examples clarify intent without constraining valid configs | Treat generated key examples as documentation artifacts rather than required user input fields |
| artifact-pattern-policy | impl-028 | medium | Placeholder governance is implementation policy, not user data | Document artifact pattern placeholder policy outside user config schema |
| mandatory-enforcement | impl-029 | high | Makes mandatory semantics explicit and enforceable in lint | For each reachable node with mandatory=true, require a matching key entry in the section mapped by node kind |
| config-input | impl-030 | high | Allows split configuration files while keeping load semantics deterministic | Support config input as either directory package (preferred) or single CUE file, with single-package enforcement for directory mode |

### 07 CUE Config Samples

#### Key-oriented Config CUE Example

```cue
i18nEntries: [
    {
        key: "fieldsTextInputLabel"
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
        key: "fieldsTextInputTooltip"
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
        key: "fieldsTextInputPlaceholder"
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
        key: "fieldsTextInputValue"
        description: "Default value for text input"
        kind: "text"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        value: "example value"
    },
]

validations: [
    {
        key: "fieldsTextInputValueValidation"
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
        key: "fieldsTagsValidation"
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

