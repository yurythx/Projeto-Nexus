-- ==============================================================================
-- PROJETO NEXUS — SEMENTE DO CATÁLOGO DE SERVIÇOS CORPORATIVOS (RONDONÓPOLIS - MT)
-- ==============================================================================
-- Cadastra serviços corporativos especializados vinculados às 5 unidades de
-- Rondonópolis e associa as soluções internas de TI à unidade Hub Aurora.
-- Idempotente: pode rodar a qualquer momento.
-- ==============================================================================

BEGIN;

DO $$
DECLARE
    v_u_centro   UUID := '20000000-0000-0000-0000-000000000001'; -- Sede Centro
    v_u_aurora   UUID := '20000000-0000-0000-0000-000000000002'; -- Hub Aurora
    v_u_polo     UUID := '20000000-0000-0000-0000-000000000003'; -- Polo Industrial
    v_u_poguba   UUID := '20000000-0000-0000-0000-000000000004'; -- CIA Poguba
    v_u_agro     UUID := '20000000-0000-0000-0000-000000000005'; -- Terminal Agro

BEGIN
    -- 1. Atualizar ferramentas de infraestrutura existentes para o Hub Aurora
    UPDATE catalog_services
    SET responsible_unidade_id = v_u_aurora
    WHERE responsible_unidade_id IS NULL;

    -- 2. Serviços da Sede Corporativa Central
    INSERT INTO catalog_services (
        id, slug, title, summary, description, category, audience,
        requirements, steps, channels, sla, cost, icon, responsible_unidade_id, status, position, published_at
    ) VALUES
    (
        '80000000-0000-0000-0000-000000000001',
        'reserva-salas-auditorio',
        'Reserva de Espaços Executivos & Auditório Marechal Rondon',
        'Agendamento de salas de reunião, espaço coworking e auditório de 120 lugares para eventos institucionais na Sede Centro.',
        'Atendimento a colaboradores e parceiros para realização de assembleias, workshops, reuniões de diretoria e eventos do ecossistema de Rondonópolis.',
        'Administração e Gestão',
        'Colaboradores e parceiros institucionais',
        ARRAY['Vínculo institucional ou convite formal', 'Reserva mínima com 48 horas de antecedência para o auditório'],
        ARRAY['Verificar disponibilidade no módulo Agenda', 'Preencher o formulário de finalidade e infraestrutura necessária (som, projetor, coffee)', 'Aguardar validação da equipe de Facilities'],
        '[{"type": "online", "label": "Módulo Agenda", "value": "/agenda"}, {"type": "telefone", "label": "Ramal Facilities", "value": "107"}, {"type": "presencial", "label": "Sede Centro", "value": "Av. Marechal Rondon, 1250 - Recepção"}]'::jsonb,
        'Até 24 horas úteis',
        'Gratuito para colaboradores',
        'building',
        v_u_centro,
        'published',
        1,
        now() - interval '30 days'
    ),
    (
        '80000000-0000-0000-0000-000000000002',
        'parecer-juridico-contratos',
        'Análise de Contratos e Parecer Jurídico de Governança',
        'Análise de minutas contratuais, convênios de cooperação técnica e pareceres de conformidade jurídica e LGPD.',
        'Serviço prestado pela Controladoria e Jurídico para resguardo patrimonial e conformidade regulatória nas operações em Mato Grosso.',
        'Jurídico e Compliance',
        'Diretores, gerentes e coordenadores de departamento',
        ARRAY['Minuta do contrato em formato editável', 'Justificativa comercial ou técnica anexada'],
        ARRAY['Autuar processo eletrônico no módulo Trâmite com destino a SEDE-CENTRO / JUR', 'Juntar documentos e justificativas', 'Acompanhar emissão do parecer assinado digitalmente no Signum'],
        '[{"type": "online", "label": "Módulo Trâmite", "value": "/tramite"}, {"type": "email", "label": "Jurídico", "value": "juridico@nexus.corp"}]'::jsonb,
        '5 a 10 dias úteis',
        'Gratuito',
        'scale',
        v_u_centro,
        'published',
        2,
        now() - interval '30 days'
    ),

    -- 3. Serviços do Hub de Inovação & P&D - Vila Aurora
    (
        '80000000-0000-0000-0000-000000000003',
        'acesso-sso-vpn',
        'Gestão de Identidades, Acesso SSO e VPN Corporativa',
        'Criação de contas, liberação de escopos no Keycloak e configuração do túnel seguro WireGuard para trabalho remoto.',
        'Centralização de autenticação com padrões modernos de segurança, política de senhas fortes e MFA.',
        'Segurança e Identidade',
        'Todos os colaboradores e prestadores autorizados',
        ARRAY['Aprovação da liderança imediata', 'Dispositivo corporativo ou compatível com a política de BYOD'],
        ARRAY['Solicitar via Help Desk ou Trâmite interno', 'Configurar autenticação em dois fatores (TOTP)', 'Instalar certificado ou perfil WireGuard'],
        '[{"type": "online", "label": "Portal IAM", "value": "/perfil"}, {"type": "telefone", "label": "SOC Hub Aurora", "value": "Ramal 5510"}]'::jsonb,
        'Até 4 horas úteis',
        'Gratuito',
        'shield-check',
        v_u_aurora,
        'published',
        3,
        now() - interval '30 days'
    ),
    (
        '80000000-0000-0000-0000-000000000004',
        'suporte-apis-agrotech',
        'Consultoria & Integração de APIs AgroTech e Telemetria',
        'Suporte técnico a desenvolvedores parceiros para integração de dados de balança, sensores climáticos e telemetria.',
        'Ambiente de sandbox e documentação OpenAPI para produtores e cooperativas conectarem seus ERPs à Nexus.',
        'Pesquisa e Desenvolvimento',
        'Engenheiros de software, integradores e empresas parceiras',
        ARRAY['Chave de API emitida e homologada', 'Contrato de confidencialidade de dados (NDA)'],
        ARRAY['Acessar documentação técnica em docs/openapi.yaml', 'Testar credenciais em ambiente de homologação', 'Validar webhooks e eventos de streaming'],
        '[{"type": "online", "label": "Documentação OpenAPI", "value": "/docs"}, {"type": "email", "label": "Engenharia", "value": "dev@nexus.corp"}]'::jsonb,
        'Até 48 horas úteis',
        'Sob consulta para terceiros',
        'code',
        v_u_aurora,
        'published',
        4,
        now() - interval '30 days'
    ),

    -- 4. Serviços do Complexo Operacional - Polo Industrial
    (
        '80000000-0000-0000-0000-000000000005',
        'vistoria-frotas-bitrem',
        'Inspeção Mecânica & Manutenção Preventiva de Frotas Bitrem',
        'Agendamento de check-list de segurança mecânica, suspensão, freios pneumáticos e pneus da frota de transporte rodoviário.',
        'Oficina mecânica do Distrito Industrial equipada para atendimento de conjuntos bitrem e rodotrem da safra agrícola.',
        'Operações e Logística',
        'Motoristas próprios e transportadoras credenciadas',
        ARRAY['Documento do cavalo mecânico e carretas atualizado', 'Laudo do tacógrafo aferido'],
        ARRAY['Agendar horário na baia de inspeção do Polo Industrial', 'Apresentar o conjunto na portaria da BR-364', 'Retirar laudo emitido pelo supervisor de operações'],
        '[{"type": "presencial", "label": "Oficina Polo Industrial", "value": "Rodovia BR-364, km 205"}, {"type": "telefone", "label": "Despacho Frotas", "value": "(66) 3439-8800"}]'::jsonb,
        'No mesmo dia com agendamento',
        'Gratuito para frota própria',
        'wrench',
        v_u_polo,
        'published',
        5,
        now() - interval '30 days'
    ),
    (
        '80000000-0000-0000-0000-000000000006',
        'requisicao-almoxarifado-central',
        'Requisição de EPIs, Peças e Suprimentos de Operação',
        'Retirada de equipamentos de proteção individual com Certificado de Aprovação (CA) e insumos de oficina.',
        'Atendimento direto às equipes de campo com registro biométrico de entrega de EPIs conforme NR-6.',
        'Operações e Logística',
        'Operadores, mecânicos e equipes de pátio',
        ARRAY['Ficha de EPI ativa no sistema', 'Autorização do supervisor imediato'],
        ARRAY['Solicitar via sistema ou formulário de requisição', 'Apresentar-se no balcão do Almoxarifado Central com crachá', 'Conferir e assinar termo de recebimento'],
        '[{"type": "presencial", "label": "Balcão Almoxarifado", "value": "Distrito Industrial - Galpão 02"}, {"type": "telefone", "label": "Ramal Almoxarifado", "value": "302"}]'::jsonb,
        'Imediato no balcão',
        'Gratuito para colaboradores',
        'boxes',
        v_u_polo,
        'published',
        6,
        now() - interval '30 days'
    ),

    -- 5. Serviços do Centro Integrado de Atendimento - CIA Poguba
    (
        '80000000-0000-0000-0000-000000000007',
        'cadastro-cooperativas-agro',
        'Cadastro e Credenciamento de Cooperativas e Produtores Rurais',
        'Homologação cadastral presencial ou remota para cooperativas agropecuárias acessarem os serviços logísticos e o terminal intermodal.',
        'Atendimento humanizado na Av. Poguba para suporte à documentação de regularidade fiscal e ambiental em Mato Grosso.',
        'Atendimento e Cidadania',
        'Produtores rurais, agrônomos e representantes de cooperativas',
        ARRAY['Contrato social ou ata de eleição da diretoria', 'Cartão CNPJ e Inscrição Estadual ativa', 'Certidão negativa de débitos estaduais (SEFAZ-MT)'],
        ARRAY['Comparecer à recepção do CIA Poguba ou abrir requerimento online', 'Entrega de cópias autenticadas ou envio digital com assinatura e-CNPJ', 'Retirada do certificado de homologação'],
        '[{"type": "presencial", "label": "CIA Poguba", "value": "Av. Poguba, 850 - Sagrada Família"}, {"type": "telefone", "label": "Triagem Poguba", "value": "(66) 3425-4100"}, {"type": "email", "label": "Protocolo", "value": "protocolo@nexus.corp"}]'::jsonb,
        'Até 48 horas úteis',
        'Gratuito',
        'briefcase',
        v_u_poguba,
        'published',
        7,
        now() - interval '30 days'
    ),
    (
        '80000000-0000-0000-0000-000000000008',
        'emissao-certidoes-institucionais',
        'Emissão de Certidões de Conformidade e Declarações de Trânsito',
        'Emissão de certidões eletrônicas com código de verificação criptográfica para comprovação de entrega e operações.',
        'Documentos validados pelo módulo Signum e com verificação pública instantânea no portal.',
        'Atendimento e Cidadania',
        'Produtores, transportadoras e órgãos públicos',
        ARRAY['Número do protocolo ou ID do processo no Trâmite', 'Chave de acesso da NFe quando aplicável'],
        ARRAY['Acessar o portal de verificação ou dirigir-se ao balcão', 'Digitar o código do documento', 'Fazer download do PDF assinado digitalmente'],
        '[{"type": "online", "label": "Verificador Público", "value": "/verificar"}, {"type": "presencial", "label": "Balcão CIA Poguba", "value": "Av. Poguba, 850"}]'::jsonb,
        'Instantâneo online',
        'Gratuito',
        'file-text',
        v_u_poguba,
        'published',
        8,
        now() - interval '30 days'
    ),

    -- 6. Serviços do Terminal Intermodal & Agronegócios - Anel Viário
    (
        '80000000-0000-0000-0000-000000000009',
        'pesagem-balanca-ferroviaria',
        'Pesagem Dinâmica em Balança Rodoferroviária & Certificação',
        'Aferição de peso bruto e tara de vagões e caminhões com balanças eletrônicas de alta precisão certificadas pelo Inmetro.',
        'Operação ininterrupta 24/7 no anel viário para garantir segurança no carregamento de composições ferroviárias rumo ao Porto de Santos.',
        'Operações e Logística',
        'Transportadoras, cooperativas e concessionária da malha ferroviária',
        ARRAY['Agendamento prévio da composição ou ordem de carregamento', 'Placa do veículo ou identificação do vagão'],
        ARRAY['Ingressar na faixa sinalizada de pesagem', 'Aguardar liberação no semáforo inteligente do pátio', 'Retirar ticket eletrônico emitido automaticamente'],
        '[{"type": "presencial", "label": "Balança Terminal Agro", "value": "Anel Viário Conrado Sales Brito, s/n"}, {"type": "telefone", "label": "Balança 24h", "value": "(66) 3411-5000"}]'::jsonb,
        'Tempo médio por conjunto: 3 minutos',
        'Tarifa conforme contrato operacional',
        'scale',
        v_u_agro,
        'published',
        9,
        now() - interval '30 days'
    ),
    (
        '80000000-0000-0000-0000-000000000010',
        'agendamento-descarga-graos',
        'Agendamento de Janela de Transbordo e Descarga nos Silos',
        'Organização de slots de descarga de soja e milho para evitar filas no anel viário e otimizar o transbordo para os vagões.',
        'Sistema integrado de fila virtual com notificação via mensagem aos motoristas.',
        'Operações e Logística',
        'Produtores e transportadoras agrícolas',
        ARRAY['Ticket de triagem fitossanitária com umidade e impureza aprovados', 'Documento fiscal de origem'],
        ARRAY['Registrar o veículo na fila virtual do Terminal Agro', 'Aguardar chamada no painel do pátio de espera', 'Posicionar na moega indicada para descarga rápida'],
        '[{"type": "online", "label": "Portal de Agendamento", "value": "/servicos/agendamento-descarga-graos"}, {"type": "telefone", "label": "Coordenação de Pátio", "value": "(66) 3411-5001"}]'::jsonb,
        'Conforme janela agendada',
        'Incluso na taxa de transbordo',
        'boxes',
        v_u_agro,
        'published',
        10,
        now() - interval '30 days'
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
        responsible_unidade_id = EXCLUDED.responsible_unidade_id,
        status = 'published',
        updated_at = now();

END $$;

COMMIT;
