---
page_title: "kubernetes-wait Data Source - terraform-provider-kubernetes-wait"
subcategory: ""
description: |-
  k8-wait todo doc
---

# kubernetes-wait (Data Source)

k8-wait todo doc

## Example Usage

```terraform
terraform {
  required_providers {
    kubernetes-wait = {
      source  = "MehdiAtBud/kubernetes-wait"
      version = "0.1.5"
    }
  }
}

# The following example shows how to issue an HTTP GET request supplying
# an optional request header.
data "kubernetes-wait" "example" {
  kubernetes_url   = "http://127.0.0.1:8001"
  resource_name    = "webhook-server"
  namespace        = "webhook-demo"
  max_elapsed_time = 10
  initial_interval = 100
  multiplier       = 1.2
  max_interval     = 500000000000000000
}




provider "http" {
  # Configuration options
}
```
