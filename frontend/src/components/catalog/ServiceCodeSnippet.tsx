"use client";

import { Check, Copy, FileCode2 } from "lucide-react";
import { useState } from "react";

interface SnippetData {
  filename: string;
  language: string;
  description: string;
  code: string;
}

const SERVICE_SNIPPETS: Record<string, SnippetData> = {
  ansible: {
    filename: "playbook-hardening.yml",
    language: "YAML",
    description: "Playbook Ansible padronizado para provisionamento e segurança de servidores Linux.",
    code: `---
- name: Hardening e Padronização Corporativa de Servidores
  hosts: all
  become: yes
  vars:
    ssh_port: 2222
    allowed_groups: ["wheel", "sysadmins"]
    packages_to_install:
      - ufw
      - fail2ban
      - htop
      - curl
      - jq

  tasks:
    - name: Atualizar cache de pacotes e aplicar atualizações de segurança
      apt:
        update_cache: yes
        upgrade: dist
        cache_valid_time: 3600
      when: ansible_os_family == "Debian"

    - name: Instalar utilitários essenciais
      package:
        name: "{{ packages_to_install }}"
        state: present

    - name: Configurar regras de firewall UFW
      ufw:
        rule: allow
        port: "{{ item }}"
        proto: tcp
      loop:
        - "{{ ssh_port }}"
        - 80
        - 443

    - name: Habilitar firewall com política restritiva
      ufw:
        state: enabled
        policy: deny
        direction: incoming

    - name: Configurar SSH seguro (bloqueio de root e senha)
      lineinfile:
        path: /etc/ssh/sshd_config
        regexp: "{{ item.regex }}"
        line: "{{ item.line }}"
        state: present
      loop:
        - { regex: '^#?PermitRootLogin', line: 'PermitRootLogin no' }
        - { regex: '^#?PasswordAuthentication', line: 'PasswordAuthentication no' }
        - { regex: '^#?X11Forwarding', line: 'X11Forwarding no' }
      notify: Reiniciar SSH

  handlers:
    - name: Reiniciar SSH
      service:
        name: sshd
        state: restarted`,
  },

  terraform: {
    filename: "main.tf",
    language: "HCL / Terraform",
    description: "Manifesto declarativo para infraestrutura como código com alta disponibilidade.",
    code: `terraform {
  required_version = ">= 1.8.0"
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.54.0"
    }
  }
}

# Definição de Rede Privada
resource "openstack_networking_network_v2" "vpc_interna" {
  name           = "nexus-private-network"
  admin_state_up = true
}

resource "openstack_networking_subnet_v2" "subnet_apps" {
  name       = "nexus-subnet-10-0-1-0"
  network_id = openstack_networking_network_v2.vpc_interna.id
  cidr       = "10.0.1.0/24"
  ip_version = 4
  dns_nameservers = ["1.1.1.1", "8.8.8.8"]
}

# Grupo de Segurança Restritivo
resource "openstack_networking_secgroup_v2" "sg_web" {
  name        = "nexus-sg-web"
  description = "Acesso HTTPS e tráfego interno seguro"
}

resource "openstack_networking_secgroup_rule_v2" "allow_https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.sg_web.id
}

# Instância de Servidor de Produção
resource "openstack_compute_instance_v2" "server_producao" {
  name            = "nexus-core-app-01"
  image_name      = "Ubuntu-22.04-LTS"
  flavor_name     = "m1.large"
  key_pair        = "deploy-key-corporativa"
  security_groups = [openstack_networking_secgroup_v2.sg_web.name]

  network {
    uuid = openstack_networking_network_v2.vpc_interna.id
  }

  metadata = {
    Ambiente   = "Producao"
    Gerenciado = "Terraform-Nexus"
  }
}`,
  },

  wireguard: {
    filename: "wg0.conf",
    language: "WireGuard Conf",
    description: "Configuração de túnel VPN ponto a ponto criptografado com ChaCha20-Poly1305.",
    code: `[Interface]
# IP Privado do Servidor Nexus na Rede Mesh
Address = 10.100.0.1/24
ListenPort = 51820
PrivateKey = aAAA...[CHAVE_PRIVADA_SERVIDOR]...==

# Habilitar encaminhamento e NAT iptables
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

# Cliente / Filial 01
[Peer]
PublicKey = bBBB...[CHAVE_PUBLICA_CLIENTE_01]...==
AllowedIPs = 10.100.0.2/32
PersistentKeepalive = 25`,
  },

  tailscale: {
    filename: "tailscale-acl.hujson",
    language: "HuJSON / ACL",
    description: "Política de acesso de privilégio mínimo para a tailnet corporativa.",
    code: `{
  // Grupos organizacionais mapeados
  "groups": {
    "group:engenharia": ["dev1@empresa.com", "dev2@empresa.com"],
    "group:suporte":    ["ops@empresa.com"]
  },

  // Tags para servidores e nós de infraestrutura
  "tagOwners": {
    "tag:server-prod": ["group:engenharia"],
    "tag:database":    ["group:engenharia"]
  },

  // Regras de Acesso (Zero Trust)
  "acls": [
    // Engenheiros acessam servidores de produção apenas via SSH
    {
      "action": "accept",
      "src":    ["group:engenharia"],
      "proto":  "tcp",
      "dst":    ["tag:server-prod:22", "tag:server-prod:443"]
    },
    // Aplicação acessa banco de dados na porta 5432
    {
      "action": "accept",
      "src":    ["tag:server-prod"],
      "proto":  "tcp",
      "dst":    ["tag:database:5432"]
    }
  ]
}`,
  },

  crowdsec: {
    filename: "crowdsec-scenario.yaml",
    language: "YAML",
    description: "Cenário de detecção de força bruta e ataques com inteligência comunitária.",
    code: `type: leaky
name: nexus/http-bf-crawl-detection
description: "Detectar tentativas reiteradas de força bruta em rotas sensíveis"
filter: "evt.Meta.log_type in ['http_access-log', 'nginx-access-log'] && evt.Parsed.request contains '/api/v1/auth/login'"
groupby: evt.Meta.source_ip
distinct: evt.Parsed.request
capacity: 5
leakspeed: 10s
blackhole: 1m
labels:
  service: http
  type: bruteforce
  remediation: true`,
  },

  keycloak: {
    filename: "docker-compose.keycloak.yml",
    language: "YAML",
    description: "Provisionamento do Identity Provider com banco de dados dedicado e SSL.",
    code: `services:
  keycloak:
    image: quay.io/keycloak/keycloak:26.6
    command: ["start", "--optimized"]
    environment:
      KC_DB: postgres
      KC_DB_URL: jdbc:postgresql://postgres:5432/keycloak
      KC_DB_USERNAME: keycloak
      KC_DB_PASSWORD: \${KEYCLOAK_DB_PASSWORD}
      KC_HOSTNAME: https://sso.empresa.com
      KC_PROXY_HEADERS: xforwarded
      KC_HTTP_ENABLED: "false"
    ports:
      - "8443:8443"
    healthcheck:
      test: ["CMD", "curl", "-f", "https://localhost:8443/health/ready"]
      interval: 10s
      timeout: 5s`,
  },

  n8n: {
    filename: "docker-compose.n8n.yml",
    language: "YAML",
    description: "Stack de orquestração de workflows do n8n com PostgreSQL e webhook seguro.",
    code: `services:
  n8n:
    image: docker.n8n.io/n8nio/n8n:latest
    restart: unless-stopped
    ports:
      - "5678:5678"
    environment:
      - N8N_HOST=automacao.empresa.com
      - N8N_PORT=5678
      - N8N_PROTOCOL=https
      - WEBHOOK_URL=https://automacao.empresa.com/
      - DB_TYPE=postgresdb
      - DB_POSTGRESDB_HOST=postgres
      - DB_POSTGRESDB_PORT=5432
      - DB_POSTGRESDB_DATABASE=n8n
      - DB_POSTGRESDB_USER=n8n
      - DB_POSTGRESDB_PASSWORD=\${N8N_DB_PASSWORD}
      - EXECUTIONS_DATA_PRUNE=true
      - EXECUTIONS_DATA_MAX_AGE=168
    volumes:
      - n8n_data:/home/node/.n8n`,
  },
};

export function ServiceCodeSnippet({ slug }: { slug: string }) {
  const snippet = SERVICE_SNIPPETS[slug];
  const [copied, setCopied] = useState(false);

  if (!snippet) return null;

  async function handleCopy() {
    if (!snippet?.code) return;
    try {
      await navigator.clipboard.writeText(snippet.code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback gracioso se clipboard API falhar
    }
  }

  return (
    <section aria-labelledby="tech-spec" className="flex flex-col gap-3 rounded-xl border border-surface-border bg-surface p-6 shadow-xs">
      <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <span className="inline-flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-accent">
            <FileCode2 size={14} aria-hidden="true" /> Exemplo Prático de Configuração & IaC
          </span>
          <h2 id="tech-spec" className="mt-1 text-lg font-bold text-foreground">
            Manifesto Técnico de Referência
          </h2>
          <p className="text-xs text-muted">{snippet.description}</p>
        </div>

        <button
          type="button"
          onClick={handleCopy}
          aria-label={copied ? "Código copiado" : "Copiar código para a área de transferência"}
          className="inline-flex items-center gap-1.5 self-start rounded-lg border border-surface-border bg-surface-2 px-3 py-1.5 text-xs font-semibold text-foreground transition-all hover:bg-surface-3 sm:self-auto"
        >
          {copied ? (
            <>
              <Check size={14} className="text-emerald-600 dark:text-emerald-400" aria-hidden="true" />
              <span className="text-emerald-700 dark:text-emerald-300">Copiado!</span>
            </>
          ) : (
            <>
              <Copy size={14} aria-hidden="true" />
              <span>Copiar código</span>
            </>
          )}
        </button>
      </div>

      <div className="relative mt-2 overflow-hidden rounded-lg border border-surface-border bg-slate-950 text-slate-100">
        <div className="flex items-center justify-between border-b border-slate-800 bg-slate-900/90 px-4 py-2 text-xs font-mono text-slate-400">
          <span className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full bg-emerald-500" aria-hidden="true" />
            {snippet.filename}
          </span>
          <span className="rounded bg-slate-800 px-2 py-0.5 text-[10px] font-semibold uppercase text-slate-300">
            {snippet.language}
          </span>
        </div>

        <pre className="max-h-[380px] overflow-x-auto p-4 font-mono text-xs leading-relaxed text-slate-200">
          <code>{snippet.code}</code>
        </pre>
      </div>

      <p className="text-[11px] text-muted">
        💡 <strong>Implantação Profissional:</strong> Nossos engenheiros entregam este serviço com CI/CD automatizado, monitoramento contínuo e documentação sob medida para a infraestrutura da sua empresa.
      </p>
    </section>
  );
}
