"use client";

import { useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { useToast } from "@/components/notifications/ToastProvider";
import { apiClient, ApiError } from "@/lib/api/client";
import { MODULE, useFeature } from "@/lib/features/FeatureFlagProvider";
import type { EgressPushVapidKey } from "@/types/api";

// PushNotificationToggle — inscrição self-service em notificações Web
// Push (roadmap §5.3/item 8, canal push do Egress). Vive ao lado de
// PwaInstallPrompt/ServiceWorkerRegistration (é a MESMA infraestrutura de
// service worker do PWA — ver public/sw.js's listener de "push" — não um
// pacote próprio), mas é renderizado dentro do perfil (MyProfileForm),
// não solto no shell: é uma preferência que a pessoa ativa quando quer,
// não um banner que aparece sozinho.
//
// "Ativar" pede permissão do navegador (Notification.requestPermission),
// assina no PushManager e manda a inscrição pro backend
// (POST /me equivalente: v1/egress/push/subscribe). "Desativar" desfaz os
// dois lados. Nada disso é persistido em nenhum React Context — o estado
// de verdade é o PushManager do navegador (getSubscription()), lido de
// novo a cada montagem.
export function PushNotificationToggle() {
  const egressEnabled = useFeature(MODULE.egress);
  const { showToast } = useToast();
  const [supported, setSupported] = useState(false);
  const [vapid, setVapid] = useState<EgressPushVapidKey | null>(null);
  const [subscribed, setSubscribed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    if (!egressEnabled) return;
    let cancelled = false;
    // Tudo (inclusive a detecção de suporte do navegador, síncrona em si)
    // roda dentro da cadeia de promises — setState direto no corpo do
    // efeito (fora de um .then()) é o padrão que
    // react-hooks/set-state-in-effect reprova, mesmo para uma checagem
    // síncrona sem efeito colateral (mesma lição de PwaInstallPrompt.tsx).
    Promise.resolve()
      .then(() => {
        const hasAPIs = "serviceWorker" in navigator && "PushManager" in window && "Notification" in window;
        if (!hasAPIs) throw new Error("push não suportado neste navegador");
        return Promise.all([
          apiClient.get<EgressPushVapidKey>("v1/egress/push/vapid-public-key").then(({ data }) => data),
          navigator.serviceWorker.ready.then((reg) => reg.pushManager.getSubscription()),
        ]);
      })
      .then(([vapidData, existingSub]) => {
        if (cancelled) return;
        setSupported(true);
        setVapid(vapidData);
        setSubscribed(existingSub != null);
      })
      .catch(() => {
        if (!cancelled) setSupported(false);
      })
      .finally(() => {
        if (!cancelled) setChecking(false);
      });
    return () => {
      cancelled = true;
    };
  }, [egressEnabled]);

  async function enable() {
    if (busy || !vapid?.public_key) return;
    setBusy(true);
    try {
      const permission = await Notification.requestPermission();
      if (permission !== "granted") {
        showToast({ title: "Permissão negada", description: "O navegador não autorizou notificações.", tone: "danger" });
        return;
      }
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapid.public_key),
      });
      const json = subscription.toJSON();
      await apiClient.post("v1/egress/push/subscribe", {
        endpoint: json.endpoint,
        keys: { p256dh: json.keys?.p256dh, auth: json.keys?.auth },
        user_agent: navigator.userAgent.slice(0, 300),
      });
      setSubscribed(true);
      showToast({ title: "Notificações ativadas", tone: "success" });
    } catch (err) {
      showToast({
        title: "Não foi possível ativar notificações",
        description: err instanceof ApiError ? err.message : "Erro inesperado",
        tone: "danger",
      });
    } finally {
      setBusy(false);
    }
  }

  async function disable() {
    if (busy) return;
    setBusy(true);
    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();
      if (subscription) {
        const endpoint = subscription.endpoint;
        await subscription.unsubscribe();
        await apiClient.post("v1/egress/push/unsubscribe", { endpoint });
      }
      setSubscribed(false);
      showToast({ title: "Notificações desativadas", tone: "success" });
    } catch (err) {
      showToast({
        title: "Não foi possível desativar notificações",
        description: err instanceof ApiError ? err.message : "Erro inesperado",
        tone: "danger",
      });
    } finally {
      setBusy(false);
    }
  }

  // Módulo desligado, navegador sem suporte (ex.: Safari/Firefox mais
  // antigos), ou operador não configurou VAPID_PUBLIC_KEY/
  // VAPID_PRIVATE_KEY neste ambiente — a seção some por completo em vez
  // de mostrar um botão que sempre falharia.
  if (!egressEnabled || checking) return null;
  if (!supported || !vapid?.enabled) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle as="h2">Notificações push</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-3">
          <p className="text-sm text-muted">
            {subscribed
              ? "Este navegador está inscrito para receber notificações da Assistência Social."
              : "Ative para receber notificações da Assistência Social mesmo com a aba fechada."}
          </p>
          <div>
            {subscribed ? (
              <Button type="button" variant="secondary" size="sm" loading={busy} onClick={() => void disable()}>
                Desativar notificações
              </Button>
            ) : (
              <Button type="button" size="sm" loading={busy} onClick={() => void enable()}>
                Ativar notificações
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// urlBase64ToUint8Array — PushManager.subscribe() exige
// applicationServerKey como Uint8Array, mas o backend expõe a chave VAPID
// em base64url (mesmo formato que webpush-go/GenerateVAPIDKeys produz).
// Conversão padrão da própria documentação do Web Push (MDN) — sem
// nenhuma lib extra para uma função de ~10 linhas.
function urlBase64ToUint8Array(base64url: string): BufferSource {
  const padding = "=".repeat((4 - (base64url.length % 4)) % 4);
  const base64 = (base64url + padding).replace(/-/g, "+").replace(/_/g, "/");
  const raw = atob(base64);
  const out = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
  // O DOM lib mais recente distingue Uint8Array<ArrayBuffer> de
  // Uint8Array<ArrayBufferLike> (SharedArrayBuffer incluso) de um jeito
  // que o construtor comum não satisfaz sem fricção — um array recém
  // alocado por `new Uint8Array(n)` nunca é compartilhado, então o cast é
  // seguro (não é `any`, só resolve uma distinção que TS não infere aqui).
  return out as BufferSource;
}
