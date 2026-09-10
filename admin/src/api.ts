import { API_BASE_URL, keycloak } from './auth';
import type {
  AuditResponse,
  Entitlement,
  InvoicesResponse,
  InvoiceView,
  Overview,
  Plan,
  Subscription,
  SubscriptionsResponse,
  SubscriptionView,
  UserDetail,
  UsersResponse,
} from './types';

/**
 * ApiError a szerver által küldött üzenetet is hordozza. Pénzt érintő
 * műveleteknél nem elég a "valami hiba történt": a felület a szerver saját
 * indoklását mutatja meg (pl. "csak nyitott számla sztornózható").
 */
export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  // A token minden hívás előtt frissül, ha 30 másodpercen belül lejárna:
  // az admin sokáig nyitva hagyja a felületet.
  try {
    await keycloak.updateToken(30);
  } catch {
    throw new ApiError(401, 'A munkamenet lejárt, jelentkezz be újra.');
  }

  let resp: Response;
  try {
    resp = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers: {
        Authorization: `Bearer ${keycloak.token ?? ''}`,
        ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
        ...init?.headers,
      },
    });
  } catch {
    throw new ApiError(0, `Az account szolgáltatás nem érhető el (${API_BASE_URL}).`);
  }

  if (resp.status === 204) return undefined as T;

  const text = await resp.text();
  let body: unknown = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = null;
    }
  }

  if (!resp.ok) {
    const message =
      body && typeof body === 'object' && typeof (body as { error?: unknown }).error === 'string'
        ? (body as { error: string }).error
        : `Ismeretlen szerverhiba (HTTP ${resp.status}).`;
    throw new ApiError(resp.status, message);
  }
  return body as T;
}

function query(params: Record<string, string | number | undefined>): string {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== '') q.set(k, String(v));
  }
  const s = q.toString();
  return s ? `?${s}` : '';
}

const post = (body?: unknown): RequestInit => ({
  method: 'POST',
  body: JSON.stringify(body ?? {}),
});

export const api = {
  overview: () => request<Overview>('/api/v1/admin/overview'),

  plans: () => request<{ plans: Plan[]; currency: string }>('/api/v1/admin/plans'),

  users: (params: { q?: string; page?: number; page_size?: number }) =>
    request<UsersResponse>(`/api/v1/admin/users${query(params)}`),

  user: (subject: string) =>
    request<UserDetail>(`/api/v1/admin/users/${encodeURIComponent(subject)}`),

  subscriptions: (params: {
    plan?: string;
    status?: string;
    subject?: string;
    page?: number;
    page_size?: number;
  }) => request<SubscriptionsResponse>(`/api/v1/admin/subscriptions${query(params)}`),

  subscription: (id: string) =>
    request<SubscriptionView>(`/api/v1/admin/subscriptions/${encodeURIComponent(id)}`),

  grantSubscription: (body: {
    subject: string;
    plan: string;
    period_days?: number;
    price_minor?: number;
    currency?: string;
    create_invoice: boolean;
  }) =>
    request<{ subscription: Subscription; warning?: string }>(
      '/api/v1/admin/subscriptions',
      post(body),
    ),

  changePlan: (id: string, body: { plan: string; price_minor?: number; period_days?: number }) =>
    request<{ subscription: Subscription }>(
      `/api/v1/admin/subscriptions/${encodeURIComponent(id)}`,
      { method: 'PATCH', body: JSON.stringify(body) },
    ),

  cancelSubscription: (id: string, atPeriodEnd: boolean) =>
    request<{ subscription: Subscription }>(
      `/api/v1/admin/subscriptions/${encodeURIComponent(id)}/cancel`,
      post({ at_period_end: atPeriodEnd }),
    ),

  reactivateSubscription: (id: string, periodDays?: number) =>
    request<{ subscription: Subscription }>(
      `/api/v1/admin/subscriptions/${encodeURIComponent(id)}/reactivate`,
      post(periodDays ? { period_days: periodDays } : {}),
    ),

  invoices: (params: {
    status?: string;
    subject?: string;
    page?: number;
    page_size?: number;
  }) => request<InvoicesResponse>(`/api/v1/admin/invoices${query(params)}`),

  invoice: (id: string) => request<InvoiceView>(`/api/v1/admin/invoices/${encodeURIComponent(id)}`),

  createInvoice: (body: {
    subject: string;
    subscription_id?: string | null;
    amount_minor: number;
    currency?: string;
    due_days?: number;
    note?: string;
  }) => request<{ invoice: InvoiceView }>('/api/v1/admin/invoices', post(body)),

  payInvoice: (id: string, externalInvoiceId?: string) =>
    request<{ invoice: InvoiceView }>(
      `/api/v1/admin/invoices/${encodeURIComponent(id)}/pay`,
      post(externalInvoiceId ? { external_invoice_id: externalInvoiceId } : {}),
    ),

  voidInvoice: (id: string, note?: string) =>
    request<{ invoice: InvoiceView }>(
      `/api/v1/admin/invoices/${encodeURIComponent(id)}/void`,
      post(note ? { note } : {}),
    ),

  grantEntitlement: (
    subject: string,
    body: { feature_key: string; days?: number; note?: string },
  ) =>
    request<{ entitlement: Entitlement; warning?: string }>(
      `/api/v1/admin/users/${encodeURIComponent(subject)}/entitlements`,
      post(body),
    ),

  revokeEntitlement: (subject: string, featureKey: string) =>
    request<{ entitlement: Entitlement }>(
      `/api/v1/admin/users/${encodeURIComponent(subject)}/entitlements/${encodeURIComponent(featureKey)}`,
      { method: 'DELETE' },
    ),

  setUserEnabled: (subject: string, enabled: boolean, note?: string) =>
    request<{ enabled: boolean; changed: boolean }>(
      `/api/v1/admin/users/${encodeURIComponent(subject)}`,
      { method: 'PATCH', body: JSON.stringify(note ? { enabled, note } : { enabled }) },
    ),

  audit: (params: {
    action?: string;
    subject?: string;
    actor?: string;
    from?: string;
    to?: string;
    page?: number;
    page_size?: number;
  }) => request<AuditResponse>(`/api/v1/admin/audit${query(params)}`),
};
