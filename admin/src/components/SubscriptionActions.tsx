import { useState, type FormEvent } from 'react';
import { api } from '../api';
import { formatDate, formatMoney, fromMinor, toMinor } from '../format';
import type { Plan, Subscription } from '../types';
import { Modal } from './Modal';
import { ErrorBox } from './States';

type Done = (message: string) => void;

function useAction(onDone: Done, onClose: () => void) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);

  const run = async (fn: () => Promise<void>, message: string) => {
    setBusy(true);
    setError(null);
    try {
      await fn();
      onClose();
      onDone(message);
    } catch (err) {
      // A hiba a dialógusban marad, hogy az admin lássa, mi nem sikerült,
      // és a beírt adatok se vesszenek el.
      setError(err);
    } finally {
      setBusy(false);
    }
  };

  return { busy, error, run };
}

function PlanPicker({
  plans,
  value,
  onChange,
  currency,
}: {
  plans: Plan[];
  value: string;
  onChange: (v: string) => void;
  currency: string;
}) {
  return (
    <div className="plan-picker">
      {plans.map((p) => (
        <label key={p.key} className={`plan-option ${value === p.key ? 'is-selected' : ''}`}>
          <input
            type="radio"
            name="plan"
            value={p.key}
            checked={value === p.key}
            onChange={() => onChange(p.key)}
          />
          <span className="plan-option__body">
            <span className="plan-option__title">
              <strong>{p.name}</strong>
              <span className="muted">{formatMoney(p.price_minor, currency)} / hó</span>
            </span>
            <span className="muted">{p.description}</span>
            <span className="plan-option__features">
              {p.feature_keys.map((f) => (
                <code key={f}>{f}</code>
              ))}
            </span>
          </span>
        </label>
      ))}
    </div>
  );
}

export function GrantDialog({
  subject,
  subjectLabel,
  plans,
  currency,
  onClose,
  onDone,
}: {
  subject: string;
  subjectLabel: string;
  plans: Plan[];
  currency: string;
  onClose: () => void;
  onDone: Done;
}) {
  const [plan, setPlan] = useState(plans[0]?.key ?? '');
  const [price, setPrice] = useState(fromMinor(plans[0]?.price_minor ?? 0));
  const [periodDays, setPeriodDays] = useState('30');
  const [createInvoice, setCreateInvoice] = useState(true);
  const { busy, error, run } = useAction(onDone, onClose);

  const selectPlan = (key: string) => {
    setPlan(key);
    const p = plans.find((x) => x.key === key);
    if (p) setPrice(fromMinor(p.price_minor));
  };

  const priceMinor = toMinor(price);
  const days = Number(periodDays);
  const invalid = !plan || priceMinor === null || !Number.isInteger(days) || days < 1;

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (invalid || priceMinor === null) return;
    void run(async () => {
      const res = await api.grantSubscription({
        subject,
        plan,
        period_days: days,
        price_minor: priceMinor,
        currency,
        create_invoice: createInvoice,
      });
      if (res.warning) throw new Error(res.warning);
    }, `Előfizetés kiadva: ${plan}.`);
  };

  return (
    <Modal
      title="Előfizetés kiadása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button type="submit" form="grant-form" className="btn btn--primary" disabled={busy || invalid}>
            {busy ? 'Mentés…' : 'Előfizetés kiadása'}
          </button>
        </>
      }
    >
      <form id="grant-form" onSubmit={submit} className="form">
        <p className="muted">
          Felhasználó: <strong>{subjectLabel}</strong>
        </p>
        <PlanPicker plans={plans} value={plan} onChange={selectPlan} currency={currency} />

        <div className="form__row">
          <label className="field">
            <span>Ár ({currency})</span>
            <input
              type="text"
              inputMode="decimal"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
            {priceMinor === null ? <span className="field__err">Érvénytelen összeg.</span> : null}
          </label>
          <label className="field">
            <span>Időszak (nap)</span>
            <input
              type="number"
              min={1}
              max={3650}
              value={periodDays}
              onChange={(e) => setPeriodDays(e.target.value)}
            />
          </label>
        </div>

        <label className="check">
          <input
            type="checkbox"
            checked={createInvoice}
            onChange={(e) => setCreateInvoice(e.target.checked)}
          />
          <span>
            Nyitott számla kiállítása is (14 napos határidővel). Terhelés nem történik – a
            beérkezést kézzel kell jelölni.
          </span>
        </label>

        {error ? <ErrorBox error={error} /> : null}
      </form>
    </Modal>
  );
}

export function ChangePlanDialog({
  subscription,
  plans,
  currency,
  onClose,
  onDone,
}: {
  subscription: Subscription;
  plans: Plan[];
  currency: string;
  onClose: () => void;
  onDone: Done;
}) {
  const [plan, setPlan] = useState(subscription.plan);
  const [price, setPrice] = useState(fromMinor(subscription.price_minor));
  const [restart, setRestart] = useState(false);
  const [periodDays, setPeriodDays] = useState('30');
  const { busy, error, run } = useAction(onDone, onClose);

  const selectPlan = (key: string) => {
    setPlan(key);
    const p = plans.find((x) => x.key === key);
    if (p) setPrice(fromMinor(p.price_minor));
  };

  const priceMinor = toMinor(price);
  const days = Number(periodDays);
  const invalid = !plan || priceMinor === null || (restart && (!Number.isInteger(days) || days < 1));

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (invalid || priceMinor === null) return;
    void run(
      () =>
        api
          .changePlan(subscription.id, {
            plan,
            price_minor: priceMinor,
            ...(restart ? { period_days: days } : {}),
          })
          .then(() => undefined),
      `Csomag módosítva: ${plan}.`,
    );
  };

  return (
    <Modal
      title="Csomag módosítása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="submit"
            form="change-form"
            className="btn btn--primary"
            disabled={busy || invalid}
          >
            {busy ? 'Mentés…' : 'Módosítás'}
          </button>
        </>
      }
    >
      <form id="change-form" onSubmit={submit} className="form">
        <p className="muted">
          Jelenlegi időszak vége: <strong>{formatDate(subscription.current_period_end)}</strong>
        </p>
        <PlanPicker plans={plans} value={plan} onChange={selectPlan} currency={currency} />

        <div className="form__row">
          <label className="field">
            <span>Ár ({currency})</span>
            <input
              type="text"
              inputMode="decimal"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
            {priceMinor === null ? <span className="field__err">Érvénytelen összeg.</span> : null}
          </label>
          <label className="field">
            <span>Új időszak hossza (nap)</span>
            <input
              type="number"
              min={1}
              max={3650}
              value={periodDays}
              disabled={!restart}
              onChange={(e) => setPeriodDays(e.target.value)}
            />
          </label>
        </div>

        <label className="check">
          <input type="checkbox" checked={restart} onChange={(e) => setRestart(e.target.checked)} />
          <span>Új elszámolási időszak indítása mostantól</span>
        </label>

        <p className="hint">
          A csomaghoz tartozó jogosultságok azonnal átállnak, lejáratuk az időszak vége.
        </p>

        {error ? <ErrorBox error={error} /> : null}
      </form>
    </Modal>
  );
}

export function CancelDialog({
  subscription,
  onClose,
  onDone,
}: {
  subscription: Subscription;
  onClose: () => void;
  onDone: Done;
}) {
  const [atPeriodEnd, setAtPeriodEnd] = useState(true);
  const { busy, error, run } = useAction(onDone, onClose);

  const confirm = () =>
    void run(
      () => api.cancelSubscription(subscription.id, atPeriodEnd).then(() => undefined),
      atPeriodEnd ? 'Az előfizetés az időszak végén lejár.' : 'Az előfizetés azonnal lemondva.',
    );

  return (
    <Modal
      title="Előfizetés lemondása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button type="button" className="btn btn--danger" onClick={confirm} disabled={busy}>
            {busy ? 'Lemondás…' : 'Lemondás megerősítése'}
          </button>
        </>
      }
    >
      <div className="form">
        <p>
          Csomag: <strong>{subscription.plan}</strong> · Időszak vége:{' '}
          <strong>{formatDate(subscription.current_period_end)}</strong>
        </p>

        <label className="check">
          <input type="radio" checked={atPeriodEnd} onChange={() => setAtPeriodEnd(true)} />
          <span>
            <strong>Az időszak végén</strong> – a hozzáférés {formatDate(subscription.current_period_end)}-ig
            megmarad.
          </span>
        </label>
        <label className="check">
          <input type="radio" checked={!atPeriodEnd} onChange={() => setAtPeriodEnd(false)} />
          <span>
            <strong>Azonnal</strong> – az AI funkciók jogosultsága rögtön elvész.
          </span>
        </label>

        <p className="hint">
          Kifizetett számlákat ez nem érint, és pénzvisszatérítés sem történik automatikusan.
        </p>

        {error ? <ErrorBox error={error} /> : null}
      </div>
    </Modal>
  );
}

export function ReactivateDialog({
  subscription,
  onClose,
  onDone,
}: {
  subscription: Subscription;
  onClose: () => void;
  onDone: Done;
}) {
  const expired = new Date(subscription.current_period_end).getTime() <= Date.now();
  const [newPeriod, setNewPeriod] = useState(expired);
  const [periodDays, setPeriodDays] = useState('30');
  const { busy, error, run } = useAction(onDone, onClose);

  const days = Number(periodDays);
  const invalid = newPeriod && (!Number.isInteger(days) || days < 1);

  const confirm = () =>
    void run(
      () =>
        api
          .reactivateSubscription(subscription.id, newPeriod ? days : undefined)
          .then(() => undefined),
      'Az előfizetés újra aktív.',
    );

  return (
    <Modal
      title="Előfizetés visszakapcsolása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="button"
            className="btn btn--primary"
            onClick={confirm}
            disabled={busy || invalid}
          >
            {busy ? 'Mentés…' : 'Visszakapcsolás'}
          </button>
        </>
      }
    >
      <div className="form">
        <p>
          Csomag: <strong>{subscription.plan}</strong> · Korábbi időszak vége:{' '}
          <strong>{formatDate(subscription.current_period_end)}</strong>
        </p>

        {expired ? (
          <p className="hint">
            A korábbi időszak már lejárt, ezért új időszak indul – enélkül azonnal lejárt
            jogosultságot adnánk vissza.
          </p>
        ) : (
          <label className="check">
            <input
              type="checkbox"
              checked={newPeriod}
              onChange={(e) => setNewPeriod(e.target.checked)}
            />
            <span>Új elszámolási időszak indítása mostantól</span>
          </label>
        )}

        {newPeriod ? (
          <label className="field">
            <span>Időszak (nap)</span>
            <input
              type="number"
              min={1}
              max={3650}
              value={periodDays}
              onChange={(e) => setPeriodDays(e.target.value)}
            />
          </label>
        ) : null}

        {error ? <ErrorBox error={error} /> : null}
      </div>
    </Modal>
  );
}

/** Az előfizetéshez tartozó műveletek gombsora, a státusztól függően. */
export function SubscriptionActions({
  subscription,
  plans,
  currency,
  onDone,
  compact,
}: {
  subscription: Subscription;
  plans: Plan[];
  currency: string;
  onDone: Done;
  compact?: boolean;
}) {
  const [dialog, setDialog] = useState<'change' | 'cancel' | 'reactivate' | null>(null);
  const live = subscription.status === 'active' || subscription.status === 'past_due';

  return (
    <div className={`actions ${compact ? 'actions--compact' : ''}`}>
      {live ? (
        <>
          <button type="button" className="btn btn--ghost" onClick={() => setDialog('change')}>
            Csomag
          </button>
          <button type="button" className="btn btn--danger-ghost" onClick={() => setDialog('cancel')}>
            Lemondás
          </button>
        </>
      ) : null}
      {!live || subscription.cancel_at_period_end ? (
        <button type="button" className="btn btn--ghost" onClick={() => setDialog('reactivate')}>
          Visszakapcsolás
        </button>
      ) : null}

      {dialog === 'change' ? (
        <ChangePlanDialog
          subscription={subscription}
          plans={plans}
          currency={currency}
          onClose={() => setDialog(null)}
          onDone={onDone}
        />
      ) : null}
      {dialog === 'cancel' ? (
        <CancelDialog
          subscription={subscription}
          onClose={() => setDialog(null)}
          onDone={onDone}
        />
      ) : null}
      {dialog === 'reactivate' ? (
        <ReactivateDialog
          subscription={subscription}
          onClose={() => setDialog(null)}
          onDone={onDone}
        />
      ) : null}
    </div>
  );
}
