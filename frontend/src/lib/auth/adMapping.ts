export interface UserServiceInfo {
  slug: string;
  name: string;
  shortName: string;
  category: string;
  unitDefault: string;
  canChangeUnit: boolean;
  featureFlag: string;
}

export const ALL_SOCIO_SERVICES: UserServiceInfo[] = [
  {
    slug: "cras",
    name: "Atendimento CRAS",
    shortName: "CRAS",
    category: "Proteção Básica",
    unitDefault: "CRAS - Central",
    canChangeUnit: true,
    featureFlag: "module_cras_enabled",
  },
  {
    slug: "centro-pop",
    name: "Atendimento Centro POP",
    shortName: "Centro POP",
    category: "População de Rua",
    unitDefault: "Centro POP",
    canChangeUnit: true,
    featureFlag: "module_centro_pop_enabled",
  },
  {
    slug: "creas",
    name: "Atendimento CREAS",
    shortName: "CREAS",
    category: "Média Complexidade",
    unitDefault: "CREAS",
    canChangeUnit: true,
    featureFlag: "module_creas_enabled",
  },
  {
    slug: "casa-da-mulher",
    name: "Atendimento Casa da Mulher",
    shortName: "Casa da Mulher",
    category: "Proteção à Mulher",
    unitDefault: "Casa da Mulher",
    canChangeUnit: true,
    featureFlag: "module_casa_mulher_enabled",
  },
  {
    slug: "conselho-tutelar",
    name: "Atendimento Conselho Tutelar",
    shortName: "Conselho Tutelar",
    category: "Garantia de Direitos",
    unitDefault: "Conselho Tutelar Central",
    canChangeUnit: true,
    featureFlag: "module_conselho_tutelar_enabled",
  },
  {
    slug: "cadunico-bolsa-familia",
    name: "Atendimento Cadastro Único",
    shortName: "CadÚnico",
    category: "Transferência de Renda",
    unitDefault: "Central CadÚnico",
    canChangeUnit: true,
    featureFlag: "module_cadunico_enabled",
  },
  {
    slug: "beneficios-eventuais",
    name: "Atendimento Benefícios Eventuais",
    shortName: "Benefícios",
    category: "Auxílios Emergenciais",
    unitDefault: "Central de Benefícios",
    canChangeUnit: true,
    featureFlag: "module_beneficios_enabled",
  },
];

export const CRAS_UNITS = [
  "CRAS - Alfredo de Castro",
  "CRAS - Ana Carla",
  "CRAS - Central",
  "CRAS - Conjunto São José",
  "CRAS - Iguaçu",
  "CRAS - Luz Dyara",
  "CRAS - Padre Lothar",
  "CRAS - Rio Vermelho",
  "CRAS - Sagrada Família",
];

export function resolveUserUnit(groups: string[] = []): string {
  const gStr = groups.join(" ").toLowerCase();
  for (const u of CRAS_UNITS) {
    const raw = u.replace("CRAS - ", "").toLowerCase();
    if (gStr.includes(raw)) {
      return u;
    }
  }
  if (gStr.includes("centro pop") || gStr.includes("pop") || gStr.includes("abordagem")) {
    return "Centro POP";
  }
  if (gStr.includes("casa da mulher") || gStr.includes("mulher")) {
    return "Casa da Mulher";
  }
  if (gStr.includes("creas")) {
    return "CREAS";
  }
  if (gStr.includes("vila operaria") || gStr.includes("vila operária")) {
    return "Conselho Tutelar Vila Operária";
  }
  if (gStr.includes("conselho")) {
    return "Conselho Tutelar Central";
  }
  return "";
}

export function getUserAssignedServices(
  groups: string[] = [],
  roles: string[] = [],
  disabledFlags: Set<string> = new Set()
): { 
  services: UserServiceInfo[]; 
  activeUnit: string; 
  isAdmin: boolean;
  isTechnician: boolean;
  isReceptionist: boolean;
} {
  const isAdmin =
    roles.includes("aurora-admin") ||
    roles.includes("admin") ||
    roles.includes("super_admin") ||
    groups.some((g) => {
      const gl = g.toLowerCase();
      return gl.includes("admin") || gl.includes("gestao") || gl.includes("gestor");
    });

  const isReceptionist = !isAdmin && groups.some((g) => {
    const gl = g.toLowerCase();
    return gl.includes("recepcao") || gl.includes("recepção") || gl.includes("atendente");
  }) && !groups.some((g) => {
    const gl = g.toLowerCase();
    return gl.includes("tecnico") || gl.includes("técnico") || gl.includes("assistente_social") || gl.includes("psicologo");
  });

  const isTechnician = isAdmin || !isReceptionist;

  const activeUnit = resolveUserUnit(groups) || (isAdmin ? "Gestão Geral SEMPRAS" : "Unidade Socioassistencial");

  const gLower = groups.map((g) => g.toLowerCase());
  const hasKw = (kw: string) => gLower.some((g) => g.includes(kw));

  const available = ALL_SOCIO_SERVICES.filter((srv) => !disabledFlags.has(srv.featureFlag));

  if (isAdmin) {
    return {
      services: available,
      activeUnit,
      isAdmin: true,
      isTechnician: true,
      isReceptionist: false,
    };
  }

  const allowed: UserServiceInfo[] = [];

  for (const srv of available) {
    let matches = false;
    switch (srv.slug) {
      case "cras":
        matches = hasKw("cras");
        break;
      case "centro-pop":
        matches = hasKw("pop") || hasKw("abordagem");
        break;
      case "creas":
        matches = hasKw("creas");
        break;
      case "casa-da-mulher":
      case "casa_mulher":
        matches = hasKw("mulher") || hasKw("abrigo");
        break;
      case "conselho-tutelar":
        matches = hasKw("conselho") || hasKw("tutelar");
        break;
      case "cadunico-bolsa-familia":
        matches = hasKw("cadunico") || hasKw("bolsa") || hasKw("cras");
        break;
      case "beneficios-eventuais":
        matches = hasKw("beneficio") || hasKw("cras");
        break;
    }

    if (matches) {
      const copy = { ...srv };
      if (copy.slug === "cras" && activeUnit.startsWith("CRAS")) {
        copy.unitDefault = activeUnit;
        copy.canChangeUnit = false;
      }
      allowed.push(copy);
    }
  }

  // Se o servidor pertence à secretaria geral mas sem setor específico, libera o CRAS como base de triagem
  if (allowed.length === 0 && (hasKw("assistencia") || hasKw("sempras"))) {
    const cras = available.find((s) => s.slug === "cras");
    if (cras) allowed.push(cras);
  }

  return {
    services: allowed,
    activeUnit,
    isAdmin: false,
    isTechnician,
    isReceptionist,
  };
}
