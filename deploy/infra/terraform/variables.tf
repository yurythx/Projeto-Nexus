variable "proxmox_endpoint" {
  description = "URL da API do Proxmox VE (ex: https://192.168.1.13:8006/)"
  type        = string
  default     = "https://192.168.1.13:8006/"
}

variable "proxmox_insecure" {
  description = "Permitir certificados TLS autoassinados no Proxmox"
  type        = bool
  default     = true
}

variable "proxmox_api_token_id" {
  description = "ID completo do Token de API (ex: terraform-prov@pve!tf-token)"
  type        = string
  default     = "terraform-prov@pve!tf-token"
}

variable "proxmox_api_token_secret" {
  description = "Segredo gerado do Token de API"
  type        = string
  sensitive   = true
}

variable "proxmox_node_name" {
  description = "Nome do nó do cluster Proxmox"
  type        = string
  default     = "pve1"
}

variable "storage_disk_pool" {
  description = "Storage para discos das VMs e rootfs dos containers"
  type        = string
  default     = "zfs"
}

variable "storage_template_pool" {
  description = "Storage onde ficam os templates LXC e imagens ISO"
  type        = string
  default     = "local"
}

variable "network_bridge" {
  description = "Interface de bridge de rede principal"
  type        = string
  default     = "vmbr0"
}

variable "admin_ssh_public_key" {
  description = "Chave SSH pública injetada nas VMs e containers para acesso do Ansible"
  type        = string
  default     = ""
}

variable "default_gateway" {
  description = "Gateway padrão da rede local"
  type        = string
  default     = "192.168.1.1"
}
