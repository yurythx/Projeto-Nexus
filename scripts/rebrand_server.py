# -*- coding: utf-8 -*-
import os
import re

print("Starting clean UTF-8 rebrand on server...")

# --- 1. page.tsx ---
page_path = '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/page.tsx'
with open(page_path, 'r', encoding='utf-8') as f:
    text = f.read()

# Replace metadata block
text = re.sub(
    r'const description =\s*".*?";',
    'const description =\n  "Assistência Social — Rede Municipal de Proteção Social Básica e Especial, programas socioassistenciais, acolhimento e garantia de direitos da Prefeitura Municipal de Rondonópolis (SEMPRAS / SUAS).";',
    text,
    flags=re.DOTALL
)

text = re.sub(r'title:\s*".*?"', 'title: "Assistência Social — SEMPRAS"', text, count=1)
text = re.sub(r'openGraph:\s*\{\s*title:\s*".*?"', 'openGraph: { title: "Assistência Social — SEMPRAS"', text)

# Replace hero content
text = re.sub(
    r'<p className="dateline">.*?</p>',
    '<p className="dateline">Assistência Social · SEMPRAS / SUAS</p>',
    text,
    count=1
)

text = re.sub(
    r'<h1 className="max-w-2xl text-4xl font-extrabold leading-\[1\.15\] text-foreground sm:text-5xl">.*?</h1>',
    '<h1 className="max-w-2xl text-4xl font-extrabold leading-[1.15] text-foreground sm:text-5xl">\n              Rede de proteção social, acolhimento e garantia de direitos para toda a família.\n            </h1>',
    text,
    flags=re.DOTALL,
    count=1
)

text = re.sub(
    r'<p className="max-w-2xl text-muted text-base">.*?</p>',
    '<p className="max-w-2xl text-muted text-base">\n              A <strong className="font-semibold text-foreground">Secretaria Municipal de Promoção e Assistência Social (SEMPRAS)</strong> promove a inclusão social, o atendimento a famílias em situação de vulnerabilidade e a defesa de direitos por meio dos CRAS, CREAS, Centro POP e serviços integrados ao SUAS.\n            </p>',
    text,
    flags=re.DOTALL,
    count=1
)

text = text.replace('Ver serviços', 'Ver programas e unidades')
text = text.replace('Falar com a gente', 'Fale com a Assistência Social')
text = text.replace('Alguns dos nossos serviços', 'Principais Programas e Unidades de Atendimento')
text = text.replace('Ver todos os serviços', 'Ver todos os programas e serviços')
text = text.replace('Não achou o que precisa?', 'Precisa de orientação ou atendimento?')
text = re.sub(
    r'<p className="max-w-md text-sm text-muted">.*?</p>',
    '<p className="max-w-md text-sm text-muted">Procure a unidade do CRAS mais próxima do seu bairro ou entre em contato com a equipe da SEMPRAS para orientações sobre benefícios e acolhimento.</p>',
    text,
    flags=re.DOTALL,
    count=1
)

with open(page_path, 'w', encoding='utf-8') as f:
    f.write(text)
print("1. page.tsx updated with clean UTF-8")

# --- 2. layout.tsx ---
layout_path = '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/layout.tsx'
with open(layout_path, 'r', encoding='utf-8') as f:
    text = f.read()

text = re.sub(
    r'export const metadata: Metadata = \{.*?\};',
    'export const metadata: Metadata = {\n  title: "Assistência Social — Prefeitura Municipal de Rondonópolis",\n  description: "Portal de Serviços Socioassistenciais — SEMPRAS / SUAS",\n};',
    text,
    flags=re.DOTALL
)

with open(layout_path, 'w', encoding='utf-8') as f:
    f.write(text)
print("2. layout.tsx updated with clean UTF-8")

# --- 3. brandingConfig.ts ---
branding_path = '/root/Projetos/Projeto-Aurora-v3/frontend/src/components/branding/brandingConfig.ts'
with open(branding_path, 'r', encoding='utf-8') as f:
    text = f.read()

new_branding = '''export const DEFAULT_BRANDING: SystemBrandingConfig = {
  appName: "Assistência Social",
  appDescription: "Secretaria Municipal de Promoção e Assistência Social — SEMPRAS",
  orgName: "Prefeitura Municipal de Rondonópolis",
  logoUrl: "",
  faviconUrl: "",
  supportEmail: "assistenciasocial@rondonopolis.mt.gov.br",
  supportPhone: "(66) 3411-5000",
  supportHours: "Segunda a Sexta, das 07h às 17h",
  highContrast: false,
  fontSizeScale: 100,
};'''

text = re.sub(r'export const DEFAULT_BRANDING: SystemBrandingConfig = \{.*?\};', new_branding, text, flags=re.DOTALL)
with open(branding_path, 'w', encoding='utf-8') as f:
    f.write(text)
print("3. brandingConfig.ts updated with clean UTF-8")

# --- 4. manifest.ts ---
manifest_path = '/root/Projetos/Projeto-Aurora-v3/frontend/src/app/manifest.ts'
with open(manifest_path, 'r', encoding='utf-8') as f:
    text = f.read()

text = re.sub(r'name:\s*"[^"]+"', 'name: "Assistência Social — SEMPRAS"', text)
text = re.sub(r'short_name:\s*"[^"]+"', 'short_name: "Assistência Social"', text)
text = re.sub(r'description:\s*"[^"]+"', 'description: "Portal de Serviços da Assistência Social — SEMPRAS / Rondonópolis"', text)
with open(manifest_path, 'w', encoding='utf-8') as f:
    f.write(text)
print("4. manifest.ts updated with clean UTF-8")

# --- 5. Logo.tsx ---
logo_path = '/root/Projetos/Projeto-Aurora-v3/frontend/src/components/ui/Logo.tsx'
with open(logo_path, 'r', encoding='utf-8') as f:
    text = f.read()

text = text.replace('aria-label="Projeto Aurora"', 'aria-label="Assistência Social"')
if '<text' not in text:
    monogram = '''      {/* Monograma: AS (Assistência Social) */}
      <text
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
    text = re.sub(
        r'<polygon points="14,9 11\.3,19 9\.6,19"[^>]+/>\s*<polygon points="14,9 18\.4,19 16\.7,19"[^>]+/>\s*<rect x="12\.3" y="14\.3" width="3\.4" height="1\.4"[^>]+/>',
        monogram,
        text
    )
with open(logo_path, 'w', encoding='utf-8') as f:
    f.write(text)
print("5. Logo.tsx updated with clean UTF-8")

print("Rebrand script completed successfully!")
