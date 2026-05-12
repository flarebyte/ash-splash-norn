export type UUID = string;
export type Version = string;
export type KeyStructKind = string; //note:text:comment
export type CommandKind = string; //validation,monitoring

export type FlagSpec = {
  kind: "string" | "number" | "boolean" | "tuple";
  name: string;
  schema: string[];
  schemas: string[][];
};

export type CommandSpec = {
  commandPath: string[];
  adminOnly: boolean;
  flags: FlagSpec[];
};

export type Command = {
  args: Record<CommandKind, CommandSpecs>;
};
