# Guia de Automação de Infraestrutura: Proxmox VE + Terraform + Ansible

Este guia documenta o padrão corporativo de **Infraestrutura como Código (IaC)** e **Gerenciamento de Configuração** do Projeto Nexus no Proxmox VE 8/9.

---

## 1. Visão Geral da Arquitetura

```
[ Desenvolvedor / CI/CD ]
          │
          ├──> 1. TERRAFORM (deploy/infra/terraform)
          │    • Comunica-se com a API Proxmox (https://192.168.1.13:8006/)
          │    • Cria VMs KVM e Containers LXC no pool 'zfs'
          │    • Conecta na rede bridge 'vmbr0'
          │    • Gera automaticamente o inventário 'inventory.ini' para o Ansible
          │
          └──> 2. ANSIBLE (deploy/infra/ansible)
               • Conecta via SSH nas máquinas provisionadas
               • Aplica role 'common' (hardening, timezone, firewall UFW)
               • Aplica role 'docker' (Docker CE oficial + compose plugin)
               • Aplica role 'nexus_deploy' (clone Git + docker compose up + healthcheck)
```

---

## 2. Estrutura de Arquivos

```
deploy/infra/
├── terraform/
│   ├── versions.tf               # Provider bpg/proxmox
│   ├── provider.tf               # Autenticação via API Token
│   ├── variables.tf              # Declaração de variáveis
│   ├── terraform.tfvars          # Valores locais (protegido por .gitignore)
│   ├── terraform.tfvars.example  # Exemplo versionado no Git
│   ├── main.tf                   # Declaração dos ambientes e geração do inventário
│   ├── outputs.tf                # IPs, IDs e caminhos gerados
│   └── modules/
│       ├── proxmox_lxc/          # Módulo para criar Containers LXC
│       └── proxmox_vm/           # Módulo para criar VMs KVM
│
└── ansible/
    ├── ansible.cfg               # Otimizações de SSH e pipelining
    ├── inventory.ini             # Gerado dinamicamente pelo Terraform
    ├── site.yml                  # Playbook orquestrador mestre
    └── roles/
        ├── common/               # Atualizações, timezone, pacotes, UFW
        ├── docker/               # Instalação oficial do Docker Engine
        └── nexus_deploy/         # Deploy da stack do Projeto Nexus
```

---

## 3. Pré-Requisitos no Nó Proxmox (`pve1`)

O usuário de automação e o token de API já foram criados no Proxmox:

* **Usuário:** `terraform-prov@pve`
* **Token ID:** `tf-token`
* **Full Token ID:** `terraform-prov@pve!tf-token`

Garanta que a permissão de administrador na raiz esteja atribuída executando no shell do Proxmox:

```bash
pveum acl modify / -user terraform-prov@pve -role Administrator
```

---

## 4. Como Executar o Terraform (Provisionamento)

Entre na pasta do Terraform:

```bash
cd deploy/infra/terraform
```

### Inicialização dos Providers
```bash
terraform init
```

### Planejamento da Infraestrutura
```bash
terraform plan
```

### Aplicação (Criação das Máquinas no Proxmox)
```bash
terraform apply
```

Ao término, o Terraform:
1. Criará o container LXC (ID 200) e a VM (ID 210) no pool ZFS do Proxmox.
2. Escreverá automaticamente o arquivo `deploy/infra/ansible/inventory.ini`.

---

## 5. Como Executar o Ansible (Configuração & Deploy)

Entre na pasta do Ansible:

```bash
cd ../ansible
```

### Testar Conectividade com Todos os Hosts
```bash
ansible all -m ping
```

### Executar o Playbook Completo
```bash
ansible-playbook site.yml
```

---

## 6. Como Adicionar Novos Containers LXC no Futuro

Para subir um novo container LXC (ex: um proxy, servidor DNS ou runner de CI), basta adicionar um bloco simples no arquivo `deploy/infra/terraform/main.tf`:

```hcl
module "lxc_novo_servico" {
  source = "./modules/proxmox_lxc"

  node_name        = var.proxmox_node_name
  vm_id            = 205
  hostname         = "meu-novo-servico"
  description      = "Novo Container LXC Dedicado"
  tags             = ["prod", "servicos", "lxc"]
  cores            = 2
  memory           = 1024
  disk_storage     = var.storage_disk_pool
  disk_size        = 10
  template_file_id = "local:vztmpl/debian-12-standard_12.12-1_amd64.tar.zst"
  network_bridge   = var.network_bridge
  ipv4_address     = "192.168.1.55/24"
  ipv4_gateway     = var.default_gateway
  ssh_public_keys  = [var.admin_ssh_public_key]
}
```

E rodar:
```bash
terraform apply
```

Em menos de 10 segundos, o container estará criado e ligado no Proxmox!
