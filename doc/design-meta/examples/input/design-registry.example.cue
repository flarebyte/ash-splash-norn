package designregistry

designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: {
        id: "input-field"
        version: "1.0.0"
      }
      supportedLanguages: ["en", "fr"]
      supportedCommandSections: ["validation", "monitoring", "transform"]
      translationPolicy: {
        requireAllSupportedLanguages: true
      }
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
          childLabels: ["textInput", "tags"]
        }
        textInput: {
          label: "textInput"
          kind: "branch"
          childLabels: ["label", "tooltip", "placeholder", "value"]
        }
        tags: {
          label: "tags"
          kind: "branch"
          childLabels: ["validation"]
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
        from: "label-path-camelCase"
        allowCycles: false
        onGeneratedKeyCollision: "error"
      }
    }
  }

  generatorCapabilities: [
    {
      keySchema: "input-field"
      target: "arb.json"
      supportsNodeKinds: ["i18n"]
      artifactPattern: "lib/l10n/app_<locale>.arb.json"
    },
    {
      keySchema: "input-field"
      target: "json"
      supportsNodeKinds: ["i18n", "text"]
      artifactPattern: "generated/config/<domain>.json"
    },
    {
      keySchema: "input-field"
      target: "cue"
      supportsNodeKinds: ["i18n", "text"]
      artifactPattern: "generated/config/<domain>.cue"
    },
  ]
}
