import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

vi.mock("server-only", () => ({}));
vi.mock("@/components/layout/PublicShell", () => ({
  PublicShell: ({ children }: { children: ReactNode }) => <main>{children}</main>,
}));

import FaqPage from "../page";

describe("FaqPage", () => {
  it("renderiza o cabeçalho, categorias e perguntas frequentes", () => {
    render(<FaqPage />);

    expect(screen.getByRole("heading", { name: "Perguntas Frequentes (FAQ)", level: 1 })).toBeInTheDocument();
    expect(screen.getByText("Plataforma e Arquitetura")).toBeInTheDocument();
    expect(screen.getByText("Serviços Especializados de TI")).toBeInTheDocument();
    expect(screen.getByText("O que é o Projeto Nexus e como foi desenvolvido?")).toBeInTheDocument();
    expect(screen.getByText("Ainda tem dúvidas?")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /fale conosco/i })).toBeInTheDocument();
  });
});
