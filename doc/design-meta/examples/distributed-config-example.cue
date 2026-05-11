app: {
  id: "c2f6c976-f498-4adb-8eb4-53136cb06a50"
  name: "nornctl"
  version: "0.1.0"
}

repositories: [
  {
    id: "92881602-dc9e-4389-befe-9a8f0519e7bf"
    name: "checkout-web"
    rootPath: "github.com/acme/checkout-web"
    branch: "main"
  },
  {
    id: "8b17d934-0d3b-40bd-8f6e-e9827b2324f4"
    name: "billing-api"
    rootPath: "github.com/acme/billing-api"
    branch: "main"
  },
]

sources: [
  {
    id: "a1f23578-ae3d-4ac1-8d6c-c7538f8da6bb"
    repositoryId: "92881602-dc9e-4389-befe-9a8f0519e7bf"
    format: "cue"
    path: "config/payment.cue"
  },
]

rules: [
  {
    id: "ee3d1516-159f-4e4e-8fe7-f0da63754b63"
    sourceId: "a1f23578-ae3d-4ac1-8d6c-c7538f8da6bb"
    targetRepositories: [
      "92881602-dc9e-4389-befe-9a8f0519e7bf",
      "8b17d934-0d3b-40bd-8f6e-e9827b2324f4",
    ]
    outputs: [
      {
        format: "json"
        path: "generated/config/payment.json"
      },
      {
        format: "yaml"
        path: "generated/config/payment.yaml"
      },
      {
        format: "go"
        path: "internal/generated/payment_config.go"
        packageName: "generated"
      },
      {
        format: "dart"
        path: "lib/generated/payment_config.dart"
      },
    ]
  },
]
