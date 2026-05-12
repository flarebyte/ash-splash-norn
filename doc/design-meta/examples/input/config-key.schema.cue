package configkey

#CommandFlagDef: {
  kind: "string" | "number" | "boolean" | "tuple"
  name: string & !=""
  schema: [...string] & ["schema", string, ...string]
  schemas?: [...[...string]]
}

#CommandSpec: {
  commandPath: [...string] & [string, ...string]
  adminOnly: bool
  flags: [...#CommandFlagDef] & [#CommandFlagDef, ...#CommandFlagDef]
}

#ValidationArgs: {
  validation?: #CommandSpec
  monitoring?: #CommandSpec
  transform?: #CommandSpec
}

#ValidationCommand: {
  args: #ValidationArgs
}

#ValidationEntry: {
  key: string & !=""
  metaArgs: ["meta", "--status", "draft" | "stable" | "experimental", "--app", "v1" | "v2"]
  commands: [...#ValidationCommand] & [#ValidationCommand, ...#ValidationCommand]
}

#I18nTranslation: {
  text: string
  context?: {
    feature?: string
    component?: string
    type?: string
    surface?: string
  }
}

#I18nEntry: {
  key: string & !=""
  description: string
  kind: "i18n"
  metaArgs: ["meta", "--status", "draft" | "stable" | "experimental", "--app", "v1" | "v2"]
  translations: {
    en: #I18nTranslation
    fr: #I18nTranslation
    [string]: #I18nTranslation
  }
}

#TextEntry: {
  key: string & !=""
  description: string
  kind: "text"
  metaArgs: ["meta", "--status", "draft" | "stable" | "experimental", "--app", "v1" | "v2"]
  value: string
}

i18nEntries: [...#I18nEntry]
textEntries: [...#TextEntry]
validations: [...#ValidationEntry]
