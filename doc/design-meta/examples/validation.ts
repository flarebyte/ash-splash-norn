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
          validation: ["--required", "--min-length", "1", "--max-length", "80"],
          monitoring: ["--warn-on-missing-translation"],
        },
      },
    ],
    "fields.textInput.minChars": [
      {
        args: {
          validation: ["--required", "--type", "number", "--min", "0", "--max", "500"],
        },
      },
      {
        args: {
          transform: ["--cast", "int"],
        },
      },
    ],
    "fields.textInput.required": [
      {
        args: {
          validation: ["--required", "--type", "boolean"],
        },
      },
    ],
  },
};
