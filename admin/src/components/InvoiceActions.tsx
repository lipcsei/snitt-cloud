import { useState, type FormEvent } from 'react';
import { api } from '../api';
import { formatDate, formatMoney, toMinor } from '../format';
import type { Invoice } from '../types';
import { Modal } from './Modal';
import { useAction, type Done } from './useAction';
import { ErrorBox } from './States';


export function PayDialog({
  invoice,
  onClose,
  onDone,
}: {
  invoice: Invoice;
  onClose: () => void;
  onDone: Done;
}) {
  const [externalId, setExternalId] = useState('');
  const { busy, error, run } = useAction(onDone, onClose);

  return (
    <Modal
      title="Számla kifizetettre jelölése"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="button"
            className="btn btn--primary"
            disabled={busy}
            onClick={() =>
              void run(
                () => api.payInvoice(invoice.id, externalId.trim() || undefined),
                `${invoice.number} kifizetettre állítva.`,
              )
            }
          >
            {busy ? 'Mentés…' : 'Kifizetettre jelölés'}
          </button>
        </>
      }
    >
      <div className="form">
        <p>
          <strong>{invoice.number}</strong> ·{' '}
          {formatMoney(invoice.amount_minor, invoice.currency)} · kiállítva:{' '}
          {formatDate(invoice.issued_at)}
        </p>
        <p className="hint">
          Ez csak könyvelési jelölés: a rendszer nem terhel semmit, a pénznek máshol kellett
          beérkeznie. A művelet nem vonható vissza.
        </p>
        <label className="field">
          <span>Külső tranzakcióazonosító (opcionális)</span>
          <input
            type="text"
            value={externalId}
            placeholder="pl. banki utalás azonosítója"
            onChange={(e) => setExternalId(e.target.value)}
          />
        </label>
        {error ? <ErrorBox error={error} /> : null}
      </div>
    </Modal>
  );
}

export function VoidDialog({
  invoice,
  onClose,
  onDone,
}: {
  invoice: Invoice;
  onClose: () => void;
  onDone: Done;
}) {
  const [note, setNote] = useState('');
  const { busy, error, run } = useAction(onDone, onClose);

  return (
    <Modal
      title="Számla sztornózása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="button"
            className="btn btn--danger"
            disabled={busy || note.trim() === ''}
            onClick={() =>
              void run(() => api.voidInvoice(invoice.id, note.trim()), `${invoice.number} sztornózva.`)
            }
          >
            {busy ? 'Sztornó…' : 'Sztornó megerősítése'}
          </button>
        </>
      }
    >
      <div className="form">
        <p>
          <strong>{invoice.number}</strong> ·{' '}
          {formatMoney(invoice.amount_minor, invoice.currency)}
        </p>
        <p className="hint">
          Sztornózni csak nyitott számlát lehet, és a művelet nem vonható vissza. Kifizetett
          számlát csak jóváíró számlával lehet visszavonni.
        </p>
        <label className="field">
          <span>Indoklás (kötelező)</span>
          <input
            type="text"
            value={note}
            placeholder="pl. téves kiállítás"
            onChange={(e) => setNote(e.target.value)}
          />
        </label>
        {error ? <ErrorBox error={error} /> : null}
      </div>
    </Modal>
  );
}

export function CreateInvoiceDialog({
  subject,
  subjectLabel,
  subscriptionId,
  currency,
  onClose,
  onDone,
}: {
  subject: string;
  subjectLabel: string;
  subscriptionId?: string | null;
  currency: string;
  onClose: () => void;
  onDone: Done;
}) {
  const [amount, setAmount] = useState('');
  const [dueDays, setDueDays] = useState('14');
  const [note, setNote] = useState('');
  const { busy, error, run } = useAction(onDone, onClose);

  const amountMinor = toMinor(amount);
  const days = Number(dueDays);
  const invalid =
    amountMinor === null || amountMinor <= 0 || !Number.isInteger(days) || days < 0 || days > 365;

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (invalid || amountMinor === null) return;
    void run(
      () =>
        api.createInvoice({
          subject,
          subscription_id: subscriptionId ?? null,
          amount_minor: amountMinor,
          currency,
          due_days: days,
          note: note.trim(),
        }),
      'Nyitott számla kiállítva.',
    );
  };

  return (
    <Modal
      title="Számla kiállítása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="submit"
            form="invoice-form"
            className="btn btn--primary"
            disabled={busy || invalid}
          >
            {busy ? 'Mentés…' : 'Kiállítás'}
          </button>
        </>
      }
    >
      <form id="invoice-form" onSubmit={submit} className="form">
        <p className="muted">
          Felhasználó: <strong>{subjectLabel}</strong>
        </p>
        <div className="form__row">
          <label className="field">
            <span>Összeg ({currency})</span>
            <input
              type="text"
              inputMode="decimal"
              value={amount}
              placeholder="2990"
              onChange={(e) => setAmount(e.target.value)}
            />
            {amount !== '' && (amountMinor === null || amountMinor <= 0) ? (
              <span className="field__err">Pozitív összeget adj meg.</span>
            ) : null}
          </label>
          <label className="field">
            <span>Fizetési határidő (nap)</span>
            <input
              type="number"
              min={0}
              max={365}
              value={dueDays}
              onChange={(e) => setDueDays(e.target.value)}
            />
          </label>
        </div>
        <label className="field">
          <span>Megjegyzés</span>
          <input type="text" value={note} onChange={(e) => setNote(e.target.value)} />
        </label>
        <p className="hint">
          A számla nyitott állapotban jön létre, terhelés nem történik.
        </p>
        {error ? <ErrorBox error={error} /> : null}
      </form>
    </Modal>
  );
}

export function InvoiceActions({
  invoice,
  onDone,
}: {
  invoice: Invoice;
  onDone: Done;
}) {
  const [dialog, setDialog] = useState<'pay' | 'void' | null>(null);
  if (invoice.status !== 'open') {
    return <span className="muted">–</span>;
  }

  return (
    <div className="actions actions--compact">
      <button type="button" className="btn btn--ghost" onClick={() => setDialog('pay')}>
        Kifizetve
      </button>
      <button type="button" className="btn btn--danger-ghost" onClick={() => setDialog('void')}>
        Sztornó
      </button>

      {dialog === 'pay' ? (
        <PayDialog invoice={invoice} onClose={() => setDialog(null)} onDone={onDone} />
      ) : null}
      {dialog === 'void' ? (
        <VoidDialog invoice={invoice} onClose={() => setDialog(null)} onDone={onDone} />
      ) : null}
    </div>
  );
}
