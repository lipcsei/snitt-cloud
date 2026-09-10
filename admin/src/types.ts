/** A pénzösszegek mindenhol a pénznem legkisebb egységében, egész számként. */
export type Minor = number;

export type Plan = {
  key: string;
  name: string;
  description: string;
  price_minor: Minor;
  feature_keys: string[];
};

export type SubscriptionStatus = 'active' | 'past_due' | 'canceled' | 'expired';
export type InvoiceStatus = 'open' | 'paid' | 'void';

export type Subscription = {
  id: string;
  subject: string;
  plan: string;
  status: SubscriptionStatus;
  current_period_start: string;
  current_period_end: string;
  price_minor: Minor;
  currency: string;
  cancel_at_period_end: boolean;
  canceled_at: string | null;
  provider: string;
  external_customer_id: string;
  external_subscription_id: string;
  created_at: string;
  updated_at: string;
};

export type UserRef = {
  subject: string;
  email: string;
  display_name: string;
};

export type SubscriptionView = Subscription & { user: UserRef };

export type Invoice = {
  id: string;
  number: string;
  subject: string;
  subscription_id: string | null;
  amount_minor: Minor;
  currency: string;
  status: InvoiceStatus;
  issued_at: string;
  due_at: string | null;
  paid_at: string | null;
  provider: string;
  external_invoice_id: string;
  note: string;
  created_at: string;
  updated_at: string;
};

export type InvoiceView = Invoice & { user: UserRef };

export type InvoiceTotals = {
  count: number;
  total_minor: Minor;
  paid_minor: Minor;
  open_minor: Minor;
};

export type Entitlement = {
  feature_key: string;
  granted_at: string;
  expires_at: string | null;
};

export type AdminUser = {
  subject: string;
  email: string;
  username: string;
  full_name: string;
  enabled: boolean;
  email_verified: boolean;
  registered_at: string | null;
  has_profile: boolean;
  display_name: string;
  locale: string;
  profile_since: string | null;
  subscription: Subscription | null;
};

export type UsersResponse = {
  users: AdminUser[];
  page: number;
  page_size: number;
  total: number;
};

export type UserDetail = {
  user: AdminUser;
  entitlements: Entitlement[];
  subscriptions: Subscription[];
  invoices: Invoice[];
  invoice_totals: InvoiceTotals;
};

export type SubscriptionsResponse = {
  subscriptions: SubscriptionView[];
  page: number;
  page_size: number;
  total: number;
};

export type InvoicesResponse = {
  invoices: InvoiceView[];
  totals: InvoiceTotals;
  page: number;
  page_size: number;
  currency: string;
};

export type PlanCount = { plan: string; count: number };

export type Overview = {
  user_count: number | null;
  profile_count: number;
  active_subscription_count: number;
  active_by_plan: PlanCount[];
  mrr_minor: Minor;
  new_profiles_30d: number;
  new_subscriptions_30d: number;
  unpaid_invoice_count: number;
  unpaid_invoice_total_minor: Minor;
  currency: string;
  plans: Plan[];
  warnings: string[];
};

export type AuditAction =
  | 'subscription.grant'
  | 'subscription.change_plan'
  | 'subscription.cancel'
  | 'subscription.reactivate'
  | 'invoice.create'
  | 'invoice.pay'
  | 'invoice.void'
  | 'entitlement.grant'
  | 'entitlement.revoke'
  | 'user.enable'
  | 'user.disable';

export type AuditEntry = {
  id: string;
  at: string;
  actor_subject: string;
  actor_label: string;
  action: AuditAction | string;
  target_type: string;
  target_id: string;
  subject: string;
  summary: string;
  detail: Record<string, unknown>;
};

export type AuditResponse = {
  entries: AuditEntry[];
  page: number;
  page_size: number;
  total: number;
  /** A szerver adja a szűrő legördülőjének a lehetséges műveleteit. */
  actions: string[];
};
