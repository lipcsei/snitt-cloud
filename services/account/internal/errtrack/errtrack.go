// Package errtrack a szolgáltatás opcionális hibajelentése egy Sentry-kompatibilis szerverre (a
// saját üzemeltetésű GlitchTipre). Minden a környezeti változókból vezérelt (lásd internal/config),
// és kikapcsolt alapállapotú: üres SENTRY_DSN esetén az SDK inicializálása sem történik meg, nincs
// hálózati forgalom, és a szolgáltatás pontosan úgy viselkedik, mintha ez a csomag nem is létezne.
//
//	flush := errtrack.Setup(cfg, log) // a run() elején
//	defer flush()                     // leálláskor kiüríti a függő eseményeket
//
// Két dolgot jelent: a handlerben keletkező pánikot (Middleware) és a visszaadott 5xx hibákat
// (Report). A 4xx kliens-hiba, azt sosem.
package errtrack

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// flushTimeout ennyit vár a leállításkor a függő események elküldésére.
const flushTimeout = 2 * time.Second

// Config a hibajelentés beállításai.
type Config struct {
	// DSN a GlitchTip-projekt DSN-je; üres esetén a hibajelentés ki van kapcsolva.
	DSN string
	// Environment a környezet neve az eseményeken (pl. development, production).
	Environment string
	// Release a kiadás azonosítója (üres: az SDK alapértelmezése).
	Release string
	// Transport az események szállítója; üresen a valódi HTTP-szállító. Csak tesztekhez.
	Transport sentry.Transport
}

// Setup a megadott beállítással bekapcsolja a hibajelentést. Sosem hibázik és sosem pánikol: üres
// vagy érvénytelen DSN esetén a hibajelentés ki marad kapcsolva (az utóbbit figyelmeztetés jelzi,
// mert egy elgépelt DSN miatt nem érdemes leállítani a szolgáltatást).
//
// A visszaadott függvény leálláskor (defer) kiüríti a függő eseményeket (legfeljebb 2 másodpercig),
// majd kikapcsolja a jelentést; kikapcsolt módban nem csinál semmit. Nem nil.
func Setup(cfg Config, log *slog.Logger) (flush func()) {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	noop := func() {}
	if cfg.DSN == "" {
		log.Debug("hibajelentés kikapcsolva (nincs SENTRY_DSN)")
		return noop
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.DSN,
		Environment: cfg.Environment,
		Release:     cfg.Release,
		Transport:   cfg.Transport,
		// Személyes adat nélkül: se IP-cím, se süti, se érzékeny fejléc (Authorization stb.).
		// Kérés törzsét az SDK ilyenkor sem csatol.
		SendDefaultPII: false,
		// Nyomkövetés (tracing) nincs: nincs TracesSampleRate. Kliensjelentés sincs: csak a
		// hibaesemények mennek ki.
		DisableClientReports: true,
		// Nélküle a szöveges pánik (panic("...")) esemény stacktrace nélkül menne ki; a
		// runtime hibáké (nil pointer, index) enélkül is teljes.
		AttachStacktrace: true,
		BeforeSend:       scrub,
	})
	if err != nil {
		// Az err szövege a nyers DSN-t is tartalmazhatja (kulccsal), ezért nem naplózzuk.
		log.Warn("a SENTRY_DSN érvénytelen, a hibajelentés kikapcsolva (alak: https://<kulcs>@<host>/<projekt-azonosító>)")
		return noop
	}
	log.Info("hibajelentés bekapcsolva", "environment", cfg.Environment, "release", cfg.Release)

	return func() {
		sentry.Flush(flushTimeout)
		sentry.CurrentHub().BindClient(nil)
	}
}

// enabled: van-e működő kliens (Setup után igaz, ha a DSN érvényes volt).
func enabled() bool { return sentry.CurrentHub().Client() != nil }

// scrub kivesz az eseményekből mindent, ami személyes adatot hordozhat: a kérés törzsét és a
// lekérdezés-szöveget (az admin felhasználókeresése a ?q= paraméterben nevet/e-mailt vihet).
// Az SDK maga sem csatol törzset, ez csak biztonsági öv.
func scrub(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event.Request != nil {
		event.Request.Data = ""
		event.Request.QueryString = ""
		event.Request.Cookies = ""
	}
	return event
}

// Middleware kérésenként külön Sentry-hubot állít fel, és a handlerben keletkező pánikot jelenti
// (pontos stacktrace-szel, egyetlen eseményként), majd újra eldobja (Repanic): a net/http saját
// pánik-kezelése (naplózás, a kapcsolat lezárása) így úgy működik tovább, ahogy eddig.
//
// A pánikból nem lesz második esemény: a pánik felszámolja a handlert, tehát a Report-ig el sem jut.
//
// Kikapcsolt hibajelentésnél átengedi a kérést, a viselkedés változatlan. A legkülső rétegnek kell
// lennie, hogy a többi middleware (CORS, hitelesítés) pánikját is lássa.
func Middleware(next http.Handler) http.Handler {
	active := sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !enabled() {
			next.ServeHTTP(w, r)
			return
		}
		active.ServeHTTP(w, r)
	})
}

// Report egy 5xx választ kiváltó hibát jelent, kérésenként egy eseményként. A hívó dolga, hogy
// csak 5xx-nél hívja, és válaszonként legfeljebb egyszer (lásd httpapi.serverError). Nem jelenti a
// megszakadt kérést (context.Canceled: a kliens elment, nem a szolgáltatás hibája).
//
// A kérés hubját használja (a Middleware állítja fel), ennek híján a globális egy másolatát;
// kikapcsolt hibajelentésnél nem csinál semmit.
func Report(r *http.Request, err error, status int) {
	if err == nil || !enabled() || errors.Is(err, context.Canceled) {
		return
	}
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetRequest(r)
		scope.SetTag("http.status_code", strconv.Itoa(status))
		// A route-minta (pl. "GET /api/v1/me"), nem a konkrét URL: csoportosításra jó.
		if r.Pattern != "" {
			scope.SetTag("http.route", r.Pattern)
		}
		hub.CaptureException(err)
	})
}
