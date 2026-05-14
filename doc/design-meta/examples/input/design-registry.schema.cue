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

#CommandKind: string

#Command: {
  // Section names are intentionally open at schema level.
  // Implementations should check membership against supportedCommandSections.
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
  // Optional for future variants that may need explicit separators.
  delimiter?: "." | "_" | "-"
  from: "label-path-camelCase"
  allowCycles?: false | *false
  onGeneratedKeyCollision?: "error" | *"error"
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
  translationPolicy?: {
    requireAllSupportedLanguages: bool | *true
  }

  metaArgsValidation: #Command

  rootLabels:   [...string]
  nodesByLabel: [string]: #SchemaNode

  keyGeneration: #KeyGenerationPolicy
  generatedKeyExamplesMode: "illustrative" | "normative" | *"illustrative"

  generatedKeyExamples: {
    i18n:      [...string]
    text:      [...string]
    validator: [...string]
  }
}

#KeySchemaRegistry: [string]: #KeySchema

#TargetFormat: "arb.json" | "json" | "yaml" | "go" | "dart"
#CapabilityNodeKind: "i18n" | "text"
#CliCommand: "validate" | "generate" | "lint" | "list"
#ArtifactPatternToken: "{schemaRef}" | "{target}" | "{locale}" | "{domain}" | "{version}"

#GeneratorCapability: {
  keySchema:         string & !=""
  target:            #TargetFormat
  supportsNodeKinds: [...#CapabilityNodeKind]
  artifactPattern:   string & !=""
  notes?:            string
  scopeFilter?:      #Command
}

#LintPolicy: {
  unknownCommandSection: "error" | "warn" | "ignore"
}

#ArtifactPatternPolicy: {
  allowedTokens: [...#ArtifactPatternToken]
  requiredByTarget?: [#TargetFormat]: [...#ArtifactPatternToken]
}

#VersioningPolicy: {
  immutableSchemaRefs: bool | *true
  incompatibleChangesRequireNewRef: bool | *true
  deprecatedMeansLintWarn: bool | *true
  supersedesIsAdvisory: bool | *true
}

#DesignRegistrySpec: {
  supportedCliCommands: [...#CliCommand]
  keySchemaRegistry:     #KeySchemaRegistry
  generatorCapabilities: [...#GeneratorCapability]
  lintPolicy?: #LintPolicy
  artifactPatternPolicy?: #ArtifactPatternPolicy
  versioningPolicy?: #VersioningPolicy

  // Every generator capability must reference a known key schema id.
  _keySchemaRefChecks: [for c in generatorCapabilities {
    keySchemaRegistry[c.keySchema]
  }]
}
