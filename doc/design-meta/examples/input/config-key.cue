i18nEntries = [
    {
        key: "fields.textInput.label"
        description: "After checkout form"
        kind: "i18n"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        translations: {
            en: {
                text: "Checkout Payment"
                context: {
                    feature: "checkout"
                    component: "payment"
                    type: "error"
                    surface: "snackbar"
                }
            }
            fr: {
                text: "Payment"
            }
        }
    },
    {
        key: "fields.textInput.tooltip"
        description: "Tooltip for text input"
        kind: "i18n"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        translations: {
            en: {
                text: "Enter the value used for checkout."
            }
            fr: {
                text: "Saisissez la valeur utilisee pour le paiement."
            }
        }
    },
    {
        key: "fields.textInput.placeholder"
        description: "Placeholder for text input"
        kind: "i18n"
        metaArgs: ["meta", "--status", "stable", "--app", "v1"]
        translations: {
            en: {
                text: "Type here"
            }
            fr: {
                text: "Saisissez ici"
            }
        }
    },
]

textEntries = [
    {
        key: "fields.textInput.value"
        description: "Default value for text input"
        kind: "text"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        value: "example value"
    },
]

validations = [
    {
        key: "fields.textInput.value.validation"
        metaArgs: ["meta", "--status", "draft", "--app", "v1"]
        commands: [
            {
                args: {
                    validation: {
                        kind: "string"
                        name: "textInput.value.validation"
                        schema: ["schema", "string", "--required", "--min-length", "1", "--max-length", "120"]
                        schemas: []
                    }
                }
            },
            {
                args: {
                    monitoring: {
                        kind: "string"
                        name: "textInput.value.monitoring"
                        schema: ["schema", "string", "--required"]
                        schemas: []
                    }
                }
            },
        ]
    },
    {
        key: "fields.tags.validation"
        metaArgs: ["meta", "--status", "experimental", "--app", "v2"]
        commands: [{
            args: {
                validation: {
                    kind: "string"
                    name: "tags.validation"
                    schema: ["schema", "string", "--alphabetic"]
                    schemas: [
                        ["schema", "repeatable", "--min-length", "1", "--max-length", "20"],
                    ]
                }
            }
        }]
    },
]
