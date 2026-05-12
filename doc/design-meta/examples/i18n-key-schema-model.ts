import { Command } from "./common";

export type NodeKind =
  | "branch"
  | "i18n"
  | "text";

export type TargetFormat = "arb.json" | "json" | "yaml" | "go" | "dart";

// A node can be reused by multiple parents, so the structure supports DAGs.
export type SchemaNode = {
  key: string;
  label: string;
  kind: NodeKind;
  mandatory?: boolean;
  helpKey?: string;
  childKeys: string[];
  validation?: Command[];
};

export type KeySchema = {
  rootKeys: string[];
  nodesByKey: Record<string, SchemaNode>;
  outputTargetsByKind?: Partial<Record<NodeKind, TargetFormat[]>>;
};

export const inputFieldSchema: KeySchema = {
  rootKeys: ["fields.textInput"],
  nodesByKey: {
    "fields.textInput": {
      key: "fields.textInput",
      label: "Text Input",
      kind: "branch",
      childKeys: [
        "fields.textInput.label",
        "fields.textInput.tooltip",
        "fields.textInput.required",
        "fields.textInput.minChars",
        "fields.textInput.maxChars",
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
    "fields.textInput.required": {
      key: "fields.textInput.required",
      label: "Required",
      kind: "boolean",
      childKeys: [],
    },
    "fields.textInput.minChars": {
      key: "fields.textInput.minChars",
      label: "Min Chars",
      kind: "number",
      childKeys: [],
    },
    "fields.textInput.maxChars": {
      key: "fields.textInput.maxChars",
      label: "Max Chars",
      kind: "number",
      childKeys: [],
    },
    "fields.textInput.help.common": {
      key: "fields.textInput.help.common",
      label: "Common Help",
      kind: "i18n",
      childKeys: [],
    },
  },
  outputTargetsByKind: {
    i18n: ["arb.json", "json"],
    string: ["json", "yaml", "go", "dart"],
    number: ["json", "yaml", "go", "dart"],
    boolean: ["json", "yaml", "go", "dart"],
    color: ["json", "yaml", "dart"],
  },
};
