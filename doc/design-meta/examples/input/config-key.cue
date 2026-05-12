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