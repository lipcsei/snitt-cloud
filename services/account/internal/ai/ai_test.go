package ai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCharge(t *testing.T) {
	tests := []struct {
		model   string
		in, out int64
		want    int64
	}{
		{"claude-opus-5-5", 1000, 200, 2000},
		{"claude-sonnet-5-5", 1000, 200, 1000},
		{"claude-haiku-5-5", 1000, 200, 50},
		{"claude-haiku-5-5", 3, 1, 1},          // 0,075 + 0,125 = 0,2 -> felfelé 1
		{"voyage-4", 1000, 0, 20},              // pontosan 20, lebegőpontos hiba nélkül
		{"voyage-4", 1001, 0, 21},              // 20,02 -> 21
		{"voyage-4", 1, 0, 1},                  // minimum 1
		{"ismeretlen-modell", 1000, 200, 2000}, // Opus súlyok
		{"claude-opus-5-5", 0, 0, 0},           // nem volt használat
	}
	for _, tt := range tests {
		if got := Charge(tt.model, tt.in, tt.out); got != tt.want {
			t.Errorf("Charge(%s, %d, %d) = %d, várt %d", tt.model, tt.in, tt.out, got, tt.want)
		}
	}
	if in, out := Weights("claude-sonnet-5-5"); in != 0.5 || out != 2.5 {
		t.Errorf("sonnet súlyok = %v / %v", in, out)
	}
}

func TestCitedIDs(t *testing.T) {
	ps := []Passage{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	tests := []struct {
		answer string
		want   string
	}{
		{"Igen [2]. Később is [1][2].", "b,a"},
		{"Lásd [3, 1] és [2,3].", "c,a,b"},
		{"Nem létező [4] és [0] hivatkozás.", ""},
		{"A videókban erre nincs válasz.", ""},
	}
	for _, tt := range tests {
		if got := strings.Join(CitedIDs(tt.answer, ps), ","); got != tt.want {
			t.Errorf("CitedIDs(%q) = %q, várt %q", tt.answer, got, tt.want)
		}
	}
}

func TestPrompts(t *testing.T) {
	up := UserPrompt(" Mit mond a rizsről? ", []Passage{
		{ID: "x", VideoTitle: "Főzés\nalapok", StartMS: 83_500, Text: " A rizst mosd meg. "},
		{ID: "y", VideoTitle: "", StartMS: 3_725_000, Text: "Második"},
	})
	for _, want := range []string{"[1] Főzés alapok @ 01:23\nA rizst mosd meg.", "[2] (cím nélkül) @ 62:05\nMásodik", "Kérdés: Mit mond a rizsről?"} {
		if !strings.Contains(up, want) {
			t.Errorf("a felhasználói prompt nem tartalmazza: %q\n%s", want, up)
		}
	}
	if !strings.Contains(SystemPrompt("en"), "ezen a nyelven: en") || !strings.Contains(SystemPrompt(""), "amelyen a kérdés szól") {
		t.Error("a rendszerprompt nyelvi utasítása hibás")
	}
}

func TestVoyageEmbed(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" || r.Header.Get("Authorization") != "Bearer vk" {
			http.Error(w, `{"detail":"bad"}`, http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		// Fordított sorrendben válaszol: az index szerint kell visszarendezni.
		_, _ = w.Write([]byte(`{"object":"list","data":[
			{"object":"embedding","embedding":[0.3,0.4],"index":1},
			{"object":"embedding","embedding":[0.1,0.2],"index":0}],
			"model":"voyage-4","usage":{"total_tokens":42}}`))
	}))
	defer srv.Close()

	v := NewVoyage(VoyageConfig{APIKey: "vk", BaseURL: srv.URL})
	res, err := v.Embed(context.Background(), []string{"egy", "kettő"}, InputQuery)
	if err != nil {
		t.Fatal(err)
	}
	if res.InputTokens != 42 || res.Model != "voyage-4" || res.Dims != 2 || res.Embeddings[0][0] != 0.1 || res.Embeddings[1][1] != 0.4 {
		t.Errorf("eredmény = %+v", res)
	}
	if got["model"] != "voyage-4" || got["input_type"] != "query" || got["output_dimension"] != float64(EmbedDims) {
		t.Errorf("kérés = %v", got)
	}

	if _, err := NewVoyage(VoyageConfig{APIKey: "rossz", BaseURL: srv.URL}).Embed(context.Background(), []string{"x"}, InputQuery); err == nil {
		t.Error("hibás kulcsnál hibát vártunk")
	}
}

// fakeMessagesAPI egy előre megírt Messages API választ ad, és megjegyzi a kérést.
func fakeMessagesAPI(t *testing.T, response string) (*httptest.Server, *map[string]any) {
	t.Helper()
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.Header.Get("X-Api-Key") != "ak" {
			http.Error(w, `{"type":"error","error":{"type":"authentication_error","message":"bad"}}`, http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

func TestClaudeAnswer(t *testing.T) {
	srv, got := fakeMessagesAPI(t, `{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-5-5",
		"content":[{"type":"thinking","thinking":"","signature":"sig"},{"type":"text","text":"A rizst meg kell mosni [2], majd főzni [1][2]."}],
		"stop_reason":"end_turn","usage":{"input_tokens":1200,"output_tokens":80}}`)

	c := NewClaude(ClaudeConfig{APIKey: "ak", BaseURL: srv.URL})
	res, err := c.Answer(context.Background(), AnswerRequest{
		Question: "Hogyan főzzem a rizst?",
		Language: "hu",
		Passages: []Passage{{ID: "p1", VideoTitle: "Főzés", Text: "főzd"}, {ID: "p2", VideoTitle: "Főzés", StartMS: 61000, Text: "mosd"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Answer != "A rizst meg kell mosni [2], majd főzni [1][2]." || strings.Join(res.CitedIDs, ",") != "p2,p1" {
		t.Errorf("válasz = %+v", res)
	}
	if res.InputTokens != 1200 || res.OutputTokens != 80 || res.Model != "claude-opus-5-5" {
		t.Errorf("használat = %+v", res)
	}
	req := *got
	if req["model"] != "claude-opus-5-5" || req["max_tokens"] != float64(16000) {
		t.Errorf("kérés = %v", req)
	}
	// Ezen a modellen a thinking kikapcsolása 400-at adna: nem küldhetjük.
	if _, ok := req["thinking"]; ok {
		t.Errorf("a kérés nem tartalmazhat thinking beállítást: %v", req["thinking"])
	}
	if oc, _ := req["output_config"].(map[string]any); oc["effort"] != "low" {
		t.Errorf("output_config = %v", req["output_config"])
	}
	msgs, _ := json.Marshal(req["messages"])
	if !strings.Contains(string(msgs), `[2] Főzés @ 01:01`) {
		t.Errorf("üzenetek = %s", msgs)
	}
}

func TestClaudeRefusal(t *testing.T) {
	srv, _ := fakeMessagesAPI(t, `{"id":"msg_2","type":"message","role":"assistant","model":"claude-opus-5-5",
		"content":[],"stop_reason":"refusal","stop_details":{"type":"refusal","category":"cyber","explanation":"x"},
		"usage":{"input_tokens":500,"output_tokens":3}}`)
	res, err := NewClaude(ClaudeConfig{APIKey: "ak", BaseURL: srv.URL}).Answer(context.Background(), AnswerRequest{
		Question: "kérdés", Passages: []Passage{{ID: "p1", Text: "x"}},
	})
	if !errors.Is(err, ErrRefusal) {
		t.Fatalf("hiba = %v, várt ErrRefusal", err)
	}
	if res.InputTokens != 500 {
		t.Errorf("a használatot elutasításkor is vissza kell adni: %+v", res)
	}
}
