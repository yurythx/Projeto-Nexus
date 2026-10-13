"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import {
  Zap,
  Play,
  Square,
  BarChart3,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Gauge,
  TrendingUp,
  Download,
  Copy,
  Check,
} from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/Card";

interface BenchmarkResult {
  id: string;
  targetName: string;
  url: string;
  concurrency: number;
  totalRequests: number;
  completedRequests: number;
  successCount: number; // 2xx
  rateLimitedCount: number; // 429
  clientErrorCount: number; // 4xx
  serverErrorCount: number; // 5xx
  durationMs: number;
  reqPerSec: number;
  latencies: number[];
  avgLatency: number;
  minLatency: number;
  maxLatency: number;
  p95Latency: number;
  p99Latency: number;
  timestamp: string;
}

interface LogEntry {
  id: number;
  statusCode: number;
  durationMs: number;
  timestamp: string;
}

interface TargetItem {
  id: string;
  name: string;
  url: string;
  description: string;
}

const TARGETS: TargetItem[] = [
  {
    id: "catalog",
    name: "Catálogo de Serviços Público",
    url: "/api/public/v1/catalog/services",
    description: "Endpoint público de leitura dos serviços e soluções",
  },
  {
    id: "blog",
    name: "Blog & Notícias",
    url: "/api/public/v1/blog/posts",
    description: "Endpoint de artigos públicos e destaque editorial",
  },
  {
    id: "branding",
    name: "Branding & Metadados",
    url: "/api/public/v1/branding",
    description: "Configurações institucionais e identidade visual",
  },
  {
    id: "health",
    name: "Healthcheck do Sistema (BFF / Stack)",
    url: "/api/health",
    description: "Verificação de disponibilidade dos microsserviços",
  },
  {
    id: "contact_ratelimit",
    name: "Formulário de Ouvidoria (Teste de Rate Limit)",
    url: "/api/public/v1/contact/messages",
    description: "Disparo contínuo para aferir ativação do bloqueio 429",
  },
];

const DEFAULT_TARGET: TargetItem = TARGETS[0]!;

export function LoadTestTab() {
  const [selectedTargetId, setSelectedTargetId] = useState<string>("catalog");
  const [concurrency, setConcurrency] = useState<number>(10);
  const [totalRequestsTarget, setTotalRequestsTarget] = useState<number>(100);
  const [isRunning, setIsRunning] = useState<boolean>(false);
  const [copied, setCopied] = useState<boolean>(false);

  // Métricas em tempo real
  const [currentProgress, setCurrentProgress] = useState<number>(0);
  const [liveReqPerSec, setLiveReqPerSec] = useState<number>(0);
  const [stats, setStats] = useState<{
    completed: number;
    success: number;
    rateLimited: number;
    clientErrors: number;
    serverErrors: number;
    avgLatency: number;
    minLatency: number;
    maxLatency: number;
    p95: number;
    p99: number;
  }>({
    completed: 0,
    success: 0,
    rateLimited: 0,
    clientErrors: 0,
    serverErrors: 0,
    avgLatency: 0,
    minLatency: 0,
    maxLatency: 0,
    p95: 0,
    p99: 0,
  });

  const [recentLogs, setRecentLogs] = useState<LogEntry[]>([]);
  const [history, setHistory] = useState<BenchmarkResult[]>([]);

  const abortControllerRef = useRef<AbortController | null>(null);
  const isRunningRef = useRef<boolean>(false);

  // Limpa controllers ao desmontar
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      isRunningRef.current = false;
    };
  }, []);

  const selectedTarget: TargetItem = TARGETS.find((t) => t.id === selectedTargetId) ?? DEFAULT_TARGET;

  const calculatePercentile = (sortedArr: number[], p: number): number => {
    if (sortedArr.length === 0) return 0;
    const index = Math.ceil((p / 100) * sortedArr.length) - 1;
    const val = sortedArr[Math.max(0, Math.min(index, sortedArr.length - 1))];
    return typeof val === "number" ? val : 0;
  };

  const handleStopTest = useCallback(() => {
    isRunningRef.current = false;
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setIsRunning(false);
  }, []);

  const handleStartTest = async () => {
    if (isRunning) return;

    setIsRunning(true);
    isRunningRef.current = true;
    abortControllerRef.current = new AbortController();

    const targetUrl = selectedTarget.url;
    const totalToRun = totalRequestsTarget;
    const workersCount = Math.min(concurrency, totalToRun);

    const latencies: number[] = [];
    let completed = 0;
    let success = 0;
    let rateLimited = 0;
    let clientErrors = 0;
    let serverErrors = 0;

    setRecentLogs([]);
    setStats({
      completed: 0,
      success: 0,
      rateLimited: 0,
      clientErrors: 0,
      serverErrors: 0,
      avgLatency: 0,
      minLatency: 0,
      maxLatency: 0,
      p95: 0,
      p99: 0,
    });
    setCurrentProgress(0);

    const startTime = performance.now();
    let lastSecondTime = startTime;
    let lastSecondCompleted = 0;

    let nextIndex = 0;

    const worker = async () => {
      while (nextIndex < totalToRun && isRunningRef.current) {
        const currentIndex = nextIndex++;
        const reqStart = performance.now();
        let status = 0;

        try {
          const res = await fetch(targetUrl, {
            method: targetUrl.includes("contact/messages") ? "POST" : "GET",
            headers: {
              "Content-Type": "application/json",
              Accept: "application/json",
            },
            body: targetUrl.includes("contact/messages")
              ? JSON.stringify({
                  name: `Benchmark Bot ${currentIndex}`,
                  email: `bench${currentIndex}@nexus.local`,
                  subject: "Teste de Carga Automatizado",
                  category: "duvida",
                  message: "Mensagem para teste de estresse de infraestrutura.",
                  consent: true,
                })
              : undefined,
            signal: abortControllerRef.current?.signal,
          });
          status = res.status;
        } catch (err: unknown) {
          if ((err as Error)?.name === "AbortError") {
            break;
          }
          status = 500;
        }

        const reqDuration = Math.round(performance.now() - reqStart);
        latencies.push(reqDuration);
        completed++;

        if (status >= 200 && status < 300) {
          success++;
        } else if (status === 429) {
          rateLimited++;
        } else if (status >= 400 && status < 500) {
          clientErrors++;
        } else {
          serverErrors++;
        }

        // Atualização a cada requisição ou lote
        const now = performance.now();
        if (now - lastSecondTime >= 500) {
          const deltaSec = (now - lastSecondTime) / 1000;
          const deltaReqs = completed - lastSecondCompleted;
          setLiveReqPerSec(Math.round(deltaReqs / deltaSec));
          lastSecondTime = now;
          lastSecondCompleted = completed;
        }

        setCurrentProgress(Math.round((completed / totalToRun) * 100));

        // Mantém os últimos 15 logs
        setRecentLogs((prev) => [
          {
            id: currentIndex,
            statusCode: status,
            durationMs: reqDuration,
            timestamp: new Date().toLocaleTimeString("pt-BR"),
          },
          ...prev.slice(0, 14),
        ]);

        // Cálculo de estatísticas incrementais
        const sum = latencies.reduce((a, b) => a + b, 0);
        const avg = Math.round(sum / latencies.length);
        const min = Math.min(...latencies);
        const max = Math.max(...latencies);

        const sorted = [...latencies].sort((a, b) => a - b);
        const p95 = calculatePercentile(sorted, 95);
        const p99 = calculatePercentile(sorted, 99);

        setStats({
          completed,
          success,
          rateLimited,
          clientErrors,
          serverErrors,
          avgLatency: avg,
          minLatency: min,
          maxLatency: max,
          p95,
          p99,
        });
      }
    };

    // Lança workers concorrentes
    const workers = Array.from({ length: workersCount }, () => worker());
    await Promise.all(workers);

    const totalDuration = Math.round(performance.now() - startTime);
    const overallReqPerSec = completed > 0 ? Math.round((completed / (totalDuration / 1000)) * 10) / 10 : 0;
    setLiveReqPerSec(overallReqPerSec);
    setIsRunning(false);
    isRunningRef.current = false;

    // Salva no histórico se completou ou interrompeu com requisições
    if (completed > 0) {
      const sorted = [...latencies].sort((a, b) => a - b);
      const benchmarkRecord: BenchmarkResult = {
        id: `bench-${Date.now()}`,
        targetName: selectedTarget.name,
        url: targetUrl,
        concurrency,
        totalRequests: totalToRun,
        completedRequests: completed,
        successCount: success,
        rateLimitedCount: rateLimited,
        clientErrorCount: clientErrors,
        serverErrorCount: serverErrors,
        durationMs: totalDuration,
        reqPerSec: overallReqPerSec,
        latencies,
        avgLatency: Math.round(latencies.reduce((a, b) => a + b, 0) / latencies.length),
        minLatency: Math.min(...latencies),
        maxLatency: Math.max(...latencies),
        p95Latency: calculatePercentile(sorted, 95),
        p99Latency: calculatePercentile(sorted, 99),
        timestamp: new Date().toLocaleTimeString("pt-BR"),
      };

      setHistory((prev) => [benchmarkRecord, ...prev.slice(0, 4)]);
    }
  };

  const handleCopySummary = () => {
    const summary = `--- Relatório de Benchmark Projeto Nexus ---
Alvo: ${selectedTarget.name} (${selectedTarget.url})
Concorrência: ${concurrency} workers
Requisições Concluídas: ${stats.completed}/${totalRequestsTarget}
Taxa de Vazão: ${liveReqPerSec} req/s
Latência Média: ${stats.avgLatency}ms (Min: ${stats.minLatency}ms, Max: ${stats.maxLatency}ms)
Latência P95: ${stats.p95}ms | P99: ${stats.p99}ms
Respostas: 2xx: ${stats.success} | 429: ${stats.rateLimited} | 4xx: ${stats.clientErrors} | 5xx: ${stats.serverErrors}
Data: ${new Date().toLocaleString("pt-BR")}`;

    navigator.clipboard.writeText(summary);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleExportJSON = () => {
    const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(history, null, 2));
    const downloadAnchor = document.createElement("a");
    downloadAnchor.setAttribute("href", dataStr);
    downloadAnchor.setAttribute("download", `nexus-benchmark-${Date.now()}.json`);
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Controles de Configuração e Disparo */}
      <Card className="border-surface-border">
        <CardHeader className="pb-4">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-3">
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <Zap size={20} />
              </span>
              <div>
                <CardTitle className="text-xl">Cockpit de Testes de Carga & Performance</CardTitle>
                <CardDescription>
                  Dispare requisições concorrentes em tempo real para aferir vazão (req/s), latências (p95/p99) e limites.
                </CardDescription>
              </div>
            </div>
            {isRunning ? (
              <Button variant="danger" onClick={handleStopTest} className="gap-2 shrink-0">
                <Square size={16} /> Interromper Teste
              </Button>
            ) : (
              <Button variant="primary" onClick={handleStartTest} className="gap-2 shrink-0">
                <Play size={16} /> Iniciar Teste de Carga
              </Button>
            )}
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-2">
            {/* 1. Escolha do Alvo */}
            <div className="flex flex-col gap-2">
              <label htmlFor="target-select" className="text-xs font-semibold uppercase tracking-wider text-muted">
                1. Endpoint / Cenário Alvo
              </label>
              <select
                id="target-select"
                value={selectedTargetId}
                onChange={(e) => setSelectedTargetId(e.target.value)}
                disabled={isRunning}
                className="w-full rounded-md border border-surface-border bg-surface px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              >
                {TARGETS.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </select>
              <span className="text-xs text-muted-foreground">{selectedTarget.description}</span>
            </div>

            {/* 2. Concorrência */}
            <div className="flex flex-col gap-2">
              <label className="text-xs font-semibold uppercase tracking-wider text-muted">
                2. Concorrência (Workers Simultâneos)
              </label>
              <div className="grid grid-cols-4 gap-2">
                {[5, 10, 25, 50].map((val) => (
                  <button
                    key={val}
                    type="button"
                    onClick={() => setConcurrency(val)}
                    disabled={isRunning}
                    className={`rounded-md border py-2 text-xs font-medium transition-colors ${
                      concurrency === val
                        ? "border-primary bg-primary/10 text-primary font-bold shadow-sm"
                        : "border-surface-border hover:bg-surface-elevated text-muted"
                    }`}
                  >
                    {val}x
                  </button>
                ))}
              </div>
              <span className="text-xs text-muted-foreground">{concurrency} requisições disparadas em paralelo</span>
            </div>

            {/* 3. Volume de Requisições */}
            <div className="flex flex-col gap-2">
              <label className="text-xs font-semibold uppercase tracking-wider text-muted">
                3. Volume Total de Requisições
              </label>
              <div className="grid grid-cols-4 gap-2">
                {[50, 100, 250, 500].map((val) => (
                  <button
                    key={val}
                    type="button"
                    onClick={() => setTotalRequestsTarget(val)}
                    disabled={isRunning}
                    className={`rounded-md border py-2 text-xs font-medium transition-colors ${
                      totalRequestsTarget === val
                        ? "border-primary bg-primary/10 text-primary font-bold shadow-sm"
                        : "border-surface-border hover:bg-surface-elevated text-muted"
                    }`}
                  >
                    {val}
                  </button>
                ))}
              </div>
              <span className="text-xs text-muted-foreground">Tamanho da amostra para o teste</span>
            </div>
          </div>

          {/* Barra de Progresso do Teste */}
          <div className="mt-6 flex flex-col gap-2">
            <div className="flex items-center justify-between text-xs text-muted">
              <span>
                Progresso: {stats.completed} de {totalRequestsTarget} requisições ({currentProgress}%)
              </span>
              <span>{isRunning ? "Executando benchmark concorrente..." : stats.completed > 0 ? "Benchmark concluído" : "Aguardando disparo"}</span>
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-surface-border">
              <div
                className={`h-full transition-all duration-150 ${
                  isRunning ? "bg-primary animate-pulse" : stats.completed > 0 ? "bg-success" : "bg-muted"
                }`}
                style={{ width: `${currentProgress}%` }}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* KPIs de Performance em Tempo Real */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {/* Vazão / Throughput */}
        <Card className="border-surface-border">
          <CardContent className="pt-5 pb-4">
            <div className="flex items-center justify-between text-muted mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Vazão Instantânea</span>
              <Gauge size={16} className="text-primary" />
            </div>
            <div className="text-2xl font-bold tracking-tight">
              {liveReqPerSec} <span className="text-xs font-normal text-muted">req/s</span>
            </div>
            <p className="text-[11px] text-muted-foreground mt-1">Capacidade de resposta da stack</p>
          </CardContent>
        </Card>

        {/* Latência Média */}
        <Card className="border-surface-border">
          <CardContent className="pt-5 pb-4">
            <div className="flex items-center justify-between text-muted mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Latência Média</span>
              <Clock size={16} className="text-info" />
            </div>
            <div className="text-2xl font-bold tracking-tight">
              {stats.avgLatency} <span className="text-xs font-normal text-muted">ms</span>
            </div>
            <p className="text-[11px] text-muted-foreground mt-1">
              Min: {stats.minLatency}ms | Max: {stats.maxLatency}ms
            </p>
          </CardContent>
        </Card>

        {/* Percentil P95 / P99 */}
        <Card className="border-surface-border">
          <CardContent className="pt-5 pb-4">
            <div className="flex items-center justify-between text-muted mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Percentis P95 / P99</span>
              <TrendingUp size={16} className="text-warning" />
            </div>
            <div className="text-2xl font-bold tracking-tight">
              {stats.p95} / {stats.p99} <span className="text-xs font-normal text-muted">ms</span>
            </div>
            <p className="text-[11px] text-muted-foreground mt-1">95% das respostas abaixo de {stats.p95}ms</p>
          </CardContent>
        </Card>

        {/* Taxa de Sucesso / Rate Limit */}
        <Card className="border-surface-border">
          <CardContent className="pt-5 pb-4">
            <div className="flex items-center justify-between text-muted mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Status das Respostas</span>
              <BarChart3 size={16} className="text-success" />
            </div>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold text-success">{stats.success}</span>
              {stats.rateLimited > 0 && (
                <span className="text-sm font-semibold text-warning" title="Rate Limited (HTTP 429)">
                  ({stats.rateLimited}x 429)
                </span>
              )}
              {stats.serverErrors > 0 && (
                <span className="text-sm font-semibold text-danger" title="Erros 5xx">
                  ({stats.serverErrors}x 5xx)
                </span>
              )}
            </div>
            <p className="text-[11px] text-muted-foreground mt-1">
              {stats.rateLimited > 0
                ? "Rate Limiter ativado com sucesso"
                : stats.serverErrors > 0
                ? "Falhas de servidor detectadas"
                : "Respostas HTTP 200 OK"}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Painel Inferior: Distribuição de Códigos HTTP & Logs Recentes */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Distribuição de Códigos */}
        <Card className="border-surface-border lg:col-span-1">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Distribuição de Status HTTP</CardTitle>
            <CardDescription className="text-xs">Proporção de códigos retornados pelo servidor</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {/* 2xx */}
            <div className="flex flex-col gap-1">
              <div className="flex items-center justify-between text-xs">
                <span className="flex items-center gap-1.5 font-medium text-success">
                  <CheckCircle2 size={13} /> 2xx (Sucesso)
                </span>
                <span className="font-bold">{stats.success} ({stats.completed > 0 ? Math.round((stats.success / stats.completed) * 100) : 0}%)</span>
              </div>
              <div className="h-1.5 w-full rounded-full bg-surface-border overflow-hidden">
                <div
                  className="h-full bg-success"
                  style={{ width: `${stats.completed > 0 ? (stats.success / stats.completed) * 100 : 0}%` }}
                />
              </div>
            </div>

            {/* 429 */}
            <div className="flex flex-col gap-1">
              <div className="flex items-center justify-between text-xs">
                <span className="flex items-center gap-1.5 font-medium text-warning">
                  <AlertTriangle size={13} /> 429 (Rate Limited / Defesa)
                </span>
                <span className="font-bold">{stats.rateLimited} ({stats.completed > 0 ? Math.round((stats.rateLimited / stats.completed) * 100) : 0}%)</span>
              </div>
              <div className="h-1.5 w-full rounded-full bg-surface-border overflow-hidden">
                <div
                  className="h-full bg-warning"
                  style={{ width: `${stats.completed > 0 ? (stats.rateLimited / stats.completed) * 100 : 0}%` }}
                />
              </div>
            </div>

            {/* 4xx */}
            <div className="flex flex-col gap-1">
              <div className="flex items-center justify-between text-xs">
                <span className="flex items-center gap-1.5 font-medium text-muted">
                  4xx (Erros Cliente)
                </span>
                <span className="font-bold">{stats.clientErrors} ({stats.completed > 0 ? Math.round((stats.clientErrors / stats.completed) * 100) : 0}%)</span>
              </div>
              <div className="h-1.5 w-full rounded-full bg-surface-border overflow-hidden">
                <div
                  className="h-full bg-muted"
                  style={{ width: `${stats.completed > 0 ? (stats.clientErrors / stats.completed) * 100 : 0}%` }}
                />
              </div>
            </div>

            {/* 5xx */}
            <div className="flex flex-col gap-1">
              <div className="flex items-center justify-between text-xs">
                <span className="flex items-center gap-1.5 font-medium text-danger">
                  5xx (Erros Servidor)
                </span>
                <span className="font-bold">{stats.serverErrors} ({stats.completed > 0 ? Math.round((stats.serverErrors / stats.completed) * 100) : 0}%)</span>
              </div>
              <div className="h-1.5 w-full rounded-full bg-surface-border overflow-hidden">
                <div
                  className="h-full bg-danger"
                  style={{ width: `${stats.completed > 0 ? (stats.serverErrors / stats.completed) * 100 : 0}%` }}
                />
              </div>
            </div>

            {/* Ações */}
            <div className="pt-2 flex items-center gap-2">
              <Button
                variant="secondary"
                size="sm"
                onClick={handleCopySummary}
                disabled={stats.completed === 0}
                className="w-full text-xs gap-1"
              >
                {copied ? <Check size={13} className="text-success" /> : <Copy size={13} />}
                {copied ? "Copiado!" : "Copiar Resumo"}
              </Button>
              {history.length > 0 && (
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={handleExportJSON}
                  className="w-full text-xs gap-1"
                >
                  <Download size={13} /> Exportar JSON
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Streaming de Logs Recentes */}
        <Card className="border-surface-border lg:col-span-2">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-base">Stream de Requisições ao Vivo</CardTitle>
                <CardDescription className="text-xs">Amostragem em tempo real das chamadas finalizadas</CardDescription>
              </div>
              <span className="text-xs text-muted flex items-center gap-1">
                <span className={`h-2 w-2 rounded-full ${isRunning ? "bg-success animate-ping" : "bg-muted"}`} />
                {isRunning ? "Ativo" : "Ocioso"}
              </span>
            </div>
          </CardHeader>
          <CardContent>
            {recentLogs.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-10 text-center text-muted">
                <Gauge size={32} className="opacity-30 mb-2" />
                <p className="text-sm">Nenhum teste de carga executado ainda nesta sessão.</p>
                <p className="text-xs text-muted-foreground">Clique em &quot;Iniciar Teste de Carga&quot; para começar a amostragem.</p>
              </div>
            ) : (
              <div className="overflow-x-auto max-h-[220px] overflow-y-auto pr-1">
                <table className="w-full text-left text-xs">
                  <thead className="border-b border-surface-border text-muted">
                    <tr>
                      <th className="py-1.5 font-semibold"># Req</th>
                      <th className="py-1.5 font-semibold">Horário</th>
                      <th className="py-1.5 font-semibold">Status HTTP</th>
                      <th className="py-1.5 font-semibold text-right">Latência</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-surface-border/50">
                    {recentLogs.map((log) => (
                      <tr key={log.id} className="hover:bg-surface-elevated/40">
                        <td className="py-1 font-mono text-muted">#{log.id + 1}</td>
                        <td className="py-1 font-mono">{log.timestamp}</td>
                        <td className="py-1">
                          <span
                            className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold ${
                              log.statusCode >= 200 && log.statusCode < 300
                                ? "bg-success/10 text-success"
                                : log.statusCode === 429
                                ? "bg-warning/10 text-warning"
                                : "bg-danger/10 text-danger"
                            }`}
                          >
                            {log.statusCode} {log.statusCode === 429 ? "Rate Limit" : ""}
                          </span>
                        </td>
                        <td className="py-1 text-right font-mono font-medium">
                          <span
                            className={
                              log.durationMs < 50
                                ? "text-success"
                                : log.durationMs < 200
                                ? "text-info"
                                : "text-warning"
                            }
                          >
                            {log.durationMs}ms
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Histórico da Sessão */}
      {history.length > 0 && (
        <Card className="border-surface-border">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Histórico dos Testes Realizados</CardTitle>
            <CardDescription className="text-xs">Benchmarks finalizados durante a sessão atual do operador</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="border-b border-surface-border text-muted font-semibold">
                  <tr>
                    <th className="py-2">Cenário</th>
                    <th className="py-2">Concorrência</th>
                    <th className="py-2">Reqs</th>
                    <th className="py-2">Vazão Média</th>
                    <th className="py-2">Latência Média</th>
                    <th className="py-2">P95 / P99</th>
                    <th className="py-2 text-right">Horário</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-surface-border/50">
                  {history.map((h) => (
                    <tr key={h.id} className="hover:bg-surface-elevated/40">
                      <td className="py-2 font-medium">{h.targetName}</td>
                      <td className="py-2">{h.concurrency} workers</td>
                      <td className="py-2 font-mono">
                        {h.completedRequests}/{h.totalRequests}
                      </td>
                      <td className="py-2 font-mono text-primary font-bold">{h.reqPerSec} req/s</td>
                      <td className="py-2 font-mono">{h.avgLatency}ms</td>
                      <td className="py-2 font-mono">
                        {h.p95Latency}ms / {h.p99Latency}ms
                      </td>
                      <td className="py-2 text-right text-muted font-mono">{h.timestamp}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
