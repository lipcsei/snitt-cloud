package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// audit egy már elvégzett admin műveletet ír a naplóba.
//
// Szándékosan a művelet UTÁN fut, és a hibája nem bukatja el a kérést: a
// pénzt érintő művelet ilyenkor már megtörtént, egy 500-as válasz csak azt
// hazudná az adminnak, hogy nem sikerült. A naplóírás hibája viszont
// Error szinten kimegy a logba, hogy ne maradjon észrevétlen.
func (a *API) audit(r *http.Request, e store.NewAuditEntry) {
	id, _ := auth.FromContext(r.Context())
	e.ActorSubject = id.Subject
	e.ActorLabel = actorLabel(id)

	if _, err := a.store.RecordAudit(r.Context(), e); err != nil {
		a.log.ErrorContext(r.Context(), "admin napló írása sikertelen",
			"error", err, "action", e.Action, "target_id", e.TargetID, "actor", e.ActorSubject)
	}
}

// actorLabel az adminról a tokenben látott, emberi olvasásra szánt címke.
func actorLabel(id auth.Identity) string {
	for _, s := range []string{id.Email, id.PreferredUsername, id.Name} {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return id.Subject
}

// moneyText összeg a pénznem legkisebb egységéből emberi alakra
// ("2 990,00 HUF"). A naplóban szöveg marad, hogy évek múlva is ugyanazt
// jelentse, akkor is, ha a felület formázása időközben változik.
func moneyText(minor int64, currency string) string {
	neg := minor < 0
	if neg {
		minor = -minor
	}
	whole, frac := minor/100, minor%100
	s := groupThousands(whole) + "," + fmt.Sprintf("%02d", frac)
	if neg {
		s = "-" + s
	}
	if currency != "" {
		s += " " + currency
	}
	return s
}

func groupThousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	lead := len(s) % 3
	if lead > 0 {
		b.WriteString(s[:lead])
	}
	for i := lead; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// ---------- napló végpont ----------

type auditResponse struct {
	Entries  []store.AuditEntry `json:"entries"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int                `json:"total"`
	Actions  []string           `json:"actions"`
}

func (a *API) handleListAudit(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	q := r.URL.Query()

	action := strings.TrimSpace(q.Get("action"))
	if action != "" && !validAuditAction(action) {
		writeError(w, http.StatusBadRequest, "ismeretlen művelet: "+action)
		return
	}

	since, err := timeParam(q.Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "érvénytelen kezdő dátum: "+q.Get("from"))
		return
	}
	until, err := timeParam(q.Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "érvénytelen záró dátum: "+q.Get("to"))
		return
	}
	// A "to" napot az admin bezárólag érti, a lekérdezés viszont kizárólagos
	// felső határral dolgozik - ezért a napra pontos értéket a következő nap
	// kezdetére toljuk.
	if until != nil && q.Get("to") != "" && len(strings.TrimSpace(q.Get("to"))) == len(dateLayout) {
		next := until.AddDate(0, 0, 1)
		until = &next
	}

	entries, total, err := a.store.ListAudit(r.Context(), store.AuditFilter{
		Actor:   strings.TrimSpace(q.Get("actor")),
		Subject: strings.TrimSpace(q.Get("subject")),
		Action:  action,
		Since:   since,
		Until:   until,
		Limit:   size,
		Offset:  (page - 1) * size,
	})
	if err != nil {
		a.log.ErrorContext(r.Context(), "napló lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a napló nem elérhető")
		return
	}

	writeJSON(w, http.StatusOK, auditResponse{
		Entries: entries, Page: page, PageSize: size, Total: total, Actions: store.AuditActions(),
	})
}

const dateLayout = "2006-01-02"

// timeParam napra pontos (2026-09-10) vagy teljes RFC3339 időbélyeget fogad.
func timeParam(v string) (*time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	if t, err := time.Parse(dateLayout, v); err == nil {
		t = t.UTC()
		return &t, nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, err
	}
	t = t.UTC()
	return &t, nil
}

func validAuditAction(s string) bool {
	for _, a := range store.AuditActions() {
		if a == s {
			return true
		}
	}
	return false
}
