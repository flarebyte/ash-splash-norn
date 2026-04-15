import { UUID } from "./common";
import { i18nLabelKey } from "./i18n";

type StructKeyKind = 'branch' | 'arb-text' | 'string' | 'number' | 'boolean'

export type StructKeySchemaNode = {
  label: string;
  kind: StructKeyKind;
  childLabels: string[];
};

export type StructKeySchema = {
  rootLabels: string[];
  nodesByLabel: Record<string, StructKeySchemaNode>;
};


export const exampleSchema: StructKeySchema = {
  rootLabels: ['text'],
  nodesByLabel: {
    text: { label: 'Text', kind: 'branch', childLabels: ['label', 'description'] },
    label: { label: 'Label', kind: 'branch', childLabels: ['labelMinChar', 'labelMaxChar'] },
    labelMinChar: { label: 'MinChars', kind: 'number', childLabels: [] },
    labelMaxChar: { label: 'MaxChars', kind: 'number', childLabels: [] },
  },
};
