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
