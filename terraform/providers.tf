terraform {
required_version = ">= 0.14.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.79.0"
    }
  }

  backend "gcs" {
    bucket  = "bitcoin-coverage-tfstate"
    prefix  = "terraform/state"
  }
}

provider "google" {
  project     = "sandbox-395722"
  region      = "europe-west1"
}