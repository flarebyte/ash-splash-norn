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

### 01 App Model

#### Application Composition Model

```ts
import { UUID, Version, Command } from "./common";

export type Application = {
  id: UUID;
  name: string;
  title: string;
  version: Version;
};

export type CliApplication = Application & {
  command: Command;
  repositoryIds: UUID[];
  sourceIds: UUID[];
  ruleIds: UUID[];
};
```

### 02 Generator Capabilities

#### Generator Capabilities Matrix

```ts
import { Command } from "./common";
import { KeySchemaRef } from "./key-schema";

export type NodeKind = "i18n" | "text";
export type TargetFormat = "arb.json" | "json" | "yaml" | "go" | "dart";

export type GeneratorCapability = {
  keySchema: KeySchemaRef;
  target: TargetFormat;
  supportsNodeKinds: NodeKind[];
  artifactPattern: string;
  notes?: string;
  scopeFilter?: Command;
};

//command with a reduced scope
const exampleScopeFilter: Command = {
  args: {
    validation: {
      commandPath: ["meta"],
      adminOnly: false,
      flags: [
        {
          kind: "string",
          name: "status",
          schema: [
            "schema",
            "string",
            "--enum",
            "stable,experimental",
            "--required",
          ],
          schemas: [],
        },
        {
          kind: "string",
          name: "app",
          schema: ["schema", "string", "--enum", "v2", "--required"],
          schemas: [],
        },
      ],
    },
  },
};

export const generatorCapabilities: GeneratorCapability[] = [
  {
    keySchema: "input-field@1",
    target: "arb.json",
    supportsNodeKinds: ["i18n"],
    artifactPattern: "lib/l10n/app_<locale>.arb.json",
    notes: "Preferred Flutter/Dart localization target.",
    scopeFilter: exampleScopeFilter,
  },
  {
    keySchema: "input-field@1",
    target: "json",
    supportsNodeKinds: ["i18n", "text"],
    artifactPattern: "generated/config/<domain>.json",
    scopeFilter: exampleScopeFilter,
  },
  {
    keySchema: "input-field@1",
    target: "yaml",
    supportsNodeKinds: ["text"],
    artifactPattern: "generated/config/<domain>.yaml",
  },
  {
    keySchema: "input-field@1",
    target: "go",
    supportsNodeKinds: ["text"],
    artifactPattern: "internal/generated/<domain>_config.go",
    notes: "Generated structs/constants may evolve with compiler releases.",
  },

  {
    keySchema: "input-field@1",
    target: "dart",
    supportsNodeKinds: ["text"],
    artifactPattern: "lib/generated/<domain>_config.dart",
    notes: "Non-i18n data models for Flutter runtime configs.",
    scopeFilter: exampleScopeFilter,
  },
];
```

### 03 Schema And Validation Model

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
```

#### I18n Key Schema Hierarchy Model

```ts
import { Command } from "./common";

export type NodeKind = "branch" | "i18n" | "text" | "validator";

// A node can be reused by multiple parents, so the structure supports DAGs.
export type SchemaNode = {
  label: string;
  kind: NodeKind;
  mandatory?: boolean;
  childLabels: string[];
  maintenance?: {
    intent?: string;
    do?: string[];
    avoid?: string[];
    examples?: string[];
  };
};

export type KeyGenerationPolicy = {
  delimiter: ".";
  from: "label-path";
};

export type KeySchemaMetadata = {
  id: string;
  version: string;
  status: "draft" | "stable" | "deprecated";
  features: string[];
  compatibleTargets?: string[];
  supersedes?: string[];
  maintenance?: {
    intent?: string;
    do?: string[];
    avoid?: string[];
  };
};

export type KeySchema = {
  metadata: KeySchemaMetadata;
  supportedLanguages: string[];
  supportedCommandSections: string[];
  metaArgsValidation: Command;
  rootLabels: string[];
  nodesByLabel: Record<string, SchemaNode>;
  keyGeneration: KeyGenerationPolicy;
  // Examples of generated canonical keys expected in config-key.cue.
  generatedKeyExamples: {
    i18n: string[];
    text: string[];
    validator: string[];
  };
};

export type KeySchemaRef = string;
export type KeySchemaRegistry = Record<KeySchemaRef, KeySchema>;

export const inputFieldSchema: KeySchema = {
  metadata: {
    id: "input-field",
    version: "1.0.0",
    status: "stable",
    features: [
      "nodesByLabel-graph",
      "label-path-key-generation",
      "meta-args-command-spec",
      "snake-knot-picker-flag-schema",
    ],
    compatibleTargets: ["arb.json", "json", "yaml", "go", "dart"],
    maintenance: {
      intent: "Canonical key schema for input field i18n/text/validator entries.",
      do: ["Create a new version when changing key generation semantics."],
      avoid: ["Do not mutate existing semantics under the same version."],
    },
  },
  supportedLanguages: ["en", "fr"],
  supportedCommandSections: ["validation", "monitoring", "transform"],
  metaArgsValidation: {
    args: {
      validation: {
        commandPath: ["meta"],
        adminOnly: false,
        flags: [
          {
            kind: "string",
            name: "status",
            schema: [
              "schema",
              "string",
              "--enum",
              "draft,stable,experimental",
              "--required",
            ],
            schemas: [],
          },
          {
            kind: "string",
            name: "app",
            schema: ["schema", "string", "--enum", "v1,v2", "--required"],
            schemas: [],
          },
        ],
      },
    },
  },
  rootLabels: ["fields"],
  keyGeneration: {
    delimiter: ".",
    from: "label-path",
  },
  generatedKeyExamples: {
    i18n: [
      "fields.textInput.label",
      "fields.textInput.tooltip",
      "fields.textInput.placeholder",
    ],
    text: ["fields.textInput.value"],
    validator: ["fields.textInput.value.validation"],
  },
  nodesByLabel: {
    fields: {
      label: "fields",
      kind: "branch",
      childLabels: ["textInput"],
    },
    textInput: {
      label: "textInput",
      kind: "branch",
      childLabels: [
        "label",
        "tooltip",
        "placeholder",
        "value",
      ],
    },
    label: {
      label: "label",
      kind: "i18n",
      mandatory: true,
      childLabels: [],
      maintenance: {
        intent: "Primary user-facing label for the field.",
        do: [
          "Keep concise and action-oriented.",
          "Ensure all supported languages are provided in input CUE.",
        ],
        avoid: [
          "Do not embed validation rules in label text.",
          "Do not duplicate tooltip content.",
        ],
        examples: ["Checkout Payment", "Card Number"],
      },
    },
    tooltip: {
      label: "tooltip",
      kind: "i18n",
      childLabels: [],
      maintenance: {
        intent: "Contextual helper text shown on demand.",
        do: ["Prefer short explanatory guidance."],
        avoid: ["Avoid repeating the exact label text."],
        examples: ["Enter the value used for checkout."],
      },
    },
    placeholder: {
      label: "placeholder",
      kind: "i18n",
      childLabels: [],
    },
    value: {
      label: "value",
      kind: "text",
      childLabels: ["validation"],
    },
    validation: {
      label: "validation",
      kind: "validator",
      childLabels: [],
    },
  },
};

export const keySchemaRegistry: KeySchemaRegistry = {
  "input-field@1": inputFieldSchema,
};
```

#### Validation Source Of Truth

Validation commands are authored in `examples/input/config-key.cue` under the `validations` section.
This CUE input is the canonical source used to compile snake-knot-picker command documents.

### 04 CUE Config Samples

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

