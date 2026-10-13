output "edge_proxy_lxc_id" {
  description = "ID do container LXC do Edge Proxy"
  value       = module.lxc_edge_proxy.id
}

output "edge_proxy_lxc_ip" {
  description = "IP do container LXC do Edge Proxy"
  value       = module.lxc_edge_proxy.ipv4_address
}

output "nexus_host_ip" {
  description = "IP do host principal de Produção do Nexus (LXC 103)"
  value       = "192.168.1.42"
}

output "ansible_inventory_path" {
  description = "Caminho do inventário gerado para o Ansible"
  value       = local_file.ansible_inventory.filename
}
