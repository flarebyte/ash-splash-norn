export type i18nLabelKey = string;
export type isoLanguage = string;

export type i18nLabels = {
  language: isoLanguage;
  labels: Record<i18nLabelKey, string>
};
