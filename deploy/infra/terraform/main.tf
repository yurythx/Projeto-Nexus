# ==============================================================================
# INFRAESTRUTURA PROXMOX VE - PROJETO NEXUS & SERVIÇOS
# ==============================================================================

# 1. Container LXC Leve: Reverse Proxy de Borda / Ferramental (Exemplo: Caddy Edge)
module "lxc_edge_proxy" {
  source = "./modules/proxmox_lxc"

  node_name        = var.proxmox_node_name
  vm_id            = 200
  hostname         = "nexus-edge-proxy"
  description      = "Reverse Proxy de Borda TLS - Projeto Nexus"
  tags             = ["prod", "edge", "proxy", "lxc"]
  cores            = 2
  memory           = 1024
  swap             = 512
  disk_storage     = var.storage_disk_pool
  disk_size        = 16
  template_file_id = "local:vztmpl/debian-12-standard_12.12-1_amd64.tar.zst"
  os_type          = "debian"
  network_bridge   = var.network_bridge
  ipv4_address     = "192.168.1.41/24"
  ipv4_gateway     = var.default_gateway
  ssh_public_keys  = var.admin_ssh_public_key != "" ? [var.admin_ssh_public_key] : []
}

# 2. VM KVM Principal: Servidor Docker Dedicado para o Projeto Nexus
# (Para provisionamento da VM completa a partir de Cloud-Init)
module "vm_nexus_prod" {
  source = "./modules/proxmox_vm"

  node_name       = var.proxmox_node_name
  vm_id           = 210
  name            = "nexus-prod-app"
  description     = "Host Docker de Produção do Projeto Nexus"
  tags            = ["prod", "docker", "nexus", "vm"]
  cores           = 4
  memory          = 8192
  disk_storage    = var.storage_disk_pool
  disk_size       = 60
  network_bridge  = var.network_bridge
  ipv4_address    = "192.168.1.42/24"
  ipv4_gateway    = var.default_gateway
  admin_user      = "root"
  ssh_public_keys = var.admin_ssh_public_key != "" ? [var.admin_ssh_public_key] : []
}

# 3. Geração Automática do Inventário do Ansible
resource "local_file" "ansible_inventory" {
  filename        = "${path.module}/../ansible/inventory.ini"
  file_permission = "0644"
  content         = <<-EOT
    # ==============================================================================
    # INVENTÁRIO GERADO AUTOMATICAMENTE VIA TERRAFORM - PROJETO NEXUS
    # Data de Geração: $${timestamp()}
    # ==============================================================================

    [all:vars]
    ansible_user=root
    ansible_python_interpreter=/usr/bin/python3
    ansible_ssh_common_args='-o StrictHostKeyChecking=no'

    # --- Hosts com Docker Engine (VMs KVM) ---
    [docker_hosts]
    nexus-prod-app ansible_host=192.168.1.42 proxmox_vmid=${module.vm_nexus_prod.id}

    # --- Containers LXC Leves (Sem Docker interno) ---
    [lxc_containers]
    nexus-edge-proxy ansible_host=192.168.1.41 proxmox_vmid=${module.lxc_edge_proxy.id}

    # --- Grupos de Aplicação ---
    [nexus_apps]
    nexus-prod-app

    [edge_proxies]
    nexus-edge-proxy
  EOT
}
