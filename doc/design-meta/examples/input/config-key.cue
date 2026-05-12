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
                        flags: [
                            {name: "--required", valueKind: "boolean", required: true},
                            {name: "--min-length", valueKind: "number", values: ["1"]},
                            {name: "--max-length", valueKind: "number", values: ["120"]},
                        ]
                    }
                }
            },
            {
                args: {
                    monitoring: {
                        kind: "string"
                        name: "textInput.value.monitoring"
                        flags: [
                            {name: "--warn-on-missing-translation", valueKind: "boolean", required: true},
                        ]
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
                    flags: [
                        {name: "--type", valueKind: "string", values: ["list"]},
                        {name: "--required", valueKind: "boolean", required: true},
                        {name: "--item-type", valueKind: "string", values: ["string"]},
                        {name: "--min-length", valueKind: "number", values: ["2"]},
                        {name: "--max-length", valueKind: "number", values: ["20"]},
                    ]
                }
            }
        }]
    },
]
