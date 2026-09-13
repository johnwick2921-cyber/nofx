package agent

import (
	"context"
	"log/slog"
	"nofx/mcp"
	"sync"
	"testing"
)

func TestAuthenticatedMessageDoesNotReplaceSharedAIClient(t *testing.T) {
	a := newTestAgentWithStore(t)
	a.config = DefaultConfig()
	a.logger = slog.Default()
	for _, user := range []string{"alice", "bob"} {
		if err := a.store.AIModel().Update(user, "openai", true, "test-key", "https://example.invalid/v1", "model-"+user); err != nil {
			t.Fatal(err)
		}
	}
	original := &staticAIClient{response: "unused"}
	a.SetAIClient(original)
	var wg sync.WaitGroup
	for i, user := range []string{"alice", "bob"} {
		wg.Add(1)
		go func(uid int64, user string) {
			defer wg.Done()
			_, err := a.HandleMessageForStoreUser(context.Background(), user, uid, "/status")
			if err != nil {
				t.Error(err)
			}
		}(int64(i+1), user)
	}
	wg.Wait()
	if a.aiClient != original {
		t.Fatal("authenticated request replaced shared/background AI client")
	}
}

func TestRequestRuntimeKeepsClientAndConversationOwnership(t *testing.T) {
	a := newTestAgentWithStore(t)
	a.config = DefaultConfig()
	a.logger = slog.Default()
	for _, user := range []string{"alice", "bob", "default"} {
		if err := a.store.AIModel().Update(user, "openai", true, "test-key", "https://example.invalid/v1", "model-"+user); err != nil {
			t.Fatal(err)
		}
	}
	first := a.requestRuntime("alice")
	second := a.requestRuntime("bob")
	if first == a || second == a || first.aiClient == second.aiClient {
		t.Fatal("request clients were shared")
	}
	if first.aiClient.(mcp.ClientEmbedder).BaseClient().Model != "model-alice" || second.aiClient.(mcp.ClientEmbedder).BaseClient().Model != "model-bob" {
		t.Fatal("selected model changed across requests")
	}
	if a.requestRuntime("unconfigured-user").aiClient != nil {
		t.Fatal("unconfigured user inherited default credentials")
	}
	if first.flowLock(42) != second.flowLock(42) {
		t.Fatal("request runtimes lost shared conversation serialization")
	}
	first.history.Add(42, "user", "retained")
	if len(a.history.Get(42)) != 1 {
		t.Fatal("conversation history detached")
	}
	first.saveSetupState(42, &SetupState{Step: "test-step"})
	if second.getSetupState(42).Step != "test-step" {
		t.Fatal("setup state detached")
	}
	second.clearSetupState(42)
	if first.getSetupState(42).Step != "" {
		t.Fatal("setup clear detached")
	}
}
