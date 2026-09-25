import os

print("--- Starting Rebrand: Aurora -> Assistência Social ---")

# 1. brandingConfig.ts
p_branding = '/root/Projetos/Projeto-Aurora-v3/frontend/src/components/branding/brandingConfig.ts'
if os.path.exists(p_branding):
    with open(p_branding, 'r', encoding='utf-8') as f:
        c = f.read()
    c = c.replace('appName: "Projeto Aurora"', 'appName: "Assistência Social"')
    c = c.replace('appDescription: "Plataforma B2B Enterprise Single-Tenant para intranets corporativas"', 'appDescription: "Secretaria Municipal de Promoção e Assistência Social — SEMPRAS"')
    c = c.replace('orgName: "Sua Organização"', 'orgName: "Prefeitura Municipal de Rondonópolis"')
    c = c.replace('supportEmail: "suporte@empresa.exemplo"', 'supportEmail: "assistenciasocial@rondonopolis.mt.gov.br"')
    c = c.replace('supportPhone: "(11) 0000-0000"', 'supportPhone: "(66) 3411-5000"')
    c = c.replace('supportHours: "Segunda a Sexta, das 08h às 18h"', 'supportHours: "Segunda a Sexta, das 07h às 17h"')
    with open(p_branding, 'w', encoding='utf-8') as f:
        f.write(c)
    print("1. brandingConfig.ts updated")

# 2. layout.tsx
p_layout = '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/layout.tsx'
if os.path.exists(p_layout):
    with open(p_layout, 'r', encoding='utf-8') as f:
        c = f.read()
    c = c.replace('title: "Projeto Aurora"', 'title: "Assistência Social — Prefeitura Municipal de Rondonópolis"')
    c = c.replace('description: "Plataforma B2B Enterprise Single-Tenant (Monólito Modular com Feature Flags) para intranets corporativas."', 'description: "Portal de Serviços Socioassistenciais — SEMPRAS / SUAS"')
    with open(p_layout, 'w', encoding='utf-8') as f:
        f.write(c)
    print("2. layout.tsx updated")

# 3. page.tsx (Home Landing Page)
p_page = '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/page.tsx'
if os.path.exists(p_page):
    with open(p_page, 'r', encoding='utf-8') as f:
        c = f.read()

    old_desc = 'const description =\n  "Projeto Aurora — comunicação, automação, infraestrutura e suporte de TI sob medida para a sua empresa: WhatsApp Business, automações com n8n, monitoramento, servidores, backup, segurança e mais.";'
    new_desc = 'const description =\n  "Assistência Social — Rede Municipal de Proteção Social Básica e Especial, programas socioassistenciais, acolhimento e garantia de direitos da Prefeitura Municipal de Rondonópolis (SEMPRAS / SUAS).";'
    c = c.replace(old_desc, new_desc)

    c = c.replace('title: "Projeto Aurora — Serviços de TI"', 'title: "Assistência Social — SEMPRAS"')
    c = c.replace('openGraph: { title: "Projeto Aurora"', 'openGraph: { title: "Assistência Social — SEMPRAS"')
    c = c.replace('<p className="dateline">Projeto Aurora · Serviços de TI</p>', '<p className="dateline">Assistência Social · SEMPRAS / SUAS</p>')
    c = c.replace(
        'Comunicação, automação e infraestrutura de TI para a sua empresa crescer sem dor de cabeça.',
        'Rede de proteção social, acolhimento e garantia de direitos para toda a família.'
    )
    c = c.replace(
        'O <strong className="font-semibold text-foreground">Projeto Aurora</strong> implanta e mantém as\n              ferramentas que sua operação precisa — do WhatsApp Business ao monitoramento de servidores — com\n              suporte direto de quem configurou tudo.',
        'A <strong className="font-semibold text-foreground">Secretaria Municipal de Promoção e Assistência Social (SEMPRAS)</strong> promove a inclusão social, o atendimento a famílias em situação de vulnerabilidade e a defesa de direitos por meio dos CRAS, CREAS, Centro POP e serviços integrados ao SUAS.'
    )
    c = c.replace('>Ver serviços<', '>Ver programas e unidades<')
    c = c.replace('>Falar com a gente<', '>Fale com a Assistência Social<')
    c = c.replace('Alguns dos nossos serviços', 'Principais Programas e Unidades de Atendimento')
    c = c.replace('Ver todos os serviços', 'Ver todos os programas e serviços')
    c = c.replace('Não achou o que precisa?', 'Precisa de orientação ou atendimento?')
    c = c.replace(
        'Conta pra gente o que sua empresa precisa — a gente vê se já entregamos algo parecido ou monta sob medida.',
        'Procure a unidade do CRAS mais próxima do seu bairro ou entre em contato com a equipe da SEMPRAS para orientações sobre benefícios e acolhimento.'
    )
    with open(p_page, 'w', encoding='utf-8') as f:
        f.write(c)
    print("3. page.tsx updated")

# 4. manifest.ts
p_manifest = '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/manifest.ts'
if os.path.exists(p_manifest):
    with open(p_manifest, 'r', encoding='utf-8') as f:
        c = f.read()
    c = c.replace('name: "Projeto Aurora"', 'name: "Assistência Social — SEMPRAS"')
    c = c.replace('short_name: "Projeto Aurora"', 'short_name: "Assistência Social"')
    c = c.replace('description: "Plataforma B2B Enterprise Single-Tenant para intranets corporativas"', 'description: "Portal de Serviços da Assistência Social — SEMPRAS / Rondonópolis"')
    with open(p_manifest, 'w', encoding='utf-8') as f:
        f.write(c)
    print("4. manifest.ts updated")

# 5. login/page.tsx
p_login = '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/login/page.tsx'
if os.path.exists(p_login):
    with open(p_login, 'r', encoding='utf-8') as f:
        c = f.read()
    c = c.replace('Projeto Aurora', 'Assistência Social')
    with open(p_login, 'w', encoding='utf-8') as f:
        f.write(c)
    print("5. login/page.tsx updated")

# 6. Logo.tsx
p_logo = '/root/Projetos/Projeto-Aurora-v3/frontend/src/components/ui/Logo.tsx'
if os.path.exists(p_logo):
    with open(p_logo, 'r', encoding='utf-8') as f:
        c = f.read()
    c = c.replace('aria-label="Projeto Aurora"', 'aria-label="Assistência Social"')
    
    old_monogram = '''      <polygon points="14,9 11.3,19 9.6,19" fill="#F8FAFC" />
      <polygon points="14,9 18.4,19 16.7,19" fill="#F8FAFC" />
      <rect x="12.3" y="14.3" width="3.4" height="1.4" fill="#F8FAFC" />'''

    new_monogram = '''      <text
        x="14"
        y="17"
        textAnchor="middle"
        fill="#F8FAFC"
        fontSize="9.5"
        fontWeight="800"
        fontFamily="sans-serif"
        letterSpacing="-0.5"
      >
        AS
      </text>'''

    c = c.replace(old_monogram, new_monogram)
    with open(p_logo, 'w', encoding='utf-8') as f:
        f.write(c)
    print("6. Logo.tsx updated")

# 7. Other public pages
for p in [
    '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/servicos/page.tsx',
    '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/sobre/page.tsx',
    '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/contato/page.tsx',
    '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/faq/page.tsx',
    '/root/Projetos/Projeto-Aurora-v3/frontend/src/components/pwa/PwaInstallPrompt.tsx',
    '/root/Projetos/Projeto-Aurora-v3/frontend/src/components/pwa/PushNotificationToggle.tsx',
    '/root/Projetos/Projeto-Aurora-v3/frontend/src/components/settings/BrandingSettingsForm.tsx',
]:
    if os.path.exists(p):
        try:
            with open(p, 'r', encoding='utf-8') as f:
                c = f.read()
            c = c.replace('Projeto Aurora', 'Assistência Social')
            with open(p, 'w', encoding='utf-8') as f:
                f.write(c)
            print(f"Updated {p}")
        except Exception as e:
            print(f"Error updating {p}: {e}")

print("--- Frontend rebrand finished successfully ---")
