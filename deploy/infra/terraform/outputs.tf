output "edge_proxy_lxc_id" {
  description = "ID do container LXC do Edge Proxy"
  value       = module.lxc_edge_proxy.id
}

output "edge_proxy_lxc_ip" {
  description = "IP do container LXC do Edge Proxy"
  value       = module.lxc_edge_proxy.ipv4_address
}

output "nexus_vm_id" {
  description = "ID da VM de Produção do Nexus"
  value       = module.vm_nexus_prod.id
}

output "nexus_vm_ip" {
  description = "IP da VM de Produção do Nexus"
  value       = module.vm_nexus_prod.ipv4_address
}

output "ansible_inventory_path" {
  description = "Caminho do inventário gerado para o Ansible"
  value       = local_file.ansible_inventory.filename
}
