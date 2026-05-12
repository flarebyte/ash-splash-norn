import { Command } from "./common";

export type NodeKind =
  | "branch"
  | "i18n"
  | "text";

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
        "fields.textInput.placeholder",
        "fields.textInput.value",
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
      validation: [
        {
          args: {
            validation: {
              kind: "string",
              name: "textInput.value.validation",
              schema: ["--required", "--min-length", "1", "--max-length", "120"],
              schemas: [],
            },
          },
        },
      ],
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
