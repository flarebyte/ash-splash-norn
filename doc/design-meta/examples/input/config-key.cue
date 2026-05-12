i18nEntries = [
    {
        key: "fields.textInput.label"
        description: "After checkout form"
        kind: "i18n"
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
        value: "example value"
    },
]

validations = [
    {
        key: "fields.textInput.value.validation"
        commands: [
            {
                args: {
                    validation: {
                        kind: "string"
                        name: "textInput.value.validation"
                        schema: ["--required", "--min-length", "1", "--max-length", "120"]
                        schemas: []
                    }
                }
            },
            {
                args: {
                    monitoring: {
                        kind: "string"
                        name: "textInput.value.monitoring"
                        schema: ["--warn-on-missing-translation"]
                        schemas: []
                    }
                }
            },
        ]
    },
    {
        key: "fields.tags.validation"
        commands: [{
            args: {
                validation: {
                    kind: "string"
                    name: "tags.validation"
                    schema: ["--type", "list", "--required"]
                    schemas: [
                        ["--item-type", "string", "--min-length", "2", "--max-length", "20"],
                    ]
                }
            }
        }]
    },
]
