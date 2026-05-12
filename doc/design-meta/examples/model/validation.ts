import { Command } from "./common";

export type ValidationPath = string;

// Each key path maps to command groups. Groups can represent phases like validation,
// transform, or export and are intentionally command-like for CLI execution.
export type ValidationSchema = {
  commandsByPath: Record<ValidationPath, Command[]>;
};

export const validationSchema: ValidationSchema = {
  commandsByPath: {
    "fields.textInput.label": [
      {
        args: {
          validation: {
            kind: "string",
            name: "textInput.label.validation",
            schema: ["--required", "--min-length", "1", "--max-length", "80"],
            schemas: [],
          },
          monitoring: {
            kind: "string",
            name: "textInput.label.monitoring",
            schema: ["--warn-on-missing-translation"],
            schemas: [],
          },
        },
      },
    ],
    "fields.textInput.minChars": [
      {
        args: {
          validation: {
            kind: "number",
            name: "textInput.minChars.validation",
            schema: ["--required", "--type", "number", "--min", "0", "--max", "500"],
            schemas: [],
          },
        },
      },
      {
        args: {
          transform: {
            kind: "number",
            name: "textInput.minChars.transform",
            schema: ["--cast", "int"],
            schemas: [],
          },
        },
      },
    ],
    "fields.textInput.required": [
      {
        args: {
          validation: {
            kind: "boolean",
            name: "textInput.required.validation",
            schema: ["--required", "--type", "boolean"],
            schemas: [],
          },
        },
      },
    ],
    "fields.tags": [
      {
        args: {
          // Repeated list item constraints.
          validation: {
            kind: "string",
            name: "tags.validation",
            schema: ["--type", "list", "--required"],
            schemas: [
              ["--item-type", "string", "--min-length", "2", "--max-length", "20"],
            ],
          },
        },
      },
    ],
    "fields.rangePairs": [
      {
        args: {
          // Repeated tuple constraints, e.g. [["0","10"],["20","30"]].
          validation: {
            kind: "tuple",
            name: "rangePairs.validation",
            schema: ["--type", "list"],
            schemas: [
              ["--tuple", "number", "number", "--tuple-rule", "item0<=item1"],
              ["--item-min", "0", "--item-max", "1000"],
            ],
          },
        },
      },
    ],
  },
};
