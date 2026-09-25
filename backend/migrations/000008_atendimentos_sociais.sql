-- +goose Up
-- Atendimentos Socioassistenciais e Prontuários do Centro POP
-- Estrutura modular para operação da rede de Assistência Social da SEMPRAS.

CREATE SEQUENCE IF NOT EXISTS atendimentos_protocol_seq START WITH 1001;

CREATE TABLE IF NOT EXISTS atendimentos (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    protocol_number      VARCHAR(30) UNIQUE NOT NULL,
    service_slug         VARCHAR(50) NOT NULL,
    unit                 VARCHAR(100) NOT NULL,
    citizen_name         VARCHAR(255) NOT NULL,
    citizen_cpf          VARCHAR(14),
    citizen_rg           VARCHAR(20),
    citizen_phone        VARCHAR(20),
    citizen_neighborhood VARCHAR(100),
    citizen_address      TEXT,
    access_form          VARCHAR(50) DEFAULT 'Demanda Espontânea',
    demand_type          VARCHAR(100) DEFAULT 'Acolhida e Atendimento Geral',
    risk_level           VARCHAR(30) DEFAULT 'Baixo',
    vulnerabilities      JSONB DEFAULT '{}'::jsonb,
    technical_notes      TEXT,
    referrals            JSONB DEFAULT '{}'::jsonb,
    status               VARCHAR(30) NOT NULL DEFAULT 'aguardando',
    created_by           UUID REFERENCES users(id) ON DELETE SET NULL,
    technician_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_atendimentos_service_unit ON atendimentos (service_slug, unit);
CREATE INDEX IF NOT EXISTS idx_atendimentos_status ON atendimentos (status);
CREATE INDEX IF NOT EXISTS idx_atendimentos_cpf ON atendimentos (citizen_cpf);
CREATE INDEX IF NOT EXISTS idx_atendimentos_created_at ON atendimentos (created_at DESC);

-- Prontuários específicos do Centro POP & População em Situação de Rua
CREATE TABLE IF NOT EXISTS centro_pop_prontuarios (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legacy_id            INTEGER,
    name                 VARCHAR(255) NOT NULL,
    preferred_name       VARCHAR(255),
    nickname             VARCHAR(100),
    cpf                  VARCHAR(14),
    rg                   VARCHAR(20),
    date_birth           DATE,
    mother_name          VARCHAR(255),
    father_name          VARCHAR(255),
    sex                  VARCHAR(20),
    gender               VARCHAR(50),
    situation            VARCHAR(100) DEFAULT 'rua',
    time_homelessness    NUMERIC(10,2),
    reason_homelessness  VARCHAR(255),
    address              TEXT,
    date_opened          DATE DEFAULT CURRENT_DATE,
    health_notes         JSONB DEFAULT '{}'::jsonb,
    photo                TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_centro_pop_cpf ON centro_pop_prontuarios (cpf);
CREATE INDEX IF NOT EXISTS idx_centro_pop_name ON centro_pop_prontuarios (name);

-- Ativação modular de cada serviço socioassistencial via Feature Flags
INSERT INTO system_features (key, description, enabled, locked, updated_at) VALUES
('module_cras_enabled', 'Atendimento CRAS (Proteção Social Básica nas 9 unidades municipais)', true, false, now()),
('module_centro_pop_enabled', 'Atendimento Centro POP & Abordagem Social para pessoas em situação de rua', true, false, now()),
('module_creas_enabled', 'Atendimento CREAS (Proteção Especial, PAEFI e Medidas Socioeducativas)', true, false, now()),
('module_casa_mulher_enabled', 'Atendimento Casa da Mulher (Enfrentamento à Violência e Acolhimento)', true, false, now()),
('module_conselho_tutelar_enabled', 'Atendimento Conselhos Tutelares (Central e Vila Operária)', true, false, now()),
('module_cadunico_enabled', 'Gestão de Cadastro Único e Transferência de Renda', true, false, now()),
('module_beneficios_enabled', 'Concessão de Benefícios Eventuais Municipais', true, false, now())
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM system_features WHERE key IN (
    'module_cras_enabled',
    'module_centro_pop_enabled',
    'module_creas_enabled',
    'module_casa_mulher_enabled',
    'module_conselho_tutelar_enabled',
    'module_cadunico_enabled',
    'module_beneficios_enabled'
);
DROP TABLE IF EXISTS centro_pop_prontuarios;
DROP TABLE IF EXISTS atendimentos;
DROP SEQUENCE IF EXISTS atendimentos_protocol_seq;
