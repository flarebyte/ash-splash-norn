# Distributed Config CLI Design

Design source for a Go CLI that manages distributed configurations across repositories.

## 01 Overview

Scope, goals, and output targets.

### 01 Intent

#### Distributed Configuration Management

This design models a Go CLI that reads CUE configuration sources and distributes generated artifacts to multiple codebases.

#### Generated Output Targets

Supported generated outputs are JSON, YAML, Go code, and Dart code.

## 02 Examples

Concrete examples collected under doc/design-meta/examples.

### 01 CLI Model

#### Application Composition Model

```ts
import { UUID, Version, Command } from "./common";
import { i18nLabelKey } from "./i18n";
import { RepositoryRef, ConfigSource, DistributionRule } from "./distributed-config-cli";

export type Application = {
  id: UUID;
  name: i18nLabelKey;
  title: i18nLabelKey;
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

### 02 CUE Config Samples

#### Key-oriented Config CUE Example

```cue
configurations = [
    {
        key: "checkoutPaymentErrorCardDeclined"
        description: "After checkout form"
        kind: "i18n"
        translations: {
            en: {
                text: "Checkout Payment"
                 context: {
                    feature: "checkout",
                    component: "payment",
                    type: "error",
                    surface: "snackbar"
                }
            }
            fr: {
                text: "Payment"
            }
     }
    },
    {
        key: 'checkoutPaymentErrorCardDeclinedColor'
        value: "#e4e4e4"
    }

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

