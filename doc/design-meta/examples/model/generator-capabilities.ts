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
    keySchema: "input-field",
    target: "arb.json",
    supportsNodeKinds: ["i18n"],
    artifactPattern: "lib/l10n/app_<locale>.arb.json",
    notes: "Preferred Flutter/Dart localization target.",
    scopeFilter: exampleScopeFilter,
  },
  {
    keySchema: "input-field",
    target: "json",
    supportsNodeKinds: ["i18n", "text"],
    artifactPattern: "generated/config/<domain>.json",
    scopeFilter: exampleScopeFilter,
  },
  {
    keySchema: "input-field",
    target: "yaml",
    supportsNodeKinds: ["text"],
    artifactPattern: "generated/config/<domain>.yaml",
  },
  {
    keySchema: "input-field",
    target: "go",
    supportsNodeKinds: ["text"],
    artifactPattern: "internal/generated/<domain>_config.go",
    notes: "Generated structs/constants may evolve with compiler releases.",
  },

  {
    keySchema: "input-field",
    target: "dart",
    supportsNodeKinds: ["text"],
    artifactPattern: "lib/generated/<domain>_config.dart",
    notes: "Non-i18n data models for Flutter runtime configs.",
    scopeFilter: exampleScopeFilter,
  },
];
