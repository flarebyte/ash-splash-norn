export type UUID = string;
export type Version = string;
export type KeyStructKind = string //note:text:comment
export type CommandKind = string;//validation,monitoring


export type Command = {
    args: Record<CommandKind, string[]>;
}