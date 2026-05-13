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
