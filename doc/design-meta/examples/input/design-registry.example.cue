package designregistry

designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field@1": {
      metadata: {
        id: "input-field"
        version: "1.0.0"
        status: "stable"
        features: [
          "nodesByLabel-graph",
          "label-path-key-generation",
          "meta-args-command-spec",
          "snake-knot-picker-flag-schema",
        ]
        compatibleTargets: ["arb.json", "json", "yaml", "go", "dart"]
      }
      supportedLanguages: ["en", "fr"]
      supportedCommandSections: ["validation", "monitoring", "transform"]
      metaArgsValidation: {
        args: {
          validation: {
            commandPath: ["meta"]
            adminOnly: false
            flags: [
              {
                kind: "string"
                name: "status"
                schema: ["schema", "string", "--enum", "draft,stable,experimental", "--required"]
                schemas: []
              },
              {
                kind: "string"
                name: "app"
                schema: ["schema", "string", "--enum", "v1,v2", "--required"]
                schemas: []
              },
            ]
          }
        }
      }
      rootLabels: ["fields"]
      nodesByLabel: {
        fields: {
          label: "fields"
          kind: "branch"
          childLabels: ["textInput"]
        }
        textInput: {
          label: "textInput"
          kind: "branch"
          childLabels: ["label", "tooltip", "placeholder", "value"]
        }
        label: {
          label: "label"
          kind: "i18n"
          mandatory: true
          childLabels: []
        }
        tooltip: {
          label: "tooltip"
          kind: "i18n"
          childLabels: []
        }
        placeholder: {
          label: "placeholder"
          kind: "i18n"
          childLabels: []
        }
        value: {
          label: "value"
          kind: "text"
          childLabels: ["validation"]
        }
        validation: {
          label: "validation"
          kind: "validator"
          childLabels: []
        }
      }
      keyGeneration: {
        delimiter: "."
        from: "label-path"
      }
      generatedKeyExamples: {
        i18n: [
          "fields.textInput.label",
          "fields.textInput.tooltip",
          "fields.textInput.placeholder",
        ]
        text: ["fields.textInput.value"]
        validator: ["fields.textInput.value.validation"]
      }
    }
  }

  generatorCapabilities: [
    {
      keySchema: "input-field@1"
      target: "arb.json"
      supportsNodeKinds: ["i18n"]
      artifactPattern: "lib/l10n/app_<locale>.arb.json"
      notes: "Preferred Flutter/Dart localization target."
    },
    {
      keySchema: "input-field@1"
      target: "json"
      supportsNodeKinds: ["i18n", "text"]
      artifactPattern: "generated/config/<domain>.json"
    },
  ]
}

designRegistryWithConfig: #DesignRegistryWithConfigSpec & {
  keySchemaRegistry: designRegistry.keySchemaRegistry
  generatorCapabilities: designRegistry.generatorCapabilities
  selectedKeySchemaRef: "input-field@1"
  strictKeySet: false
  configKeyIndex: {
    i18nKeys: [
      "fields.textInput.label",
      "fields.textInput.tooltip",
      "fields.textInput.placeholder",
    ]
    textKeys: ["fields.textInput.value"]
    validatorKeys: [
      "fields.textInput.value.validation",
      "fields.tags.validation",
    ]
    commandSections: ["validation", "monitoring"]
  }
}
