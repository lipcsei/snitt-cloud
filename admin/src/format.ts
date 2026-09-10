import type { InvoiceStatus, SubscriptionStatus } from './types';

const dateTime = new Intl.DateTimeFormat('hu-HU', {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
});

const dateOnly = new Intl.DateTimeFormat('hu-HU', {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});

export function formatDateTime(iso: string | null | undefined): string {
  if (!iso) return '–';
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? '–' : dateTime.format(d);
}

export function formatDate(iso: string | null | undefined): string {
  if (!iso) return '–';
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? '–' : dateOnly.format(d);
}

/**
 * A szerver a pénzt a pénznem legkisebb egységében küldi. A HUF-nak
 * gyakorlatban nincs váltópénze, de a tárolás egységes, ezért itt osztunk.
 */
export function formatMoney(minor: number, currency: string): string {
  try {
    return new Intl.NumberFormat('hu-HU', {
      style: 'currency',
      currency,
      maximumFractionDigits: currency === 'HUF' ? 0 : 2,
    }).format(minor / 100);
  } catch {
    // Ismeretlen ISO kód esetén ne dőljön el a tábla.
    return `${(minor / 100).toLocaleString('hu-HU')} ${currency}`;
  }
}

/** A minor egységbe konvertál egy beírt összeget (pl. "2990" -> 299000). */
export function toMinor(value: string): number | null {
  const normalized = value.replace(/\s/g, '').replace(',', '.');
  if (normalized === '') return null;
  const n = Number(normalized);
  if (!Number.isFinite(n) || n < 0) return null;
  return Math.round(n * 100);
}

export function fromMinor(minor: number): string {
  return String(minor / 100);
}

export const SUBSCRIPTION_STATUS_LABELS: Record<SubscriptionStatus, string> = {
  active: 'Aktív',
  past_due: 'Fizetési késedelem',
  canceled: 'Lemondva',
  expired: 'Lejárt',
};

export const INVOICE_STATUS_LABELS: Record<InvoiceStatus, string> = {
  open: 'Nyitott',
  paid: 'Kifizetve',
  void: 'Sztornó',
};

export function daysUntil(iso: string): number {
  return Math.ceil((new Date(iso).getTime() - Date.now()) / 86_400_000);
}

/**
 * A napló műveleteinek magyar címkéi. A kulcs a szerver gépi neve; ismeretlen
 * kulcsnál a nyers érték látszik, nem üres cella - a napló akkor is olvasható
 * marad, ha a szerver új műveletet vezet be a felület előtt.
 */
export const AUDIT_ACTION_LABELS: Record<string, string> = {
  'subscription.grant': 'Előfizetés kiadva',
  'subscription.change_plan': 'Csomagváltás',
  'subscription.cancel': 'Előfizetés lemondva',
  'subscription.reactivate': 'Előfizetés visszakapcsolva',
  'invoice.create': 'Számla kiállítva',
  'invoice.pay': 'Számla kifizetve',
  'invoice.void': 'Számla sztornózva',
};

export function auditActionLabel(action: string): string {
  return AUDIT_ACTION_LABELS[action] ?? action;
}
