// Package ai a felhős AI funkciók szolgáltatói: beágyazás (Voyage AI) és
// válaszadás a felhasználó saját videótárából (Claude), valamint a
// szolgáltatói tokenek Snitt AI tokenre váltása.
//
// # Snitt AI token
//
// A felhasználó egyetlen egységben fizet: 1 Snitt AI token nagyjából egy
// Claude Opus 5.5 bemeneti tokennek felel meg. Minden más modell és a
// kimeneti tokenek a szolgáltatói árarány szerinti súllyal számítanak, így
// az egyenleg modellváltáskor sem csúszik el az önköltségtől.
package ai

// weight egy modell súlya: hány Snitt AI tokent ér egy bemeneti és egy
// kimeneti szolgáltatói token, milliomodokban. Egész számmal számolunk, hogy
// a lebegőpontos kerekítés (0.02 * 1000 = 20.000000000000004) ne vigyen fel
// egy tokent a felfelé kerekítésnél.
type weight struct {
	inMicro, outMicro int64
}

const micro = 1_000_000

// weights a modellenkénti súlytábla. Ismeretlen modell az Opus súlyát kapja:
// inkább számoljunk többet, mint hogy egy új, drágább modell ingyen fusson.
var weights = map[string]weight{
	"claude-opus-5-5":   {1 * micro, 5 * micro},
	"claude-sonnet-5-5": {micro / 2, 5 * micro / 2}, // 0,5 / 2,5
	"claude-haiku-5-5":  {micro / 40, micro / 8},    // 0,025 / 0,125
	"voyage-4":          {micro / 50, 0},            // 0,02 / 0
}

const fallbackModel = "claude-opus-5-5"

// Weights a modell súlya (bemenet, kimenet) tört számként - a
// dokumentációhoz és a tesztekhez.
func Weights(model string) (in, out float64) {
	w := lookupWeight(model)
	return float64(w.inMicro) / micro, float64(w.outMicro) / micro
}

func lookupWeight(model string) weight {
	if w, ok := weights[model]; ok {
		return w
	}
	return weights[fallbackModel]
}

// Charge a levonandó Snitt AI token: ceil(input*w_in + output*w_out). Ha
// bármennyi token felhasználásra került, legalább 1.
func Charge(model string, inputTokens, outputTokens int64) int64 {
	inputTokens, outputTokens = max(inputTokens, 0), max(outputTokens, 0)
	if inputTokens == 0 && outputTokens == 0 {
		return 0
	}
	w := lookupWeight(model)
	total := inputTokens*w.inMicro + outputTokens*w.outMicro
	charged := (total + micro - 1) / micro
	return max(charged, 1)
}
