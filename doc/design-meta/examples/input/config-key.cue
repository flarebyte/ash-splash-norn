configurations = [
    {
        key: "checkoutPaymentErrorCardDeclined"
        description: "After checkout form"
        kind: "i18n"
        translations: {
            en: {
                text: "Checkout Payment"
                 context: {
                    feature: "checkout",
                    component: "payment",
                    type: "error",
                    surface: "snackbar"
                }
            }
            fr: {
                text: "Payment"
            }
     }
    },
    {
        key: 'checkoutPaymentErrorCardDeclinedColor'
        value: "#e4e4e4"
    }

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
