import { useState, type FormEvent } from 'react';
import { api } from '../api';
import { Modal } from './Modal';
import { ErrorBox } from './States';
import { useAction, type Done } from './useAction';

/**
 * Kézi jogosultság kiadása. A kulcs szabad szöveg: a csomagokon kívüli
 * kulcsok (béta-hozzáférés, egyedi megállapodás) éppen ettől lehetségesek.
 * A csomaghoz tartozó kulcsokra a szerver figyelmeztet, mert azokat a
 * következő előfizetés-művelet felülírja.
 */
export function GrantEntitlementDialog({
  subject,
  subjectLabel,
  managedKeys,
  onClose,
  onDone,
}: {
  subject: string;
  subjectLabel: string;
  managedKeys: string[];
  onClose: () => void;
  onDone: Done;
}) {
  const [key, setKey] = useState('');
  const [forever, setForever] = useState(true);
  const [days, setDays] = useState('30');
  const [note, setNote] = useState('');
  const { busy, error, run } = useAction(onDone, onClose);

  const trimmed = key.trim();
  const validKey = /^[a-z0-9]([a-z0-9._-]{0,62}[a-z0-9])?$/.test(trimmed);
  const dayCount = Number(days);
  const validDays = forever || (Number.isInteger(dayCount) && dayCount >= 1 && dayCount <= 3650);
  const managed = managedKeys.includes(trimmed);

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (!validKey || !validDays) return;
    void run(
      () =>
        api.grantEntitlement(subject, {
          feature_key: trimmed,
          days: forever ? undefined : dayCount,
          note: note.trim() || undefined,
        }),
      `Jogosultság kiadva: ${trimmed}.`,
    );
  };

  return (
    <Modal
      title="Jogosultság kiadása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="submit"
            form="grant-entitlement-form"
            className="btn btn--primary"
            disabled={busy || !validKey || !validDays}
          >
            {busy ? 'Mentés…' : 'Kiadás'}
          </button>
        </>
      }
    >
      <form id="grant-entitlement-form" onSubmit={submit} className="form">
        <p className="muted">
          Felhasználó: <strong>{subjectLabel}</strong>
        </p>

        <label className="field">
          <span>Jogosultság kulcsa</span>
          <input
            type="text"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            placeholder="beta-access"
            autoFocus
          />
          {trimmed !== '' && !validKey ? (
            <span className="field__err">
              Csak kisbetű, szám, pont, kötőjel és aláhúzás; betűvel vagy számmal kezdődik.
            </span>
          ) : null}
        </label>

        {managed ? (
          <div className="banner banner--warn" role="status">
            <span>
              Ez a kulcs csomaghoz tartozik. A felhasználó következő előfizetés-műveletekor
              (kiadás, csomagváltás, lemondás) <strong>felülíródik</strong> – tartós hozzáféréshez
              inkább előfizetést adj ki.
            </span>
          </div>
        ) : null}

        <label className="check">
          <input type="checkbox" checked={forever} onChange={(e) => setForever(e.target.checked)} />
          <span>Ne járjon le</span>
        </label>

        {!forever ? (
          <label className="field">
            <span>Futamidő (nap)</span>
            <input
              type="number"
              min={1}
              max={3650}
              value={days}
              onChange={(e) => setDays(e.target.value)}
            />
          </label>
        ) : null}

        <label className="field">
          <span>Megjegyzés (a naplóba kerül)</span>
          <input
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="pl. egyedi megállapodás"
          />
        </label>

        {error ? <ErrorBox error={error} /> : null}
      </form>
    </Modal>
  );
}

export function RevokeEntitlementDialog({
  subject,
  featureKey,
  onClose,
  onDone,
}: {
  subject: string;
  featureKey: string;
  onClose: () => void;
  onDone: Done;
}) {
  const { busy, error, run } = useAction(onDone, onClose);

  return (
    <Modal
      title="Jogosultság visszavonása"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="button"
            className="btn btn--danger"
            disabled={busy}
            onClick={() =>
              void run(
                () => api.revokeEntitlement(subject, featureKey),
                `Jogosultság visszavonva: ${featureKey}.`,
              )
            }
          >
            {busy ? 'Visszavonás…' : 'Visszavonom'}
          </button>
        </>
      }
    >
      <p>
        A(z) <code>{featureKey}</code> jogosultság megszűnik. A desktop app a következő
        indításkor már nem oldja fel a hozzá tartozó funkciót.
      </p>
      <p className="muted">A művelet a naplóba kerül.</p>
      {error ? <ErrorBox error={error} /> : null}
    </Modal>
  );
}

/**
 * Fiók letiltása / engedélyezése. Ez az egyetlen admin művelet, ami az
 * identitást birtokló rendszerbe (Keycloak) ír, ezért külön megerősítés.
 */
export function AccountStateDialog({
  subject,
  subjectLabel,
  enabled,
  onClose,
  onDone,
}: {
  subject: string;
  subjectLabel: string;
  enabled: boolean;
  onClose: () => void;
  onDone: Done;
}) {
  const [note, setNote] = useState('');
  const { busy, error, run } = useAction(onDone, onClose);
  const disabling = enabled;

  return (
    <Modal
      title={disabling ? 'Fiók letiltása' : 'Fiók engedélyezése'}
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn--ghost" onClick={onClose} disabled={busy}>
            Mégsem
          </button>
          <button
            type="button"
            className={disabling ? 'btn btn--danger' : 'btn btn--primary'}
            disabled={busy}
            onClick={() =>
              void run(
                () => api.setUserEnabled(subject, !enabled, note.trim() || undefined),
                disabling ? 'A fiók letiltva.' : 'A fiók engedélyezve.',
              )
            }
          >
            {busy ? 'Mentés…' : disabling ? 'Letiltom' : 'Engedélyezem'}
          </button>
        </>
      }
    >
      {disabling ? (
        <>
          <p>
            <strong>{subjectLabel}</strong> nem tud többé belépni, és a már kiadott
            munkamenete sem frissül tovább.
          </p>
          <p className="muted">
            Az előfizetése, a jogosultságai és a számlái érintetlenek maradnak – ez csak a
            belépést zárja. A desktop app fiók nélkül továbbra is működik nála, csak a felhő
            funkciók állnak le.
          </p>
        </>
      ) : (
        <p>
          <strong>{subjectLabel}</strong> újra be tud lépni. A jogosultságai változatlanok.
        </p>
      )}

      <label className="field">
        <span>Indoklás (a naplóba kerül)</span>
        <input
          type="text"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder={disabling ? 'pl. visszaélés gyanúja' : 'pl. tisztázódott'}
        />
      </label>

      {error ? <ErrorBox error={error} /> : null}
    </Modal>
  );
}
