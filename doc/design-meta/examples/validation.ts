import { Command, KeyStructKind } from "./common";

export type ValidationSchema = {
  commandByKeyStruct: Record<KeyStructKind, Command[]>;
};
