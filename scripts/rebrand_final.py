# -*- coding: utf-8 -*-
import os
import re

base_dir = '/root/Projetos/Projeto-Aurora-v3'
if not os.path.exists(base_dir):
    base_dir = '/root/Projetos/assistencia-social'

print(f"Applying final rebrand in: {base_dir}")

def replace_in_file(rel_path, replacements):
    path = os.path.join(base_dir, rel_path)
    if not os.path.exists(path):
        print(f"File not found: {path}")
        return
    with open(path, 'r', encoding='utf-8') as f:
        c = f.read()
    for old, new in replacements:
        c = c.replace(old, new)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(c)
    print(f"Updated: {rel_path}")

# 1. messages/pt-BR.json
replace_in_file('frontend/messages/pt-BR.json', [
    ('"dateline": "Projeto Aurora · Plataforma B2B Enterprise"', '"dateline": "Assistência Social · SEMPRAS"'),
    ('"boasVindasSemNome": "Bem-vindo(a) ao Projeto Aurora"', '"boasVindasSemNome": "Bem-vindo(a) à Assistência Social"'),
])

# 2. messages/en.json
replace_in_file('frontend/messages/en.json', [
    ('"dateline": "Projeto Aurora · Enterprise B2B Platform"', '"dateline": "Assistência Social · SEMPRAS"'),
    ('"boasVindasSemNome": "Welcome to Projeto Aurora"', '"boasVindasSemNome": "Welcome to Assistência Social"'),
])

# 3. page.tsx
replace_in_file('frontend/src/app/page.tsx', [
    ('SEMPRAS / SUAS', 'SEMPRAS'),
    ('serviços integrados ao SUAS', 'serviços socioassistenciais integrados'),
    ('(SEMPRAS / SUAS)', '(SEMPRAS)'),
    ('Assistência Social · SEMPRAS / SUAS', 'Assistência Social · SEMPRAS'),
])

# 4. layout.tsx
replace_in_file('frontend/src/app/layout.tsx', [
    ('SEMPRAS / SUAS', 'SEMPRAS'),
])

# 5. blog & public pages
replace_in_file('frontend/src/app/blog/(site)/page.tsx', [
    ('Notícias e artigos do Projeto Aurora.', 'Notícias e comunicados da Assistência Social — SEMPRAS.'),
    ('Blog — Projeto Aurora', 'Blog — Assistência Social'),
])
replace_in_file('frontend/src/app/blog/(site)/[slug]/page.tsx', [
    ('|| "Projeto Aurora"', '|| "Assistência Social"'),
])
replace_in_file('frontend/src/app/(protected)/exemplos/page.tsx', [
    ('Projeto Aurora', 'Assistência Social'),
])
replace_in_file('frontend/src/app/(protected)/integracoes/page.tsx', [
    ('Projeto Aurora', 'Assistência Social'),
])
replace_in_file('frontend/src/components/monitoring/PlatformMonitoringDashboard.tsx', [
    ('plataforma Projeto Aurora', 'plataforma de Assistência Social'),
])
replace_in_file('frontend/public/sw.js', [
    ('Service worker do Projeto Aurora', 'Service worker da Assistência Social'),
    ('payload.title || "Projeto Aurora"', 'payload.title || "Assistência Social"'),
])

print("File replacements completed!")
