-- Migration 000044: Sistema de Perfis, Localidades e Setores / Departamentos
-- SEMPRAS - Secretaria Municipal de Promoção e Assistência Social

-- 1. Tabela de Localidades (Polos / Unidades socioassistenciais)
CREATE TABLE IF NOT EXISTS localidades (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome        TEXT NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    tipo        TEXT NOT NULL, -- CRAS, CREAS, CENTRO_POP, CASA_MULHER, CASA_ABRIGO, CONSELHO_TUTELAR, GESTAO_SEMPRAS, OUTROS
    grupo_ad    TEXT NOT NULL DEFAULT '',
    endereco    TEXT NOT NULL DEFAULT '',
    telefone    TEXT NOT NULL DEFAULT '',
    bairro      TEXT NOT NULL DEFAULT '',
    ativo       BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_localidades_slug ON localidades (slug);
CREATE INDEX IF NOT EXISTS idx_localidades_tipo ON localidades (tipo);
CREATE INDEX IF NOT EXISTS idx_localidades_ativo ON localidades (ativo);

CREATE TRIGGER trg_localidades_set_updated_at
    BEFORE UPDATE ON localidades
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- 2. Tabela de Setores / Departamentos (amarrados a cada Localidade)
CREATE TABLE IF NOT EXISTS setores (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    localidade_id UUID NOT NULL REFERENCES localidades(id) ON DELETE CASCADE,
    nome          TEXT NOT NULL,
    slug          TEXT NOT NULL,
    tipo          TEXT NOT NULL DEFAULT 'TECNICO', -- TECNICO, RECEPCAO, GERENCIA, ADMINISTRATIVO
    grupo_ad      TEXT NOT NULL DEFAULT '',
    descricao     TEXT NOT NULL DEFAULT '',
    ativo         BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_setores_localidade_slug UNIQUE (localidade_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_setores_localidade ON setores (localidade_id);
CREATE INDEX IF NOT EXISTS idx_setores_slug ON setores (slug);

CREATE TRIGGER trg_setores_set_updated_at
    BEFORE UPDATE ON setores
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- 3. Tabela de Perfis de Acesso (RBAC)
CREATE TABLE IF NOT EXISTS perfis (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome        TEXT NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    descricao   TEXT NOT NULL DEFAULT '',
    grupo_ad    TEXT NOT NULL DEFAULT '',
    nivel       TEXT NOT NULL DEFAULT 'tecnico', -- admin, gestao, gerencia, tecnico, recepcao
    permissoes  TEXT[] NOT NULL DEFAULT '{}',
    ativo       BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_perfis_slug ON perfis (slug);
CREATE INDEX IF NOT EXISTS idx_perfis_nivel ON perfis (nivel);

CREATE TRIGGER trg_perfis_set_updated_at
    BEFORE UPDATE ON perfis
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- 4. Associação Usuário <-> Perfis
CREATE TABLE IF NOT EXISTS usuario_perfis (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    perfil_id  UUID NOT NULL REFERENCES perfis(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, perfil_id)
);

-- 5. Associação Usuário <-> Localidades (Permite 1 ou mais localidades)
CREATE TABLE IF NOT EXISTS usuario_localidades (
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    localidade_id  UUID NOT NULL REFERENCES localidades(id) ON DELETE CASCADE,
    is_principal   BOOLEAN NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, localidade_id)
);

CREATE INDEX IF NOT EXISTS idx_usuario_localidades_user ON usuario_localidades (user_id);
CREATE INDEX IF NOT EXISTS idx_usuario_localidades_loc ON usuario_localidades (localidade_id);

-- 6. Associação Usuário <-> Setores / Departamentos
CREATE TABLE IF NOT EXISTS usuario_setores (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    setor_id      UUID NOT NULL REFERENCES setores(id) ON DELETE CASCADE,
    is_principal  BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, setor_id)
);

CREATE INDEX IF NOT EXISTS idx_usuario_setores_user ON usuario_setores (user_id);
CREATE INDEX IF NOT EXISTS idx_usuario_setores_setor ON usuario_setores (setor_id);

-- ============================================================
-- SEEDS INICIAIS DE LOCALIDADES
-- ============================================================
INSERT INTO localidades (id, nome, slug, tipo, grupo_ad, endereco, telefone, bairro)
VALUES
    ('10000000-0000-0000-0000-000000000001', 'CRAS - Alfredo de Castro', 'cras-alfredo-de-castro', 'CRAS', 'Grupo_CRAS_Alfredo_de_Castro', 'Rua Principal, s/n', '(66) 3411-5101', 'Residencial Alfredo de Castro'),
    ('10000000-0000-0000-0000-000000000002', 'CRAS - Ana Carla', 'cras-ana-carla', 'CRAS', 'Grupo_CRAS_Ana_Carla', 'Av. dos Imigrantes, Qd. 12', '(66) 3411-5102', 'Jardim Ana Carla'),
    ('10000000-0000-0000-0000-000000000003', 'CRAS - Central', 'cras-central', 'CRAS', 'Grupo_CRAS_Central', 'Rua Rio Branco, 450', '(66) 3411-5103', 'Centro'),
    ('10000000-0000-0000-0000-000000000004', 'CRAS - Conjunto São José', 'cras-conjunto-sao-jose', 'CRAS', 'Grupo_CRAS_Conjunto_Sao_Jose', 'Rua D, s/n', '(66) 3411-5104', 'Conjunto São José'),
    ('10000000-0000-0000-0000-000000000005', 'CRAS - Iguaçu', 'cras-iguacu', 'CRAS', 'Grupo_CRAS_Iguacu', 'Av. Lions Internacional, 210', '(66) 3411-5105', 'Vila Iguaçu'),
    ('10000000-0000-0000-0000-000000000006', 'CRAS - Luz Dyara', 'cras-luz-dyara', 'CRAS', 'Grupo_CRAS_Luz_Dyara', 'Rua das Flores, 88', '(66) 3411-5106', 'Jardim Luz Dyara'),
    ('10000000-0000-0000-0000-000000000007', 'CRAS - Padre Lothar', 'cras-padre-lothar', 'CRAS', 'Grupo_CRAS_Padre_Lothar', 'Rua Dom Pedro II, 1020', '(66) 3411-5107', 'Vila Operária'),
    ('10000000-0000-0000-0000-000000000008', 'CRAS - Rio Vermelho', 'cras-rio-vermelho', 'CRAS', 'Grupo_CRAS_Rio_Vermelho', 'Av. Bandeirantes, 1550', '(66) 3411-5108', 'Jardim Rio Vermelho'),
    ('10000000-0000-0000-0000-000000000009', 'CRAS - Sagrada Família', 'cras-sagrada-familia', 'CRAS', 'Grupo_CRAS_Sagrada_Familia', 'Rua São Paulo, 310', '(66) 3411-5109', 'Parque Sagrada Família'),
    ('10000000-0000-0000-0000-000000000010', 'Centro POP & Abordagem Social', 'centro-pop', 'CENTRO_POP', 'Grupo_Centro_POP', 'Rua Fernando Corrêa, 1200', '(66) 3411-5110', 'Vila Aurora'),
    ('10000000-0000-0000-0000-000000000011', 'CREAS — Proteção Especial', 'creas', 'CREAS', 'Grupo_CREAS', 'Rua Arnaldo Estevão, 780', '(66) 3411-5111', 'Centro'),
    ('10000000-0000-0000-0000-000000000012', 'Casa da Mulher', 'casa-da-mulher', 'CASA_MULHER', 'Grupo_Casa_Mulher', 'Rua Poxoréo, 650', '(66) 3411-5112', 'Vila Aurora'),
    ('10000000-0000-0000-0000-000000000013', 'Casa Abrigo', 'casa-abrigo', 'CASA_ABRIGO', 'Grupo_Casa_Abrigo', 'Endereço Sigiloso', '(66) 3411-5113', 'Zona Sul'),
    ('10000000-0000-0000-0000-000000000014', 'Conselho Tutelar Central', 'conselho-tutelar-central', 'CONSELHO_TUTELAR', 'Grupo_Conselho_Tutelar_Central', 'Av. Cuiabá, 920', '(66) 3411-5114', 'Centro'),
    ('10000000-0000-0000-0000-000000000015', 'Conselho Tutelar Vila Operária', 'conselho-tutelar-vila-operaria', 'CONSELHO_TUTELAR', 'Grupo_Conselho_Tutelar_Vila_Operaria', 'Rua Castelo Branco, 400', '(66) 3411-5115', 'Vila Operária'),
    ('10000000-0000-0000-0000-000000000016', 'Sede SEMPRAS — Gestão Geral', 'gestao-sempras', 'GESTAO_SEMPRAS', 'Grupo_Assistencia_Social', 'Av. Duque de Caxias, 1000', '(66) 3411-5000', 'Vila Aurora')
ON CONFLICT (slug) DO UPDATE SET 
    nome = EXCLUDED.nome,
    tipo = EXCLUDED.tipo,
    grupo_ad = EXCLUDED.grupo_ad,
    endereco = EXCLUDED.endereco,
    telefone = EXCLUDED.telefone,
    bairro = EXCLUDED.bairro;

-- ============================================================
-- SEEDS DE SETORES / DEPARTAMENTOS POR LOCALIDADE
-- Para cada localidade cadastrada: Gerência, Equipe Técnica e Recepção
-- ============================================================
INSERT INTO setores (localidade_id, nome, slug, tipo, grupo_ad, descricao)
SELECT 
    l.id,
    'Gerência / Coordenação - ' || l.nome,
    'gerencia',
    'GERENCIA',
    'Grupo_AS_Gestao',
    'Setor responsável pela coordenação executiva e administrativa da unidade'
FROM localidades l
ON CONFLICT (localidade_id, slug) DO NOTHING;

INSERT INTO setores (localidade_id, nome, slug, tipo, grupo_ad, descricao)
SELECT 
    l.id,
    'Equipe Técnica - ' || l.nome,
    'tecnico',
    'TECNICO',
    'Grupo_Tecnico_Social',
    'Corpo técnico socioassistencial (Assistentes Sociais e Psicólogos)'
FROM localidades l
ON CONFLICT (localidade_id, slug) DO NOTHING;

INSERT INTO setores (localidade_id, nome, slug, tipo, grupo_ad, descricao)
SELECT 
    l.id,
    'Recepção e Triagem - ' || l.nome,
    'recepcao',
    'RECEPCAO',
    'Grupo_Recepcao',
    'Porta de entrada, recepção ao público, triagem inicial e fila de senhas'
FROM localidades l
ON CONFLICT (localidade_id, slug) DO NOTHING;

-- ============================================================
-- SEEDS DE PERFIS DE ACESSO
-- ============================================================
INSERT INTO perfis (id, nome, slug, descricao, grupo_ad, nivel, permissoes)
VALUES
    (
        '20000000-0000-0000-0000-000000000001',
        'Administrador Geral',
        'admin',
        'Acesso total e irrestrito a todas as localidades, setores e configurações do sistema.',
        'Grupo_Assistencia_Social_Admin',
        'admin',
        ARRAY['admin:total', 'atendimento:gerenciar', 'prontuario:gerenciar', 'localidades:gerenciar', 'perfis:gerenciar']
    ),
    (
        '20000000-0000-0000-0000-000000000002',
        'Gestor SEMPRAS',
        'gestor',
        'Gestão central da secretaria, visualização analítica, relatórios e auditoria da rede.',
        'Grupo_AS_Gestao',
        'gestao',
        ARRAY['atendimento:view_all', 'stats:view_all', 'relatorios:geral', 'localidades:view']
    ),
    (
        '20000000-0000-0000-0000-000000000003',
        'Gerente / Coordenador da Unidade',
        'gerente_unidade',
        'Coordenação operacional e gerencial de uma ou mais unidades da assistência social.',
        'Grupo_AS_Coord',
        'gerencia',
        ARRAY['atendimento:manage_unit', 'fila:gerenciar', 'stats:unit', 'relatorios:unit']
    ),
    (
        '20000000-0000-0000-0000-000000000004',
        'Técnico Social (Assistente Social / Psicólogo)',
        'tecnico',
        'Atendimento de referência, chamada de fila, evolução de parecer técnico confidencial e prontuários.',
        'Grupo_Tecnico_Social',
        'tecnico',
        ARRAY['atendimento:chamar', 'atendimento:evoluir_parecer', 'prontuario:sigiloso', 'encaminhamentos:criar']
    ),
    (
        '20000000-0000-0000-0000-000000000005',
        'Recepção e Triagem',
        'recepcao',
        'Acolhida inicial do cidadão, cadastro rápido, triagem não-sigilosa e inclusão na fila.',
        'Grupo_Recepcao',
        'recepcao',
        ARRAY['atendimento:triagem', 'fila:encaminhar', 'cidadao:buscar']
    )
ON CONFLICT (slug) DO UPDATE SET
    nome = EXCLUDED.nome,
    descricao = EXCLUDED.descricao,
    grupo_ad = EXCLUDED.grupo_ad,
    nivel = EXCLUDED.nivel,
    permissoes = EXCLUDED.permissoes;

-- ============================================================
-- VÍNCULO DOS USUÁRIOS DE TESTE COM SEUS PERFIS, LOCALIDADES E SETORES
-- ============================================================

-- 1. admin -> Perfil Administrador Geral, Localidade Sede SEMPRAS
INSERT INTO usuario_perfis (user_id, perfil_id)
SELECT u.id, p.id FROM users u, perfis p WHERE u.username = 'admin' AND p.slug = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_localidades (user_id, localidade_id, is_principal)
SELECT u.id, l.id, true FROM users u, localidades l WHERE u.username = 'admin' AND l.slug = 'gestao-sempras'
ON CONFLICT DO NOTHING;

-- 2. tecnico.cras -> Perfil Técnico, Localidade CRAS Ana Carla, Setor Técnico
INSERT INTO usuario_perfis (user_id, perfil_id)
SELECT u.id, p.id FROM users u, perfis p WHERE u.username = 'tecnico.cras' AND p.slug = 'tecnico'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_localidades (user_id, localidade_id, is_principal)
SELECT u.id, l.id, true FROM users u, localidades l WHERE u.username = 'tecnico.cras' AND l.slug = 'cras-ana-carla'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_setores (user_id, setor_id, is_principal)
SELECT u.id, s.id, true 
FROM users u
JOIN localidades l ON l.slug = 'cras-ana-carla'
JOIN setores s ON s.localidade_id = l.id AND s.slug = 'tecnico'
WHERE u.username = 'tecnico.cras'
ON CONFLICT DO NOTHING;

-- 3. recepcao.cras -> Perfil Recepção, Localidade CRAS Ana Carla, Setor Recepção
INSERT INTO usuario_perfis (user_id, perfil_id)
SELECT u.id, p.id FROM users u, perfis p WHERE u.username = 'recepcao.cras' AND p.slug = 'recepcao'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_localidades (user_id, localidade_id, is_principal)
SELECT u.id, l.id, true FROM users u, localidades l WHERE u.username = 'recepcao.cras' AND l.slug = 'cras-ana-carla'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_setores (user_id, setor_id, is_principal)
SELECT u.id, s.id, true 
FROM users u
JOIN localidades l ON l.slug = 'cras-ana-carla'
JOIN setores s ON s.localidade_id = l.id AND s.slug = 'recepcao'
WHERE u.username = 'recepcao.cras'
ON CONFLICT DO NOTHING;

-- 4. tecnico.pop -> Perfil Técnico, Localidade Centro POP, Setor Técnico
INSERT INTO usuario_perfis (user_id, perfil_id)
SELECT u.id, p.id FROM users u, perfis p WHERE u.username = 'tecnico.pop' AND p.slug = 'tecnico'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_localidades (user_id, localidade_id, is_principal)
SELECT u.id, l.id, true FROM users u, localidades l WHERE u.username = 'tecnico.pop' AND l.slug = 'centro-pop'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_setores (user_id, setor_id, is_principal)
SELECT u.id, s.id, true 
FROM users u
JOIN localidades l ON l.slug = 'centro-pop'
JOIN setores s ON s.localidade_id = l.id AND s.slug = 'tecnico'
WHERE u.username = 'tecnico.pop'
ON CONFLICT DO NOTHING;

-- 5. recepcao.pop -> Perfil Recepção, Localidade Centro POP, Setor Recepção
INSERT INTO usuario_perfis (user_id, perfil_id)
SELECT u.id, p.id FROM users u, perfis p WHERE u.username = 'recepcao.pop' AND p.slug = 'recepcao'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_localidades (user_id, localidade_id, is_principal)
SELECT u.id, l.id, true FROM users u, localidades l WHERE u.username = 'recepcao.pop' AND l.slug = 'centro-pop'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_setores (user_id, setor_id, is_principal)
SELECT u.id, s.id, true 
FROM users u
JOIN localidades l ON l.slug = 'centro-pop'
JOIN setores s ON s.localidade_id = l.id AND s.slug = 'recepcao'
WHERE u.username = 'recepcao.pop'
ON CONFLICT DO NOTHING;

-- 6. tecnico.mulher -> Perfil Técnico, Localidade Casa da Mulher, Setor Técnico
INSERT INTO usuario_perfis (user_id, perfil_id)
SELECT u.id, p.id FROM users u, perfis p WHERE u.username = 'tecnico.mulher' AND p.slug = 'tecnico'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_localidades (user_id, localidade_id, is_principal)
SELECT u.id, l.id, true FROM users u, localidades l WHERE u.username = 'tecnico.mulher' AND l.slug = 'casa-da-mulher'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_setores (user_id, setor_id, is_principal)
SELECT u.id, s.id, true 
FROM users u
JOIN localidades l ON l.slug = 'casa-da-mulher'
JOIN setores s ON s.localidade_id = l.id AND s.slug = 'tecnico'
WHERE u.username = 'tecnico.mulher'
ON CONFLICT DO NOTHING;

-- 7. tecnico.creas -> Perfil Técnico, Localidade CREAS, Setor Técnico
INSERT INTO usuario_perfis (user_id, perfil_id)
SELECT u.id, p.id FROM users u, perfis p WHERE u.username = 'tecnico.creas' AND p.slug = 'tecnico'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_localidades (user_id, localidade_id, is_principal)
SELECT u.id, l.id, true FROM users u, localidades l WHERE u.username = 'tecnico.creas' AND l.slug = 'creas'
ON CONFLICT DO NOTHING;

INSERT INTO usuario_setores (user_id, setor_id, is_principal)
SELECT u.id, s.id, true 
FROM users u
JOIN localidades l ON l.slug = 'creas'
JOIN setores s ON s.localidade_id = l.id AND s.slug = 'tecnico'
WHERE u.username = 'tecnico.creas'
ON CONFLICT DO NOTHING;
