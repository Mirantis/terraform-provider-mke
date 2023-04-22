
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    mirantis-installers = {
      version = ">= 0.9.0"
      source  = "mirantis.com/providers/mirantis-installers"
    }
    mirantis-msr-connect = {
      version = ">= 0.9.0"
      source  = "mirantis.com/providers/mirantis-msr-connect"
    }
    mke = {
      version = ">= 0.9.0"
      source  = "mirantis.com/providers/mke"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "2.16.0"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
      version = "2.19.0"
	}
  }
}
