output "nexus_app_lxc_id" {
  description = "ID do container LXC de Producao do Nexus"
  value       = module.lxc_nexus_app.id
}

output "nexus_app_lxc_ip" {
  description = "IP do container LXC de Producao do Nexus"
  value       = module.lxc_nexus_app.ipv4_address
}

output "edge_proxy_lxc_id" {
  description = "ID do container LXC do Edge Proxy"
  value       = module.lxc_edge_proxy.id
}

output "edge_proxy_lxc_ip" {
  description = "IP do container LXC do Edge Proxy"
  value       = module.lxc_edge_proxy.ipv4_address
}

output "ansible_inventory_path" {
  description = "Caminho do inventário gerado para o Ansible"
  value       = local_file.ansible_inventory.filename
}
