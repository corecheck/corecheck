resource "null_resource" "configure_db" {
  triggers = {
    instance_id    = var.db_instance_id
    inventory_path = var.ansible_inventory_path
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -e
      cd "${path.root}/../ansible"
      ansible-playbook playbooks/00_setup_disk.yml playbooks/01_install_docker.yml playbooks/02_ssh.yml playbooks/03_setup_db.yml
    EOT

    environment = {
      DB_USER                   = var.db_user
      DB_PASSWORD               = var.db_password
      ANSIBLE_INVENTORY         = var.ansible_inventory_path
      ANSIBLE_HOST_KEY_CHECKING = "False"
      ANSIBLE_PRIVATE_KEY_FILE  = pathexpand("~/.ssh/id_ed25519")
    }
  }
}
