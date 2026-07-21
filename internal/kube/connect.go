package kube

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Clients wraps a typed client and the resolved context name.
type Clients struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Context   string
}

// Connect loads kubeconfig (KUBECONFIG / default) and optional context override.
func Connect(contextName string) (*Clients, error) {
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
	raw, err := cfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	// Bound blast radius for a local UI process.
	raw.Timeout = 15 * time.Second
	raw.QPS = 20
	raw.Burst = 40
	rawCfg, err := cfg.RawConfig()
	if err != nil {
		return nil, err
	}
	ctxName := contextName
	if ctxName == "" {
		ctxName = rawCfg.CurrentContext
	}
	cs, err := kubernetes.NewForConfig(raw)
	if err != nil {
		return nil, err
	}
	return &Clients{Clientset: cs, Config: raw, Context: ctxName}, nil
}
