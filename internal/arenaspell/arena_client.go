package arenaspell

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type applyRequest struct {
    MatchID   string `json:"match_id"`
    CasterID  uint   `json:"caster_id"`
    SpellType string `json:"spell_type"`
}

func getArenaHTTPAddr() string {
    if v := os.Getenv("ARENA_HTTP_ADDR"); v != "" {
        return v
    }
    return "http://localhost:8082"
}

// ApplySpellEffectViaArena calls Arena HTTP to apply the effect in the match
func ApplySpellEffectViaArena(matchID string, casterID uint, spellType string) error {
    payload := applyRequest{MatchID: matchID, CasterID: casterID, SpellType: spellType}
    b, _ := json.Marshal(payload)
    url := fmt.Sprintf("%s/api/v1/arena/spells/apply", getArenaHTTPAddr())
    req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 {
        return fmt.Errorf("arena apply failed: status %d", resp.StatusCode)
    }
    return nil
}


