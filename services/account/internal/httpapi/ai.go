package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lipcsei/snitt-cloud/services/account/internal/ai"
	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/errtrack"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// Felhős AI funkciók: a desktop app helyben indexeli a videók átiratát, a
// felhőt csak két dologra használja - szövegek beágyazására (szemantikus
// keresés) és egy rövid, forrásolt válaszra a megtalált részletekből. Mindkét
// hívás Snitt AI tokenbe kerül, amit az AI tárca egyenlegéből vonunk le.
//
// A sorrend minden hívásnál: szolgáltató bekapcsolva? (503) -> kérés
// érvényes? (400) -> van egyenleg? (402) -> szolgáltatóhívás -> levonás a
// ténylegesen elszámolt tokenek szerint. Az egyenleg-ellenőrzés és a levonás
// közt a felhasználó egy kéréssel túlköltheti a maradékát; ezt elnyeljük
// (lásd store.AIWallet.Debit), cserébe nem kell előre becsülni a költséget.

// Embedder a beágyazó szolgáltató (éles: ai.Voyage).
type Embedder interface {
	Model() string
	Embed(ctx context.Context, texts []string, inputType string) (ai.EmbedResult, error)
}

// Answerer a válaszadó szolgáltató (éles: ai.Claude).
type Answerer interface {
	Model() string
	Answer(ctx context.Context, req ai.AnswerRequest) (ai.AnswerResult, error)
}

// A kérések korlátai: a desktop app egy hívásban legfeljebb ennyit küld.
const (
	maxEmbedTexts      = 128
	maxEmbedTextRunes  = 8000
	minQuestionRunes   = 2
	maxQuestionRunes   = 1000
	maxPassages        = 40
	maxPassageRunes    = 60000
	maxEmbedBodyBytes  = 6 << 20 // 128 * 8000 rúna, rúnánként legfeljebb 4 bájt, plusz JSON
	maxAnswerBodyBytes = 1 << 20

	// aiCallTimeout a szolgáltatóhívás felső korlátja. A szerver általános
	// írási időkorlátja ennél rövidebb, ezért ezeken a végpontokon kitoljuk.
	aiCallTimeout = 150 * time.Second
)

// errNoAITokens a 402 válasz szövege.
const errNoAITokens = "Elfogytak az AI tokenjeid. Vásárolj AI csomagot vagy tokencsomagot a folytatáshoz."

type aiBalanceResponse struct {
	// Enabled: van legalább egy bekapcsolt AI szolgáltató.
	Enabled       bool `json:"enabled"`
	EmbedEnabled  bool `json:"embed_enabled"`
	AnswerEnabled bool `json:"answer_enabled"`
	// PeriodTokens a most költhető havi keret (lejárt vagy lemondott
	// csomagnál 0).
	PeriodTokens       int64      `json:"period_tokens"`
	PeriodEnd          *time.Time `json:"period_end"`
	PurchasedTokens    int64      `json:"purchased_tokens"`
	TotalTokens        int64      `json:"total_tokens"`
	SubscriptionActive bool       `json:"subscription_active"`
}

func (a *API) handleAIBalance(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "hiányzó identitás")
		return
	}
	wallet, err := a.store.AIWallet(r.Context(), id.Subject)
	if err != nil {
		a.log.ErrorContext(r.Context(), "AI tárca lekérdezése sikertelen", "error", err)
		serverError(w, r, err, http.StatusInternalServerError, "az AI egyenleg nem elérhető")
		return
	}
	now := time.Now()
	writeJSON(w, http.StatusOK, aiBalanceResponse{
		Enabled:            a.embedder != nil || a.answerer != nil,
		EmbedEnabled:       a.embedder != nil,
		AnswerEnabled:      a.answerer != nil,
		PeriodTokens:       wallet.ValidPeriodTokens(now),
		PeriodEnd:          wallet.PeriodEnd,
		PurchasedTokens:    max(wallet.PurchasedTokens, 0),
		TotalTokens:        wallet.Total(now),
		SubscriptionActive: wallet.SubscriptionActive(now),
	})
}

type aiEmbedRequest struct {
	Texts     []string `json:"texts"`
	InputType string   `json:"input_type"`
}

type aiEmbedResponse struct {
	Embeddings    [][]float32 `json:"embeddings"`
	Model         string      `json:"model"`
	Dims          int         `json:"dims"`
	ChargedTokens int64       `json:"charged_tokens"`
	TotalTokens   int64       `json:"total_tokens"`
}

func (a *API) handleAIEmbed(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "hiányzó identitás")
		return
	}
	if a.embedder == nil {
		writeError(w, http.StatusServiceUnavailable, "az AI keresés jelenleg nem elérhető")
		return
	}
	var req aiEmbedRequest
	if !decodeLargeBody(w, r, maxEmbedBodyBytes, &req) {
		return
	}
	if msg := validateEmbed(req); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if !a.requireAITokens(w, r, id.Subject) {
		return
	}

	ctx, cancel := a.aiCallContext(w, r)
	defer cancel()
	res, err := a.embedder.Embed(ctx, req.Texts, req.InputType)
	if err != nil {
		a.log.ErrorContext(r.Context(), "AI beágyazás sikertelen", "error", err)
		serverError(w, r, err, http.StatusBadGateway, "az AI szolgáltató nem elérhető, próbáld újra később")
		return
	}
	// A beágyazó súlyát a beállított modell adja: a válaszban visszakapott
	// modellnév eltérése ne vigye át a hívást az Opus súlyára.
	charged := ai.Charge(a.embedder.Model(), res.InputTokens, 0)
	total := a.debitAI(r, id.Subject, store.NewAIUsage{
		Kind: store.AIUsageEmbed, Model: res.Model, InputTokens: res.InputTokens, ChargedTokens: charged,
	})
	writeJSON(w, http.StatusOK, aiEmbedResponse{
		Embeddings: res.Embeddings, Model: res.Model, Dims: res.Dims,
		ChargedTokens: charged, TotalTokens: total,
	})
}

func validateEmbed(req aiEmbedRequest) string {
	if req.InputType != ai.InputQuery && req.InputType != ai.InputDocument {
		return `az input_type "query" vagy "document" lehet`
	}
	if len(req.Texts) == 0 || len(req.Texts) > maxEmbedTexts {
		return "a texts 1–128 szöveget tartalmazhat"
	}
	for _, t := range req.Texts {
		if strings.TrimSpace(t) == "" {
			return "a texts nem tartalmazhat üres szöveget"
		}
		if utf8.RuneCountInString(t) > maxEmbedTextRunes {
			return "egy szöveg legfeljebb 8000 karakter lehet"
		}
	}
	return ""
}

type aiAnswerRequest struct {
	Question string       `json:"question"`
	Language string       `json:"language"`
	Passages []ai.Passage `json:"passages"`
}

type aiAnswerResponse struct {
	Answer        string   `json:"answer"`
	CitedIDs      []string `json:"cited_ids"`
	Model         string   `json:"model"`
	ChargedTokens int64    `json:"charged_tokens"`
	TotalTokens   int64    `json:"total_tokens"`
}

func (a *API) handleAIAnswer(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "hiányzó identitás")
		return
	}
	if a.answerer == nil {
		writeError(w, http.StatusServiceUnavailable, "az AI válaszadás jelenleg nem elérhető")
		return
	}
	var req aiAnswerRequest
	if !decodeLargeBody(w, r, maxAnswerBodyBytes, &req) {
		return
	}
	if msg := validateAnswer(&req); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if !a.requireAITokens(w, r, id.Subject) {
		return
	}

	ctx, cancel := a.aiCallContext(w, r)
	defer cancel()
	res, err := a.answerer.Answer(ctx, ai.AnswerRequest{Question: req.Question, Language: req.Language, Passages: req.Passages})
	if errors.Is(err, ai.ErrRefusal) {
		// A szolgáltató a visszautasított kérést is kiszámlázza, ezért mi is
		// levonjuk - különben a visszautasítás ingyenes ismétlésre csábítana.
		a.debitAI(r, id.Subject, store.NewAIUsage{
			Kind: store.AIUsageAnswer, Model: res.Model, InputTokens: res.InputTokens, OutputTokens: res.OutputTokens,
			ChargedTokens: ai.Charge(res.Model, res.InputTokens, res.OutputTokens),
		})
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Az AI erre a kérdésre nem adott választ. Fogalmazd át a kérdést."})
		return
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "AI válasz sikertelen", "error", err)
		serverError(w, r, err, http.StatusBadGateway, "az AI szolgáltató nem elérhető, próbáld újra később")
		return
	}
	charged := ai.Charge(res.Model, res.InputTokens, res.OutputTokens)
	total := a.debitAI(r, id.Subject, store.NewAIUsage{
		Kind: store.AIUsageAnswer, Model: res.Model, InputTokens: res.InputTokens, OutputTokens: res.OutputTokens,
		ChargedTokens: charged,
	})
	cited := res.CitedIDs
	if cited == nil {
		cited = []string{}
	}
	writeJSON(w, http.StatusOK, aiAnswerResponse{
		Answer: res.Answer, CitedIDs: cited, Model: res.Model,
		ChargedTokens: charged, TotalTokens: total,
	})
}

func validateAnswer(req *aiAnswerRequest) string {
	req.Question = strings.TrimSpace(req.Question)
	if n := utf8.RuneCountInString(req.Question); n < minQuestionRunes || n > maxQuestionRunes {
		return "a kérdés 2–1000 karakter lehet"
	}
	req.Language = strings.TrimSpace(req.Language)
	if req.Language != "" && !validLocale(req.Language) {
		return "érvénytelen language (várt formátum: hu vagy hu-HU)"
	}
	if len(req.Passages) == 0 || len(req.Passages) > maxPassages {
		return "a passages 1–40 részletet tartalmazhat"
	}
	total := 0
	for _, p := range req.Passages {
		if strings.TrimSpace(p.ID) == "" {
			return "minden részletnek kell id"
		}
		total += utf8.RuneCountInString(p.Text)
	}
	if total > maxPassageRunes {
		return "a részletek összesen legfeljebb 60000 karakter lehetnek"
	}
	return ""
}

// requireAITokens 402-t ír, ha a felhasználónak nincs költhető AI tokenje.
func (a *API) requireAITokens(w http.ResponseWriter, r *http.Request, subject string) bool {
	wallet, err := a.store.AIWallet(r.Context(), subject)
	if err != nil {
		a.log.ErrorContext(r.Context(), "AI tárca lekérdezése sikertelen", "error", err)
		serverError(w, r, err, http.StatusInternalServerError, "az AI egyenleg nem elérhető")
		return false
	}
	if wallet.Total(time.Now()) <= 0 {
		writeError(w, http.StatusPaymentRequired, errNoAITokens)
		return false
	}
	return true
}

// debitAI levonja a felhasznált tokeneket, és a levonás utáni egyenleget
// adja. Ha a levonás nem sikerül, a választ akkor sem tartjuk vissza - a
// szolgáltatót már kifizettük -, de a hiba a logba és a hibajelentőbe megy.
func (a *API) debitAI(r *http.Request, subject string, u store.NewAIUsage) int64 {
	// A kliens bontása ne szakítsa félbe a levonást.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
	defer cancel()
	wallet, err := a.store.DebitAITokens(ctx, subject, u)
	if err != nil {
		a.log.ErrorContext(r.Context(), "AI tokenek levonása sikertelen", "error", err, "charged", u.ChargedTokens, "kind", u.Kind)
		errtrack.Report(r, err, http.StatusInternalServerError)
		if w, err := a.store.AIWallet(ctx, subject); err == nil {
			return w.Total(time.Now())
		}
		return 0
	}
	return wallet.Total(time.Now())
}

// aiCallContext a szolgáltatóhívás környezete: időkorláttal, és a szerver
// általános írási határidejét erre a kérésre kitolva.
func (a *API) aiCallContext(w http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc) {
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(aiCallTimeout + 15*time.Second))
	return context.WithTimeout(r.Context(), aiCallTimeout)
}

// decodeLargeBody a decodeBody nagyobb méretkorlátú változata az AI
// végpontokhoz; itt az üres törzs hiba.
func decodeLargeBody(w http.ResponseWriter, r *http.Request, limit int64, out any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "a kérés törzse túl nagy")
			return false
		}
		writeError(w, http.StatusBadRequest, "érvénytelen kérés törzs: "+err.Error())
		return false
	}
	return true
}
