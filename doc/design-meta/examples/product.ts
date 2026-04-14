import { UUID, Version } from "./common";
import { i18nLabelKey } from "./i18n";

export type ProductTag= 'spreadsheet' | 'form'

export type Product = {
    id: UUID;
    tags: ProductTag[];
    name: i18nLabelKey;
    title: i18nLabelKey;
    version: Version;
    schemaId: UUID;
}