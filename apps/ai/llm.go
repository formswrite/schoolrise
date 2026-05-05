package ai

import (
	"os"
	"strings"
	"sync"

	baml "github.com/boundaryml/baml/engine/language_client_go/pkg"
)

const (
	defaultPrimaryClient = "SchoolRisePrimary"
	stubProviderName     = "stub"
	liveProviderName     = "openai+anthropic-fallback"
	defaultLiveModel     = "gpt-5-mini → claude-3-5-haiku"
	stubModel            = "stub-llm"
)

var (
	registryMu    sync.RWMutex
	testRegistry  *baml.ClientRegistry
	stubMode      bool
	resolvedModel = defaultLiveModel
)

func init() {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if isPlaceholderKey(apiKey) {
		stubMode = true
		resolvedModel = stubModel
	}
}

func isPlaceholderKey(k string) bool {
	low := strings.ToLower(strings.TrimSpace(k))
	switch low {
	case "", "sk-...", "sk-xxx", "stub", "placeholder", "changeme":
		return true
	}
	return strings.HasPrefix(low, "sk-placeholder") || strings.HasPrefix(low, "sk-stub")
}

func SetTestRegistry(r *baml.ClientRegistry) {
	registryMu.Lock()
	testRegistry = r
	if r != nil {
		stubMode = true
		resolvedModel = stubModel
	}
	registryMu.Unlock()
}

func SetStubMode(on bool) {
	registryMu.Lock()
	stubMode = on
	if on {
		resolvedModel = stubModel
	} else {
		resolvedModel = defaultLiveModel
	}
	registryMu.Unlock()
}

func currentRegistry() *baml.ClientRegistry {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return testRegistry
}

func providerName() string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if stubMode && testRegistry == nil {
		return stubProviderName
	}
	return liveProviderName
}

func providerModel() string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return resolvedModel
}

func providerEnabled() bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return testRegistry != nil || !stubMode
}
