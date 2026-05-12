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

### 01 CLI Model

#### Application Composition Model

```ts
import { UUID, Version, Command } from "./common";
import { RepositoryRef, ConfigSource, DistributionRule } from "./distributed-config-cli";

export type Application = {
  id: UUID;
  name: string;
  title: string;
  version: Version;
};

export type CliApplication = Application & {
  command: Command;
  repositories: RepositoryRef[];
  sources: ConfigSource[];
  rules: DistributionRule[];
};
```

#### TypeScript CLI Domain Model

```ts
import { UUID, Version } from "./common";

export type ConfigFormat = "cue";
export type OutputFormat = "json" | "yaml" | "go" | "dart";

export type RepositoryRef = {
  id: UUID;
  name: string;
  rootPath: string;
  branch?: string;
};

export type ConfigSource = {
  id: UUID;
  repositoryId: UUID;
  format: ConfigFormat;
  path: string;
};

export type OutputTarget = {
  format: OutputFormat;
  path: string;
  packageName?: string;
};

export type DistributionRule = {
  id: UUID;
  sourceId: UUID;
  targetRepositories: UUID[];
  outputs: OutputTarget[];
};

export type DistributedConfigCliApp = {
  id: UUID;
  name: string;
  version: Version;
  repositories: RepositoryRef[];
  sources: ConfigSource[];
  rules: DistributionRule[];
};
```

### 02 Generator Capabilities

#### Generator Capabilities Matrix

```ts
export type NodeKind = "i18n" | "text";
export type TargetFormat = "arb.json" | "json" | "yaml" | "go" | "dart";
export type SupportStatus = "stable" | "experimental";

export type GeneratorCapability = {
  target: TargetFormat;
  supportsNodeKinds: NodeKind[];
  artifactPattern: string;
  status: SupportStatus;
  notes?: string;
};

export const generatorCapabilities: GeneratorCapability[] = [
  {
    target: "arb.json",
    supportsNodeKinds: ["i18n"],
    artifactPattern: "lib/l10n/app_<locale>.arb.json",
    status: "stable",
    notes: "Preferred Flutter/Dart localization target.",
  },
  {
    target: "json",
    supportsNodeKinds: ["i18n", "text"],
    artifactPattern: "generated/config/<domain>.json",
    status: "stable",
  },
  {
    target: "yaml",
    supportsNodeKinds: ["text"],
    artifactPattern: "generated/config/<domain>.yaml",
    status: "stable",
  },
  {
    target: "go",
    supportsNodeKinds: ["text"],
    artifactPattern: "internal/generated/<domain>_config.go",
    status: "experimental",
    notes: "Generated structs/constants may evolve with compiler releases.",
  },
  {
    target: "dart",
    supportsNodeKinds: ["text"],
    artifactPattern: "lib/generated/<domain>_config.dart",
    status: "experimental",
    notes: "Non-i18n data models for Flutter runtime configs.",
  },
];
```

### 03 Schema And Validation Model

#### I18n Key Schema Hierarchy Model

```ts
import { Command } from "./common";

export type NodeKind =
  | "branch"
  | "i18n"
  | "text"
  | "validator";

// A node can be reused by multiple parents, so the structure supports DAGs.
export type SchemaNode = {
  key: string;
  label: string;
  kind: NodeKind;
  mandatory?: boolean;
  helpKey?: string;
  childKeys: string[];
};

export type KeySchema = {
  supportedLanguages: string[];
  supportedCommandSections: string[];
  metaArgsValidation: Command[];
  rootKeys: string[];
  nodesByKey: Record<string, SchemaNode>;
};

export const inputFieldSchema: KeySchema = {
  supportedLanguages: ["en", "fr"],
  supportedCommandSections: ["validation", "monitoring", "transform"],
  metaArgsValidation: [
    {
      args: {
        validation: {
          kind: "tuple",
          name: "metaArgs.validation",
          schema: ["--type", "list", "--required", "--min-items", "5"],
          schemas: [
            // Expected tokenized shape: ["meta", "--status", "<value>", "--app", "<value>"].
            ["--item-type", "string", "--min-length", "1"],
            ["--item0-eq", "meta"],
            ["--item1-eq", "--status"],
            ["--item3-eq", "--app"],
          ],
        },
      },
    },
  ],
  rootKeys: ["fields.textInput"],
  nodesByKey: {
    "fields.textInput": {
      key: "fields.textInput",
      label: "Text Input",
      kind: "branch",
      childKeys: [
        "fields.textInput.label",
        "fields.textInput.tooltip",
        "fields.textInput.placeholder",
        "fields.textInput.value",
        "fields.textInput.value.validation",
      ],
    },
    "fields.textInput.label": {
      key: "fields.textInput.label",
      label: "Label",
      kind: "i18n",
      mandatory: true,
      helpKey: "fields.textInput.help.common",
      childKeys: [],
    },
    "fields.textInput.tooltip": {
      key: "fields.textInput.tooltip",
      label: "Tooltip",
      kind: "i18n",
      helpKey: "fields.textInput.help.common",
      childKeys: [],
    },
    "fields.textInput.placeholder": {
      key: "fields.textInput.placeholder",
      label: "Placeholder",
      kind: "i18n",
      childKeys: [],
    },
    "fields.textInput.value": {
      key: "fields.textInput.value",
      label: "Value",
      kind: "text",
      childKeys: [],
    },
    "fields.textInput.value.validation": {
      key: "fields.textInput.value.validation",
      label: "Value Validation",
      kind: "validator",
      childKeys: [],
    },
    "fields.textInput.help.common": {
      key: "fields.textInput.help.common",
      label: "Common Help",
      kind: "i18n",
      childKeys: [],
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
i18nEntries = [
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

textEntries = [
    {
        key: "fields.textInput.value"
        description: "Default value for text input"
        kind: "text"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        value: "example value"
    },
]

validations = [
    {
        key: "fields.textInput.value.validation"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        commands: [
            {
                args: {
                    validation: {
                        kind: "string"
                        name: "textInput.value.validation"
                        schema: ["--required", "--min-length", "1", "--max-length", "120"]
                        schemas: []
                    }
                }
            },
            {
                args: {
                    monitoring: {
                        kind: "string"
                        name: "textInput.value.monitoring"
                        schema: ["--warn-on-missing-translation"]
                        schemas: []
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
                    kind: "string"
                    name: "tags.validation"
                    schema: ["--type", "list", "--required"]
                    schemas: [
                        ["--item-type", "string", "--min-length", "2", "--max-length", "20"],
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

