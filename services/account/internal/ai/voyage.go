package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Beágyazás bemenet-fajtái: a Voyage a kereső kérdést és a tárolt
// dokumentumot más előtaggal kódolja, ezért a kettőt meg kell különböztetni.
const (
	InputQuery    = "query"
	InputDocument = "document"
)

// EmbedDims a kért vektorhossz. A desktop app helyi indexe erre a méretre
// épül, ezért rögzített, nem konfigurálható.
const EmbedDims = 1024

// EmbedResult a beágyazás eredménye, a bemenet sorrendjében.
type EmbedResult struct {
	Embeddings  [][]float32
	Model       string
	Dims        int
	InputTokens int64
}

// VoyageConfig a Voyage AI kliens beállításai.
type VoyageConfig struct {
	APIKey string
	Model  string
	// BaseURL a tesztekhez cserélhető; üresen a valódi Voyage API.
	BaseURL string
}

// Voyage a Voyage AI beágyazási REST API kliense. Hivatalos Go SDK nincs,
// egyetlen végpontot hívunk, ehhez elég a net/http.
type Voyage struct {
	cfg    VoyageConfig
	client *http.Client
}

// NewVoyage összeállítja a klienst.
func NewVoyage(cfg VoyageConfig) *Voyage {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.voyageai.com"
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Model == "" {
		cfg.Model = "voyage-4"
	}
	return &Voyage{cfg: cfg, client: &http.Client{Timeout: 60 * time.Second}}
}

// Model a használt beágyazó modell.
func (v *Voyage) Model() string { return v.cfg.Model }

// Embed a szövegek vektorait kéri le. A Voyage a válaszban adja meg a
// ténylegesen elszámolt tokenszámot (usage.total_tokens) - ez a mérés alapja.
func (v *Voyage) Embed(ctx context.Context, texts []string, inputType string) (EmbedResult, error) {
	reqBody, err := json.Marshal(map[string]any{
		"input":            texts,
		"model":            v.cfg.Model,
		"input_type":       inputType,
		"output_dimension": EmbedDims,
	})
	if err != nil {
		return EmbedResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.cfg.BaseURL+"/v1/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		return EmbedResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+v.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := v.client.Do(req)
	if err != nil {
		return EmbedResult{}, fmt.Errorf("voyage hívás: %w", err)
	}
	defer resp.Body.Close()
	// 128 darab 1024 dimenziós vektor JSON-ban néhány MB.
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return EmbedResult{}, fmt.Errorf("voyage válasz olvasása: %w", err)
	}
	if resp.StatusCode >= 300 {
		var e struct {
			Detail string `json:"detail"`
		}
		_ = json.Unmarshal(data, &e)
		return EmbedResult{}, fmt.Errorf("voyage %d: %s", resp.StatusCode, e.Detail)
	}
	var out struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
		Model string `json:"model"`
		Usage struct {
			TotalTokens int64 `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return EmbedResult{}, fmt.Errorf("voyage válasz értelmezése: %w", err)
	}
	if len(out.Data) != len(texts) {
		return EmbedResult{}, fmt.Errorf("voyage: %d szövegre %d vektor érkezett", len(texts), len(out.Data))
	}
	sort.Slice(out.Data, func(i, j int) bool { return out.Data[i].Index < out.Data[j].Index })
	res := EmbedResult{
		Embeddings:  make([][]float32, len(out.Data)),
		Model:       out.Model,
		Dims:        EmbedDims,
		InputTokens: out.Usage.TotalTokens,
	}
	if res.Model == "" {
		res.Model = v.cfg.Model
	}
	for i, d := range out.Data {
		if len(d.Embedding) == 0 {
			return EmbedResult{}, errors.New("voyage: üres vektor a válaszban")
		}
		res.Embeddings[i] = d.Embedding
	}
	res.Dims = len(res.Embeddings[0])
	return res, nil
}
