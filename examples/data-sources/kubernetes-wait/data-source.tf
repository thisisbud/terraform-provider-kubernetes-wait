/*terraform {
  required_providers {
    kubernetes-wait = {
      source  = "MehdiAtBud/kubernetes-wait"
      version = "0.1.7"
    }
  }
}

# The following example shows how to issue an HTTP GET request supplying
# an optional request header.
data "kubernetes-wait" "example" {
  resource_name    = "webhook-server"
  namespace        = "webhook-demo"
  max_elapsed_time = 10
  initial_interval = 100
  multiplier       = 1.2
  max_interval     = 500000
}
*/