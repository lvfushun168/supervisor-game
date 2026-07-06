package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"supervisor-game/internal/config"
	appcrypto "supervisor-game/internal/crypto"
	"supervisor-game/internal/model"
)

func TestPatrolDeltas(t *testing.T) {
	rule := DefaultPatrolRule()
	cases := []struct {
		status         string
		warningDelta   int
		violationDelta int
	}{
		{"normal", 0, 0},
		{"suspicious", 1, 0},
		{"violation", 1, 1},
		{"using_phone", 1, 1},
		{"sleeping", 1, 1},
		{"absent", 1, 1},
		{"uncertain", 1, 0},
	}
	for _, item := range cases {
		warningDelta, violationDelta := patrolDeltas(item.status, rule)
		if warningDelta != item.warningDelta || violationDelta != item.violationDelta {
			t.Fatalf("%s delta = %d/%d, want %d/%d", item.status, warningDelta, violationDelta, item.warningDelta, item.violationDelta)
		}
	}
}

func TestMappedActionKeyUsesSceneJSON(t *testing.T) {
	got := mappedActionKey(`{"normal":"patrol_normal","uncertain":"patrol_suspicious"}`, "uncertain")
	if got != "patrol_suspicious" {
		t.Fatalf("mapped action = %q", got)
	}
}

func TestParseModelPatrolResultRejectsInvalidStatus(t *testing.T) {
	_, err := parseModelPatrolResult([]byte(`{"status":"dancing","confidence":0.8,"reason":"bad"}`))
	var patrolErr PatrolError
	if !errors.As(err, &patrolErr) || patrolErr.Code != PatrolErrModelResponseInvalid {
		t.Fatalf("expected model response invalid, got %v", err)
	}
}

func TestCallVisionModelParsesOpenAICompatibleResponse(t *testing.T) {
	key := "test-key"
	encrypted, err := appcrypto.EncryptString(key, "secret-token")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Fatalf("authorization header was not set")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": `{"status":"using_phone","confidence":0.91,"reason":"检测到手机","actionKey":"patrol_phone"}`}},
			},
		})
	}))
	defer server.Close()

	svc := Service{cfg: config.Config{ConfigEncryptionKey: key}}
	result, err := svc.callVisionModel(model.ModelConfig{
		Enabled:         true,
		BaseURL:         server.URL,
		APIKeyEncrypted: encrypted,
		Model:           "vision-test",
		TimeoutMS:       1000,
		RetryCount:      0,
		Prompt:          "check",
	}, DefaultPatrolRule(), "data:image/jpeg;base64,aaa")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "using_phone" || result.ActionKey != "patrol_phone" {
		t.Fatalf("unexpected result %#v", result)
	}
}
