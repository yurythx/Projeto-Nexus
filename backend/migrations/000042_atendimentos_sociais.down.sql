-- 000042_atendimentos_sociais.down.sql
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
