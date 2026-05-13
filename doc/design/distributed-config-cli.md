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

export type NodeKind = "i18n" | "text";
export type TargetFormat = "arb.json" | "json" | "yaml" | "go" | "dart";

export type GeneratorCapability = {
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
    target: "arb.json",
    supportsNodeKinds: ["i18n"],
    artifactPattern: "lib/l10n/app_<locale>.arb.json",
    notes: "Preferred Flutter/Dart localization target.",
    scopeFilter: exampleScopeFilter,
  },
  {
    target: "json",
    supportsNodeKinds: ["i18n", "text"],
    artifactPattern: "generated/config/<domain>.json",
    scopeFilter: exampleScopeFilter,
  },
  {
    target: "yaml",
    supportsNodeKinds: ["text"],
    artifactPattern: "generated/config/<domain>.yaml",
  },
  {
    target: "go",
    supportsNodeKinds: ["text"],
    artifactPattern: "internal/generated/<domain>_config.go",
    notes: "Generated structs/constants may evolve with compiler releases.",
  },

  {
    target: "dart",
    supportsNodeKinds: ["text"],
    artifactPattern: "lib/generated/<domain>_config.dart",
    notes: "Non-i18n data models for Flutter runtime configs.",
    scopeFilter: exampleScopeFilter,
  },
];
```

### 03 Schema And Validation Model

#### I18n Key Schema Hierarchy Model

```ts
import { Command } from "./common";

export type NodeKind = "branch" | "i18n" | "text" | "validator";

// A node can be reused by multiple parents, so the structure supports DAGs.
export type SchemaNode = {
  label: string;
  kind: NodeKind;
  mandatory?: boolean;
  helpLabel?: string;
  childLabels: string[];
};

export type KeyGenerationPolicy = {
  delimiter: ".";
  from: "label-path";
};

export type KeySchema = {
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

export const inputFieldSchema: KeySchema = {
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
      "fields.textInput.helpCommon",
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
      helpLabel: "helpCommon",
      childLabels: [],
    },
    tooltip: {
      label: "tooltip",
      kind: "i18n",
      helpLabel: "helpCommon",
      childLabels: [],
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
    helpCommon: {
      label: "helpCommon",
      kind: "i18n",
      childLabels: [],
    },
  },
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

#### Distributed Config CUE Example

```cue
app: {
  id: "c2f6c976-f498-4adb-8eb4-53136cb06a50"
  name: "nornctl"
  version: "0.1.0"
}

repositories: [
  {
    id: "92881602-dc9e-4389-befe-9a8f0519e7bf"
    name: "checkout-web"
    rootPath: "github.com/acme/checkout-web"
    branch: "main"
  },
  {
    id: "8b17d934-0d3b-40bd-8f6e-e9827b2324f4"
    name: "billing-api"
    rootPath: "github.com/acme/billing-api"
    branch: "main"
  },
]

sources: [
  {
    id: "a1f23578-ae3d-4ac1-8d6c-c7538f8da6bb"
    repositoryId: "92881602-dc9e-4389-befe-9a8f0519e7bf"
    format: "cue"
    path: "config/payment.cue"
  },
]

rules: [
  {
    id: "ee3d1516-159f-4e4e-8fe7-f0da63754b63"
    sourceId: "a1f23578-ae3d-4ac1-8d6c-c7538f8da6bb"
    targetRepositories: [
      "92881602-dc9e-4389-befe-9a8f0519e7bf",
      "8b17d934-0d3b-40bd-8f6e-e9827b2324f4",
    ]
    outputs: [
      {
        format: "json"
        path: "generated/config/payment.json"
      },
      {
        format: "yaml"
        path: "generated/config/payment.yaml"
      },
      {
        format: "go"
        path: "internal/generated/payment_config.go"
        packageName: "generated"
      },
      {
        format: "dart"
        path: "lib/generated/payment_config.dart"
      },
    ]
  },
]
```

