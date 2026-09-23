import { useState } from 'react';
import { useT } from '../../i18n';
import { useAuth } from '../../AuthProvider';
import '../../styles/email-verify.css';

const DISMISS_KEY = 'snitt.emailVerifyDismissed';

function readDismissed(): boolean {
  try {
    return localStorage.getItem(DISMISS_KEY) === '1';
  } catch {
    return false;
  }
}

function writeDismissed(): void {
  try {
    localStorage.setItem(DISMISS_KEY, '1');
  } catch {
    /* privát böngészés / letiltott tárhely - a kártya ilyenkor a lap újratöltéséig marad látható */
  }
}

/**
 * Lebegő, nem-modális kártya a bejelentkezett, de meg nem erősített e-mail című felhasználóknak.
 * Sosem blokkol semmit (a realm `verifyEmail` auto-assign-ja szándékosan ki van kapcsolva, l.
 * sso/realms/snitt-realm.json) - a Snitt a megerősítés nélkül is ugyanúgy használható, ez csak egy
 * felkínált lehetőség. A ConsentBanner mintáját követi (nem-modális, ugyanaz a `.btn` gombstílus),
 * de a tetején jelenik meg (nem alul), hogy a két sáv sosem fedje egymást.
 */
export default function EmailVerifyBanner() {
  const t = useT();
  const { ready, authenticated, profile, verifyEmail } = useAuth();
  const [dismissed, setDismissed] = useState(readDismissed);

  if (!ready || !authenticated || !profile || profile.emailVerified || dismissed) return null;

  const dismiss = () => {
    writeDismissed();
    setDismissed(true);
  };

  return (
    <section className="email-verify-card" role="region" aria-labelledby="email-verify-card-title">
      <p id="email-verify-card-title" className="email-verify-card-title">
        {t.emailVerify.title}
      </p>
      <p className="email-verify-card-text">{t.emailVerify.message}</p>
      <div className="email-verify-card-actions">
        <button type="button" className="btn btn-primary" onClick={verifyEmail}>
          {t.emailVerify.confirm}
        </button>
        <button type="button" className="btn btn-outline" onClick={dismiss}>
          {t.emailVerify.dismiss}
        </button>
      </div>
    </section>
  );
}
