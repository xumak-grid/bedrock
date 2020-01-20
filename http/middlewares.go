package http

import (
	"context"
	"net/http"

	"github.com/xumak-grid/bedrock/k8s"
)

// ContextKey represents a string key for request context.
type ContextKey string

// K8sClient context key
const K8sClient = ContextKey("k8client")

// K8sAEMClient context key
const K8sAEMClient = ContextKey("aemk8scli")

// CertManagerClientKey context key
const CertManagerClientKey = ContextKey("certManagerClient")

// WithKubeClient adapts a handler with a KubeClient.
func WithKubeClient() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg, _ := k8s.BuildKubeConfig()
			client := k8s.NewKubeClient(cfg)
			aem, _ := k8s.AEMClient(cfg)
			certManagerClient, _ := k8s.CertManagerClient(cfg)
			ctx := context.WithValue(r.Context(), K8sClient, client)
			ctx = context.WithValue(ctx, K8sAEMClient, aem)
			ctx = context.WithValue(ctx, CertManagerClientKey, certManagerClient)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
