import { UUID, Version } from "./common";
import { i18nLabelKey } from "./i18n";
export type Application = {
    id: UUID;
    name: i18nLabelKey;
    title: i18nLabelKey;
    version: Version;
}