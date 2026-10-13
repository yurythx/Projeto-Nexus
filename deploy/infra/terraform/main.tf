# ==============================================================================
# INFRAESTRUTURA PROXMOX VE - PROJETO NEXUS & SERVIÇOS
# ==============================================================================

# 1. Container LXC Leve: Serviços / CI-CD Runner / Borda TLS
module "lxc_edge_proxy" {
  source = "./modules/proxmox_lxc"

  node_name        = var.proxmox_node_name
  vm_id            = 200
  hostname         = "nexus-edge-proxy"
  description      = "Container Dedicado de Serviços e CI/CD Runner - Projeto Nexus"
  tags             = ["prod", "edge", "runner", "lxc"]
  cores            = 2
  memory           = 1024
  swap             = 512
  disk_storage     = var.storage_disk_pool
  disk_size        = 16
  template_file_id = "local:vztmpl/debian-12-standard_12.12-1_amd64.tar.zst"
  os_type          = "debian"
  network_bridge   = var.network_bridge
  ipv4_address     = "192.168.1.45/24"
  ipv4_gateway     = var.default_gateway
  ssh_public_keys  = var.admin_ssh_public_key != "" ? [var.admin_ssh_public_key] : []
}

# 2. Geração Automática do Inventário do Ansible
resource "local_file" "ansible_inventory" {
  filename        = "${path.module}/../ansible/inventory.ini"
  file_permission = "0644"
  content         = <<-EOT
    # ==============================================================================
    # INVENTÁRIO GERADO AUTOMATICAMENTE VIA TERRAFORM - PROJETO NEXUS
    # ==============================================================================

    [all:vars]
    ansible_user=root
    ansible_python_interpreter=/usr/bin/python3
    ansible_ssh_common_args='-o StrictHostKeyChecking=no'

    # --- Hosts com Docker Engine (Produção Principal) ---
    [docker_hosts]
    nexus-prod-app ansible_host=192.168.1.42 proxmox_vmid=103

    # --- Containers LXC Leves (Serviços e Runner) ---
    [lxc_containers]
    nexus-edge-proxy ansible_host=192.168.1.45 proxmox_vmid=${module.lxc_edge_proxy.id}

    # --- Grupos de Aplicação ---
    [nexus_apps]
    nexus-prod-app

    [edge_proxies]
    nexus-edge-proxy

    [github_runners]
    nexus-edge-proxy
  EOT
}
