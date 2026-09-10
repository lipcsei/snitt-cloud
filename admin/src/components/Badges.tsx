import { INVOICE_STATUS_LABELS, SUBSCRIPTION_STATUS_LABELS } from '../format';
import type { InvoiceStatus, SubscriptionStatus } from '../types';

const subTone: Record<SubscriptionStatus, string> = {
  active: 'ok',
  past_due: 'warn',
  canceled: 'danger',
  expired: 'muted',
};

const invoiceTone: Record<InvoiceStatus, string> = {
  open: 'warn',
  paid: 'ok',
  void: 'muted',
};

export function SubscriptionBadge({
  status,
  cancelAtPeriodEnd,
}: {
  status: SubscriptionStatus;
  cancelAtPeriodEnd?: boolean;
}) {
  return (
    <span className="badges">
      <span className={`pill pill--${subTone[status] ?? 'muted'}`}>
        {SUBSCRIPTION_STATUS_LABELS[status] ?? status}
      </span>
      {cancelAtPeriodEnd && status === 'active' ? (
        <span className="pill pill--warn">Időszak végén lejár</span>
      ) : null}
    </span>
  );
}

export function InvoiceBadge({ status }: { status: InvoiceStatus }) {
  return (
    <span className={`pill pill--${invoiceTone[status] ?? 'muted'}`}>
      {INVOICE_STATUS_LABELS[status] ?? status}
    </span>
  );
}

export function PlanBadge({ plan }: { plan: string }) {
  return <span className="pill pill--plan">{plan}</span>;
}
