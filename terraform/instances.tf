resource "google_service_account" "core-backend" {
  account_id   = "core-backend"
  display_name = "Core Backend"
}

resource "google_compute_disk" "core-data" {
  name                      = "core-data"
  type                      = "pd-ssd"
  zone                      = "europe-west1-b"
  physical_block_size_bytes = 4096
  size                      = 10
  lifecycle {
    prevent_destroy = false
  }
}

resource "google_compute_address" "core" {
  name         = "core"
  address_type = "EXTERNAL"
  region       = "europe-west1"
  lifecycle {
    prevent_destroy = false
  }
}

resource "google_compute_instance" "core" {
  attached_disk {
    device_name = "data"
    mode        = "READ_WRITE"
    source      = google_compute_disk.core-data.self_link
  }

  boot_disk {
    auto_delete = true
    device_name = "core"

    initialize_params {
      image = "projects/ubuntu-os-cloud/global/images/ubuntu-2204-jammy-v20230727"
      size  = 10
      type  = "pd-balanced"
    }

    mode = "READ_WRITE"
  }

  can_ip_forward      = false
  deletion_protection = false
  enable_display      = false

  labels = {
    goog-ec-src = "vm_add-tf"
  }

  machine_type = "e2-micro"

  metadata = {
    ssh-keys = "aureleoules:ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEM79mi/xHOtZw+bUfOH8soMjCyO5qOdpLls1tXnR2AD aurele@oules.com\nubuntu:ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAID10ieUzFSRpKXI1lR5BAMqe3rz7cyMBKBIaYIJyXGub bitcoin-coverage-ci"
  }

  name = "core"

  network_interface {
    access_config {
      nat_ip       = google_compute_address.core.address
      network_tier = "PREMIUM"
    }

    subnetwork = "projects/sandbox-395722/regions/europe-west1/subnetworks/default"
  }

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }

  service_account {
    email  = google_service_account.core-backend.email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  shielded_instance_config {
    enable_integrity_monitoring = true
    enable_secure_boot          = false
    enable_vtpm                 = true
  }

  tags = ["http-server", "https-server"]
  zone = "europe-west1-b"
}
