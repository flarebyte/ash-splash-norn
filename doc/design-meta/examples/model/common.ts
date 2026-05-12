export type UUID = string;
export type Version = string;
export type KeyStructKind = string //note:text:comment
export type CommandKind = string;//validation,monitoring

export type FlagSpec = {
    name: string;
    valueKind?: 'string' | 'number' | 'boolean';
    required?: boolean;
    values?: string[];
    description?: string;
}

export type ConstraintSpec = {
    kind: 'string' | 'number' | 'boolean' | 'tuple' | 'list';
    name: string;
    flags: FlagSpec[];
}

export type Command = {
    args: Record<CommandKind, ConstraintSpec>;
}
