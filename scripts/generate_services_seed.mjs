import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.join(__dirname, "..");

const CATEGORIES = {
  "whatsapp-api": "Comunicação",
  "whatsapp-evolution": "Comunicação",
  "evolution-api": "Comunicação",
  chatwoot: "Comunicação",
  n8n: "Automação",
  ansible: "Automação",
  glpi: "Gestão de TI",
  zabbix: "Gestão de TI",
  terraform: "Infraestrutura",
  proxmox: "Infraestrutura",
  truenas: "Infraestrutura",
  nextcloud: "Infraestrutura",
  aapanel: "Infraestrutura",
  rustdesk: "Acesso Remoto",
  "windows-server": "Acesso Remoto",
  samba: "Acesso Remoto",
  keycloak: "Segurança e Identidade",
  tailscale: "Segurança e Identidade",
  crowdsec: "Segurança e Identidade",
  "twenty-crm": "Produtividade e CRM",
  papermark: "Produtividade e CRM",
  plone: "Produtividade e CRM",
  redes: "Redes e Física",
  cabeamento: "Redes e Física",
  manutencao: "Redes e Física",
  backup: "Redes e Física",
};

const ICON_MAP = {
  "whatsapp-api": "shield",
  "whatsapp-evolution": "message-circle",
  "evolution-api": "message-circle",
  chatwoot: "message-square",
  n8n: "filter",
  ansible: "cpu",
  glpi: "wrench",
  zabbix: "eye",
  terraform: "cloud",
  proxmox: "server",
  truenas: "hard-drive",
  nextcloud: "cloud",
  aapanel: "layout-dashboard",
  rustdesk: "monitor",
  "windows-server": "users",
  samba: "folder-open",
  keycloak: "key-round",
  tailscale: "network",
  crowdsec: "shield-check",
  "twenty-crm": "building",
  papermark: "file-text",
  plone: "globe",
  redes: "network",
  cabeamento: "layers",
  manutencao: "cpu",
  backup: "archive",
};

async function main() {
  console.log("Fetching services-content.ts from papermoon repo...");
  const res = await fetch(
    "https://raw.githubusercontent.com/yurythx/papermoon/main/frontend/src/lib/services-content.ts"
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const tsContent = await res.text();

  const iconNames = [
    "MessageSquare",
    "Filter",
    "Shield",
    "Wrench",
    "Eye",
    "Server",
    "HardDrive",
    "Cloud",
    "LayoutDashboard",
    "MessageCircle",
    "Network",
    "Layers",
    "Cpu",
    "Archive",
    "Monitor",
    "Users",
    "FolderOpen",
    "Globe",
    "KeyRound",
    "Building2",
    "FileText",
    "ShieldCheck",
  ];
  const iconMocks = iconNames
    .map((n) => `const ${n} = "${n.toLowerCase()}";`)
    .join("\n");

  let jsContent = tsContent
    .replace(/import\s*\{[\s\S]*?\}\s*from\s*['"]lucide-react['"];?/, iconMocks)
    .replace(/export\s+interface[\s\S]*?(\r?\n})/g, "")
    .replace(/export\s+type[\s\S]*?;/g, "")
    .replace(/\s+as\s+const\b/g, "")
    .replace(/:\s*ServiceContent\[\]/g, "")
    .replace(/\bexport\s+/g, "");

  // Cut off anything after the SERVICES array definition
  const closeBracketIndex = jsContent.lastIndexOf("];");
  if (closeBracketIndex !== -1) {
    jsContent = jsContent.substring(0, closeBracketIndex + 2);
  }

  // Run in VM / sandbox function
  const runModule = new Function(`${jsContent}\nreturn SERVICES;`);
  const rawServices = runModule();
  console.log(`Loaded ${rawServices.length} services from PaperMoon.`);

  // Add Terraform and Ansible
  const terraformService = {
    slug: "terraform",
    name: "Terraform",
    tagline: "Infraestrutura como Código (IaC) para ambientes multi-cloud e locais.",
    description:
      "Provisione, configure e gerencie servidores, redes e storages de forma declarativa, versionada e automatizada com HashiCorp Terraform.",
    about:
      "O Terraform é a ferramenta padrão de mercado para Infraestrutura como Código (IaC), permitindo descrever topologias inteiras de servidores, redes, firewalls e storages em arquivos de configuração declarativos (HCL).\n\nEm vez de criar máquinas virtuais e configurar portas manualmente via painéis gráficos — processo lento e propenso a falhas humanas —, o Terraform documenta o estado desejado da infraestrutura no Git. Qualquer alteração passa por revisão de código, é pré-visualizada com precisão (terraform plan) e aplicada de forma atômica e previsível (terraform apply).\n\nA equipe configura os módulos de infraestrutura sob medida para Proxmox, AWS, Azure, Google Cloud ou provedores locais, estabelecendo repositório Git com controle de versão, estado remoto seguro e travamento contra concorrência.",
    differentials: [
      "Provisionamento declarativo e reprodutível: elimine configurações manuais e divergências",
      "Suporte multi-provedor: compatível com Proxmox, AWS, Azure, GCP, Cloudflare e bare-metal",
      "Histórico e governança: toda mudança na infraestrutura é auditável e versionada no Git",
      "Prevenção de falhas com plano de execução: saiba exatamente o que será alterado antes de aplicar",
      "Modularidade: componentes reutilizáveis para criação ágil de novos ambientes (Dev, Homologação, Produção)",
    ],
    papermoonDoes: [
      "Mapeia a arquitetura e define a estrutura de módulos Terraform",
      "Escreve os manifests HCL para provisionamento de VMs, redes e recursos",
      "Configura backend remoto seguro (MinIO/S3) com travamento de estado (state locking)",
      "Implementa pipeline de CI/CD para validação automática (plan/apply)",
      "Treina a equipe técnica nas operações diárias do Terraform",
    ],
    clientDoes: [
      "Fornecer credenciais de acesso ou APIs com as permissões adequadas",
      "Disponibilizar repositório Git para hospedagem dos manifests",
      "Validar a topologia e dimensionamento dos servidores sugeridos",
      "Aprovar as mudanças de infraestrutura antes da execução em produção",
    ],
    steps: [
      {
        num: "01",
        title: "Levantamento e Arquitetura",
        description:
          "Mapeamos todos os servidores, redes, sub-redes e regras de firewall atuais e planejadas para a organização.",
      },
      {
        num: "02",
        title: "Desenvolvimento dos Módulos IaC",
        description:
          "Codificamos a infraestrutura em linguagem declarativa HCL, parametrizando recursos e definindo variáveis de ambiente seguras.",
      },
      {
        num: "03",
        title: "Validação e Execução Controlada",
        description:
          "Executamos o plano de provisionamento em ambiente de testes, conferimos a conformidade e realizamos a entrega em produção com estado travado e monitorado.",
      },
    ],
    featureGroups: [
      {
        title: "Automação e Confiabilidade",
        items: [
          "Criação automática de máquinas virtuais e containers LXC no Proxmox",
          "Provisionamento de instâncias, VPCs, discos e rotas em provedores cloud",
          "Execução idempotente: apenas as alterações necessárias são aplicadas",
          "Destruição e recriação controlada de ambientes descartáveis",
        ],
      },
      {
        title: "Segurança e Estado",
        items: [
          "State file armazenado em storage seguro e criptografado",
          "Travamento de concorrência para evitar execuções simultâneas conflitantes",
          "Gerenciamento seguro de segredos e chaves de acesso via variáveis de ambiente",
          "Detecção de desvio (drift detection) entre a infraestrutura real e o código",
        ],
      },
    ],
    faqs: [
      {
        q: "O Terraform substitui ferramentas como o Ansible?",
        a: "Não, eles são complementares. O Terraform é ideal para provisionar a infraestrutura (criar a VM, configurar a rede e os discos). O Ansible é ideal para configurar o sistema operacional e instalar os softwares dentro dessa VM.",
      },
      {
        q: "Funciona com servidores locais ou somente em nuvem?",
        a: "Funciona perfeitamente em servidores locais com Proxmox VE, VMware ou OpenStack, assim como em nuvens públicas (AWS, Azure, GCP, Oracle Cloud) e híbridas.",
      },
      {
        q: "O que acontece se alguém alterar uma máquina manualmente?",
        a: "O Terraform detecta a divergência (drift) na próxima execução e permite restaurar o estado padronizado ou atualizar o código para refletir a nova realidade.",
      },
    ],
    prerequisites: [
      "Acesso com privilégios de API ao Proxmox VE ou provedor de nuvem contratado",
      "Repositório Git institucional (GitHub, GitLab ou Gitea)",
      "Bucket S3 ou MinIO compatível para armazenamento remoto do estado do Terraform",
    ],
  };

  const ansibleService = {
    slug: "ansible",
    name: "Ansible",
    tagline: "Automação de configurações, orquestração e padronização sem agentes.",
    description:
      "Gerencie o ciclo de vida, configurações e segurança de dezenas ou centenas de servidores Linux e Windows através de playbooks automatizados e idempotentes.",
    about:
      "O Ansible é uma plataforma de automação de TI open-source criada pela Red Hat que simplifica radicalmente a administração de sistemas. Diferente de soluções legadas que exigem a instalação de softwares pesados (agentes) em cada servidor, o Ansible opera de forma agentless, conectando-se diretamente via SSH em servidores Linux ou WinRM em servidores Windows.\n\nCom o Ansible, rotinas operacionais críticas — como aplicar atualizações de segurança em lote, instalar pacotes padronizados, criar usuários corporativos, configurar regras de firewall e auditar conformidade — deixam de ser feitas manualmente em cada servidor. Tudo é descrito em playbooks claros e declarativos em formato YAML.\n\nA equipe desenvolve playbooks e roles customizados para a infraestrutura da organização, implementando inventários dinâmicos, hardening de segurança e rotinas de automação com total garantia de idempotência: a execução só altera o que realmente precisa ser corrigido.",
    differentials: [
      "Arquitetura Agentless: não requer instalação nem manutenção de agentes nos servidores gerenciados",
      "Idempotência garantida: executar um playbook várias vezes só altera o que estiver fora do padrão",
      "Playbooks legíveis em YAML: documentação viva e compreensível de todas as configurações do parque",
      "Hardening e conformidade em lote: aplique políticas de segurança em dezenas de servidores em minutos",
      "Inventários dinâmicos integrados ao Proxmox, Active Directory e nuvem",
    ],
    papermoonDoes: [
      "Estrutura o inventário de servidores e grupos de máquinas da organização",
      "Desenvolve playbooks e roles modulares para as demandas operacionais",
      "Implementa rotinas automatizadas de atualização e hardening de segurança",
      "Configura o ambiente de execução centralizado (Ansible Runner / Semaphore)",
      "Capacita a equipe interna para execução e manutenção dos playbooks",
    ],
    clientDoes: [
      "Disponibilizar acesso SSH com chave pública ou WinRM nos servidores-alvo",
      "Definir a lista de servidores a serem gerenciados e suas funções",
      "Informar as regras de negócio e softwares necessários para cada grupo de máquinas",
      "Aprovar as janelas de manutenção para execução de atualizações em lote",
    ],
    steps: [
      {
        num: "01",
        title: "Mapeamento do Parque e Acessos",
        description:
          "Cadastramos os servidores no inventário, organizando-os por grupos funcionais (banco de dados, web, proxies, aplicações) e configuramos as chaves de comunicação segura.",
      },
      {
        num: "02",
        title: "Criação dos Playbooks e Políticas",
        description:
          "Codificamos as rotinas de configuração, regras de firewall, pacotes e hardening em roles e playbooks testados em ambiente de homologação.",
      },
      {
        num: "03",
        title: "Automação em Produção e Monitoramento",
        description:
          "Executamos a padronização no parque de servidores, gerando relatórios de conformidade e habilitando a execução periódica de rotinas automatizadas.",
      },
    ],
    featureGroups: [
      {
        title: "Gerenciamento de Configuração",
        items: [
          "Instalação e atualização padronizada de pacotes de software",
          "Configuração de serviços de rede, Nginx, Docker, PostgreSQL e bancos de dados",
          "Gerenciamento centralizado de usuários, grupos e chaves SSH autorizadas",
          "Distribuição e sincronização de arquivos de configuração com templates Jinja2",
        ],
      },
      {
        title: "Segurança e Hardening",
        items: [
          "Aplicação de regras de firewall UFW / iptables em massa",
          "Desativação de serviços desnecessários e portas vulneráveis",
          "Aplicação de patches críticos de segurança do sistema operacional em minutos",
          "Auditoria automatizada de conformidade com diretrizes governamentais e CIS Benchmarks",
        ],
      },
    ],
    faqs: [
      {
        q: "Preciso instalar algum software nos servidores gerenciados?",
        a: "Não. O Ansible é agentless. Ele precisa apenas do serviço SSH padrão em servidores Linux ou WinRM em servidores Windows, além do interpretador Python padrão do sistema.",
      },
      {
        q: "Existe risco de um playbook quebrar um servidor em produção?",
        a: "Os playbooks são desenvolvidos com o princípio da idempotência e testados previamente. Além disso, o Ansible possui o modo de simulação (--check / dry-run), que mostra exatamente o que seria alterado antes de qualquer comando ser executado.",
      },
      {
        q: "É possível agendar execuções automáticas?",
        a: "Sim. Podemos configurar execuções agendadas via cron, pipelines de CI/CD ou interfaces web como Semaphore / AWX para auditoria e aplicação contínua de configurações.",
      },
    ],
    prerequisites: [
      "Acesso SSH com privilégios sudo/root nos servidores Linux ou WinRM nos servidores Windows",
      "Servidores na mesma rede privada ou acessíveis via VPN corporativa (ex.: Tailscale)",
      "Inventário básico dos servidores a serem gerenciados",
    ],
  };

  const allServices = [...rawServices, terraformService, ansibleService];
  console.log(`Total services to seed: ${allServices.length}`);

  let position = 10;
  const sqlStatements = [];

  for (const s of allServices) {
    const category = CATEGORIES[s.slug] || "Geral";
    const icon = ICON_MAP[s.slug] || "file-text";

    // Build Markdown description
    let md = `## Sobre o Serviço\n\n${s.about || s.description}\n\n`;

    if (s.differentials && s.differentials.length > 0) {
      md += `## Diferenciais\n\n${s.differentials.map((d) => `- ${d}`).join("\n")}\n\n`;
    }

    if (s.papermoonDoes && s.papermoonDoes.length > 0) {
      md += `## O que nós entregamos\n\n${s.papermoonDoes.map((p) => `- ${p}`).join("\n")}\n\n`;
    }

    if (s.clientDoes && s.clientDoes.length > 0) {
      md += `## Responsabilidades do Cliente\n\n${s.clientDoes.map((c) => `- ${c}`).join("\n")}\n\n`;
    }

    if (s.featureGroups && s.featureGroups.length > 0) {
      md += `## Recursos e Funcionalidades\n\n`;
      for (const fg of s.featureGroups) {
        md += `### ${fg.title}\n\n${fg.items.map((it) => `- ${it}`).join("\n")}\n\n`;
      }
    }

    if (s.faqs && s.faqs.length > 0) {
      md += `## Perguntas Frequentes (FAQ)\n\n`;
      for (const faq of s.faqs) {
        md += `### ${faq.q}\n\n${faq.a}\n\n`;
      }
    }

    const requirements = s.prerequisites || [];
    const steps = (s.steps || []).map(
      (st) => `${st.num}. ${st.title}: ${st.description}`
    );

    const channels = [
      {
        type: "online",
        label: "Portal Institucional",
        value: "https://papermoon.cloud",
      },
      {
        type: "email",
        label: "Atendimento e Suporte",
        value: "contato@papermoon.cloud",
      },
      {
        type: "telefone",
        label: "WhatsApp Oficial",
        value: "+55 66 99216-3988",
      },
    ];

    const audience =
      "Empresas, órgãos públicos e organizações que buscam infraestrutura e serviços de TI dedicados, sob medida e com controle total dos seus dados.";
    const sla = "Implantação em até 5 a 10 dias úteis; suporte contínuo.";
    const cost = "Sob consulta / Implementação e serviços gerenciados";

    // Escape SQL strings
    const esc = (txt) => (txt ? txt.replace(/'/g, "''") : "");
    const escArr = (arr) =>
      "ARRAY[" +
      arr.map((item) => `'${esc(item)}'`).join(", ") +
      "]::text[]";
    const channelsJson = JSON.stringify(channels).replace(/'/g, "''");

    const sql = `
INSERT INTO catalog_services (
    slug,
    title,
    summary,
    description,
    category,
    audience,
    requirements,
    steps,
    channels,
    sla,
    cost,
    icon,
    status,
    position,
    published_at
) VALUES (
    '${esc(s.slug)}',
    '${esc(s.name)}',
    '${esc(s.tagline)}',
    '${esc(md.trim())}',
    '${esc(category)}',
    '${esc(audience)}',
    ${requirements.length > 0 ? escArr(requirements) : "'{}'::text[]"},
    ${steps.length > 0 ? escArr(steps) : "'{}'::text[]"},
    '${channelsJson}'::jsonb,
    '${esc(sla)}',
    '${esc(cost)}',
    '${esc(icon)}',
    'published',
    ${position},
    now()
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    summary = EXCLUDED.summary,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    audience = EXCLUDED.audience,
    requirements = EXCLUDED.requirements,
    steps = EXCLUDED.steps,
    channels = EXCLUDED.channels,
    sla = EXCLUDED.sla,
    cost = EXCLUDED.cost,
    icon = EXCLUDED.icon,
    position = EXCLUDED.position,
    updated_at = now();
`.trim();

    sqlStatements.push(sql);
    position += 10;
  }

  const slugsList = allServices.map((s) => `'${s.slug}'`).join(", ");

  const migrationFileContent = `-- +goose Up
-- Carga inicial de serviços especializados de TI (PaperMoon + Terraform + Ansible).
-- Podem ser ativados (published), despublicados (draft) ou arquivados (archived)
-- individualmente em Configurações > Gestão de Serviços (/gestao/servicos).

${sqlStatements.join("\n\n")}

-- +goose Down
DELETE FROM catalog_services WHERE slug IN (${slugsList});
`;

  const migrationPath = path.join(
    ROOT,
    "backend",
    "migrations",
    "000142_seed_papermoon_services.sql"
  );
  fs.writeFileSync(migrationPath, migrationFileContent, "utf8");
  console.log(`Generated migration: ${migrationPath}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
