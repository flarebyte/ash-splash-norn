package designregistry

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
  from: "label-path-camelCase"
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
