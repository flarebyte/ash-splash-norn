package designregistry

import "list"

#FlagKind: "string" | "number" | "boolean" | "tuple"

#CommandFlagDef: {
  kind:   #FlagKind
  name:   string & !=""
  schema: [...string]
  schema: ["schema", ...string]
  schemas?: [...[...string]]
}

#CommandSpec: {
  commandPath: [...string]
  commandPath: [string, ...string]
  adminOnly:   bool
  flags:       [...#CommandFlagDef]
}

#CommandKind: "validation" | "monitoring" | "transform" | string

#Command: {
  args: [#CommandKind]: #CommandSpec
}

#NodeKind: "branch" | "i18n" | "text" | "validator"

#NodeMaintenance: {
  intent?:   string
  do?:       [...string]
  avoid?:    [...string]
  examples?: [...string]
}

#SchemaNode: {
  label:       string & !=""
  kind:        #NodeKind
  mandatory?:  bool
  childLabels: [...string]
  maintenance?: #NodeMaintenance
}

#KeyGenerationPolicy: {
  delimiter: "."
  from:      "label-path"
}

#KeySchemaMetadata: {
  id:      string & !=""
  version: string & !=""
  status:  "draft" | "stable" | "deprecated"

  features?:          [...string]
  compatibleTargets?: [...#TargetFormat]
  supersedes?:        [...string]

  maintenance?: {
    intent?: string
    do?:     [...string]
    avoid?:  [...string]
  }
}

#KeySchema: {
  metadata: #KeySchemaMetadata

  supportedLanguages:       [...string]
  supportedCommandSections: [...string]

  metaArgsValidation: #Command

  rootLabels:   [...string]
  nodesByLabel: [string]: #SchemaNode

  keyGeneration: #KeyGenerationPolicy

  generatedKeyExamples: {
    i18n:      [...string]
    text:      [...string]
    validator: [...string]
  }
}

#KeySchemaRegistry: [string]: #KeySchema

#TargetFormat: "arb.json" | "json" | "yaml" | "go" | "dart"
#CapabilityNodeKind: "i18n" | "text"

#GeneratorCapability: {
  keySchema:         string & !=""
  target:            #TargetFormat
  supportsNodeKinds: [...#CapabilityNodeKind]
  artifactPattern:   string & !=""
  notes?:            string
  scopeFilter?:      #Command
}

#DesignRegistrySpec: {
  keySchemaRegistry:     #KeySchemaRegistry
  generatorCapabilities: [...#GeneratorCapability]

  // Every generator capability must reference a known key schema id.
  _keySchemaRefChecks: [for c in generatorCapabilities {
    keySchemaRegistry[c.keySchema]
  }]
}

#ConfigKeyIndex: {
  i18nKeys:      [...string]
  textKeys:      [...string]
  validatorKeys: [...string]
  commandSections: [...string]
}

#DesignRegistryWithConfigSpec: {
  keySchemaRegistry:     #KeySchemaRegistry
  generatorCapabilities: [...#GeneratorCapability]
  _keySchemaRefChecks: [for c in generatorCapabilities {
    keySchemaRegistry[c.keySchema]
  }]

  selectedKeySchemaRef: string & != ""
  selectedKeySchema:    keySchemaRegistry[selectedKeySchemaRef]
  configKeyIndex:       #ConfigKeyIndex
  strictKeySet?: bool | *false

  // config keys must be allowed by the selected key schema examples
  _i18nSubsetChecks: [for k in configKeyIndex.i18nKeys if strictKeySet {
    list.Contains(selectedKeySchema.generatedKeyExamples.i18n, k) & true
  }]
  _textSubsetChecks: [for k in configKeyIndex.textKeys if strictKeySet {
    list.Contains(selectedKeySchema.generatedKeyExamples.text, k) & true
  }]
  _validatorSubsetChecks: [for k in configKeyIndex.validatorKeys if strictKeySet {
    list.Contains(selectedKeySchema.generatedKeyExamples.validator, k) & true
  }]

  // and every expected key should be present in config-key
  _i18nCoverageChecks: [for k in selectedKeySchema.generatedKeyExamples.i18n {
    list.Contains(configKeyIndex.i18nKeys, k) & true
  }]
  _textCoverageChecks: [for k in selectedKeySchema.generatedKeyExamples.text {
    list.Contains(configKeyIndex.textKeys, k) & true
  }]
  _validatorCoverageChecks: [for k in selectedKeySchema.generatedKeyExamples.validator {
    list.Contains(configKeyIndex.validatorKeys, k) & true
  }]

  // command sections used in config-key must be declared by selected key schema
  _commandSectionChecks: [for s in configKeyIndex.commandSections {
    list.Contains(selectedKeySchema.supportedCommandSections, s) & true
  }]
}
