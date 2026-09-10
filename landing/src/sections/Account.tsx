import { Link } from 'react-router-dom';
import { useAuth } from '../AuthProvider';

export default function Account() {
  const { ready, authenticated, profile, register, login } = useAuth();

  return (
    <section id="fiok" className="section section-alt">
      <div className="container">
        <div className="account">
          <div className="account-copy">
            <span className="eyebrow eyebrow-pink">Fiók</span>
            <h2>Az alkalmazás fiók nélkül is mindent tud</h2>
            <p>
              Telepíted, megnyitod, dolgozol vele. Nincs regisztrációs fal, nincs próbaidőszak, és
              nem kell internet ahhoz, hogy keress a saját videóidban. Az indexed a te gépeden él,
              egyetlen adatbázisfájlban.
            </p>
            <p>
              A regisztráció itt a weboldalon <strong>a készülő extrákhoz</strong> tartozik: ezeken
              még dolgozunk, és amint elérhetők, a fiókoddal tudod majd használni őket. Amíg nem
              kérsz belőlük, semmit nem veszítesz — az asztali alkalmazás ugyanúgy megy tovább.
            </p>

            {authenticated && profile ? (
              <div className="account-cta">
                <Link to="/profil" className="btn btn-primary">
                  Profil megnyitása
                </Link>
                <span className="account-note">
                  Bejelentkezve mint <strong>{profile.name}</strong>
                </span>
              </div>
            ) : (
              <div className="account-cta">
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={register}
                  disabled={!ready}
                >
                  Regisztráció
                </button>
                <button type="button" className="btn btn-outline" onClick={login} disabled={!ready}>
                  Bejelentkezés
                </button>
              </div>
            )}
          </div>

          <ul className="account-list">
            <li>
              <strong>Fiók nélkül</strong>
              <span>
                Mappa indexelése, link behúzása, feliratok beolvasása, Whisper átirat, keresés,
                klipvágás — vagyis az egész alkalmazás.
              </span>
            </li>
            <li>
              <strong>Fiókkal, később</strong>
              <span>
                A készülő kiegészítők. Az alap működés ettől nem lesz fizetős vagy korlátozott.
              </span>
            </li>
            <li>
              <strong>Amit a fiók nem csinál</strong>
              <span>
                Nem tölti fel a videóidat, és nem szinkronizálja az indexedet. A könyvtárad a gépeden
                marad.
              </span>
            </li>
          </ul>
        </div>
      </div>
    </section>
  );
}
