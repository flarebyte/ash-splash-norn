import { UUID } from "./common";
import { i18nLabelKey } from "./i18n";

type StructKeyKindExamples = 'branch' | 'i18n' | 'string' | 'number' | 'boolean' | 'color'

export type StructKeySchemaNode = {
  label: string;
  kind: StructKeyKindExamples;
  mandatory?: boolean;
  childLabels: string[];
};

export type StructKeySchema = {
  rootLabels: string[];
  nodesByLabel: Record<string, StructKeySchemaNode>;
};


export const exampleSchema: StructKeySchema = {
  rootLabels: ['text'],
  nodesByLabel: {
    textInput: { label: 'TextInput', kind: 'branch', childLabels: ['textInputLabel', 'labelMinChar', 'labelMaxChar'] },
    textInputLabel: { label: 'Label', kind: 'i18n', mandatory: true, childLabels: [] },
    labelMinChar: { label: 'MinChars', kind: 'number', childLabels: [] },
    labelMaxChar: { label: 'MaxChars', kind: 'number', childLabels: [] },
  },
};
