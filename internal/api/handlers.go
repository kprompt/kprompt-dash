package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kprompt/kprompt-dash/internal/kube"
)

// Mount registers /api/v1 routes.
func Mount(mux *http.ServeMux, cl *kube.Clients) {
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "context": cl.Context})
	})
	mux.HandleFunc("GET /api/v1/namespaces", listNamespaces(cl))
	mux.HandleFunc("GET /api/v1/namespaces/{ns}/deployments", listDeployments(cl))
	mux.HandleFunc("GET /api/v1/namespaces/{ns}/pods", listPods(cl))
}

func listNamespaces(cl *kube.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := cl.Clientset.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{Limit: 500})
		if err != nil {
			writeErr(w, err)
			return
		}
		out := make([]map[string]string, 0, len(list.Items))
		for _, ns := range list.Items {
			out = append(out, map[string]string{
				"name":  ns.Name,
				"phase": string(ns.Status.Phase),
				"age":   age(ns.CreationTimestamp.Time),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": out, "context": cl.Context})
	}
}

func listDeployments(cl *kube.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ns := r.PathValue("ns")
		list, err := cl.Clientset.AppsV1().Deployments(ns).List(r.Context(), metav1.ListOptions{Limit: 500})
		if err != nil {
			writeErr(w, err)
			return
		}
		out := make([]map[string]any, 0, len(list.Items))
		for _, d := range list.Items {
			out = append(out, map[string]any{
				"name":      d.Name,
				"namespace": d.Namespace,
				"ready":     fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Status.Replicas),
				"replicas":  d.Status.Replicas,
				"ready_n":   d.Status.ReadyReplicas,
				"age":       age(d.CreationTimestamp.Time),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": out, "namespace": ns})
	}
}

func listPods(cl *kube.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ns := r.PathValue("ns")
		list, err := cl.Clientset.CoreV1().Pods(ns).List(r.Context(), metav1.ListOptions{Limit: 500})
		if err != nil {
			writeErr(w, err)
			return
		}
		out := make([]map[string]any, 0, len(list.Items))
		for _, p := range list.Items {
			out = append(out, map[string]any{
				"name":      p.Name,
				"namespace": p.Namespace,
				"phase":     string(p.Status.Phase),
				"ready":     podReady(&p),
				"restarts":  restartCount(&p),
				"node":      p.Spec.NodeName,
				"age":       age(p.CreationTimestamp.Time),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": out, "namespace": ns})
	}
}

func podReady(p *corev1.Pod) string {
	var ready, total int
	for _, c := range p.Status.ContainerStatuses {
		total++
		if c.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

func restartCount(p *corev1.Pod) int32 {
	var n int32
	for _, c := range p.Status.ContainerStatuses {
		n += c.RestartCount
	}
	return n
}

func age(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t).Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadGateway, map[string]string{
		"error": err.Error(),
	})
}
