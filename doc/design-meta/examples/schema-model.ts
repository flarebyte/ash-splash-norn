import { UUID } from "./common";

export type SchemaValueType = 'id' | '_' | '~';

export type IdAlphabet =
  | 'lower-alpha'
  | 'alpha'
  | 'digit'
  | 'lower-alnum'
  | 'alnum'
  | 'lower-hex'
  | 'upper-hex'
  | 'hex';

export type KeySchemaConfig = {
  maxDepth: number;
  minIdChars: number;
  maxIdChars: number;
  idAlphabet: IdAlphabet;
  extraIdChars: string[];
};

export type KeySchemaNode = {
  label: string;
  valueTypes: SchemaValueType[];
  childLabels: string[];
};

export type KeySchema = {
  config: KeySchemaConfig;
  rootLabels: string[];
  anchorLabels: string[];
  nodesByLabel: Record<string, KeySchemaNode>;
};

export type KeySchemaStore = {
  schema: Record<UUID, KeySchema>
}

export const exampleSchema: KeySchema = {
  config: {
    maxDepth: 8,
    minIdChars: 1,
    maxIdChars: 64,
    idAlphabet: 'lower-hex',
    extraIdChars: ['-'],
  },
  rootLabels: ['tenant', 'department', 'region'],
  anchorLabels: ['dashboard', 'profile'],
  nodesByLabel: {
    tenant: { label: 'tenant', valueTypes: ['id'], childLabels: ['group', 'department', 'region', 'dashboard', 'profile'] },
    group: { label: 'group', valueTypes: ['id'], childLabels: ['dashboard', 'profile'] },
    department: { label: 'department', valueTypes: ['id'], childLabels: ['team', 'profile'] },
    team: { label: 'team', valueTypes: ['id'], childLabels: ['dashboard', 'profile'] },
    region: { label: 'region', valueTypes: ['id'], childLabels: ['dashboard', 'profile'] },
    dashboard: { label: 'dashboard', valueTypes: ['id'], childLabels: ['note', 'language', 'thumbnail', 'like', 'user'] },
    note: { label: 'note', valueTypes: ['id'], childLabels: ['text', 'language', 'thumbnail', 'like'] },
    like: { label: 'like', valueTypes: ['_'], childLabels: ['count', 'user', 'member', 'subscriber'] },
    text: { label: 'text', valueTypes: ['_'], childLabels: [] },
    count: { label: 'count', valueTypes: ['_'], childLabels: [] },
    language: { label: 'language', valueTypes: ['_'], childLabels: [] },
    thumbnail: { label: 'thumbnail', valueTypes: ['_'], childLabels: [] },
    user: { label: 'user', valueTypes: ['~', '_'], childLabels: [] },
    member: { label: 'member', valueTypes: ['id', '_'], childLabels: [] },
    subscriber: { label: 'subscriber', valueTypes: ['id', '_'], childLabels: [] },
    profile: { label: 'profile', valueTypes: ['id'], childLabels: [] },
  },
};

