// Mesma regra de curingas do backend (internal/platform/auth/rbac.go):
// "*" concede tudo, "recurso:*" concede toda ação do recurso. No frontend
// isto só decide o que MOSTRAR — a autorização efetiva é sempre do
// backend (middleware RequirePermission, A01).
export function hasPermission(granted: readonly string[] | undefined, want: string): boolean {
  if (!granted) return false;
  const [resource] = want.split(":");
  return granted.some((g) => g === "*" || g === want || g === `${resource}:*`);
}
