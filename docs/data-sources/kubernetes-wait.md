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
      version = "0.1.14"
    }
  }
}

provider "kubernetes-wait" {
  host  = "https://${data.google_container_cluster.my_cluster.endpoint}"
  token = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.my_cluster.master_auth[0].cluster_ca_certificate,
  )
}

data "google_client_config" "default" {
}

data "google_container_cluster" "my_cluster" {
  name     = "liam-app"
  location = "europe-west2"
  project  = "experimental-project-191516"
}


data "kubernetes-wait" "example" {
  resource_name    = "diode"
  namespace        = "infra"
  max_elapsed_time = 10
  initial_interval = 100
  multiplier       = "1.2"
  max_interval     = 5000
}
```
