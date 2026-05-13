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
