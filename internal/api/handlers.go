package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kprompt/kprompt-dash/internal/kube"
)

const (
	maxEvents   = 20
	logTailLines int64 = 80
	maxLogBytes        = 32 << 10
)

// Mount registers /api/v1 routes.
func Mount(mux *http.ServeMux, cl *kube.Clients) {
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "context": cl.Context})
	})
	mux.HandleFunc("GET /api/v1/namespaces", listNamespaces(cl))
	mux.HandleFunc("GET /api/v1/namespaces/{ns}/deployments", listDeployments(cl))
	mux.HandleFunc("GET /api/v1/namespaces/{ns}/deployments/{name}", getDeployment(cl))
	mux.HandleFunc("GET /api/v1/namespaces/{ns}/pods", listPods(cl))
	mux.HandleFunc("GET /api/v1/namespaces/{ns}/pods/{name}", getPod(cl))
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

func getDeployment(cl *kube.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ns := r.PathValue("ns")
		name := r.PathValue("name")
		d, err := cl.Clientset.AppsV1().Deployments(ns).Get(r.Context(), name, metav1.GetOptions{})
		if err != nil {
			writeErr(w, err)
			return
		}
		events := listEvents(r, cl, ns, "Deployment", name)
		var logs string
		var logPod string
		if podName, err := pickDeploymentPod(r, cl, d); err == nil && podName != "" {
			logPod = podName
			logs, _ = tailPodLogs(r, cl, ns, podName)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind":       "Deployment",
			"name":       d.Name,
			"namespace":  d.Namespace,
			"context":    cl.Context,
			"age":        age(d.CreationTimestamp.Time),
			"ready":      fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Status.Replicas),
			"replicas":   d.Status.Replicas,
			"labels":     d.Labels,
			"conditions": deployConditions(d),
			"events":     events,
			"log_pod":    logPod,
			"logs":       logs,
		})
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

func getPod(cl *kube.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ns := r.PathValue("ns")
		name := r.PathValue("name")
		p, err := cl.Clientset.CoreV1().Pods(ns).Get(r.Context(), name, metav1.GetOptions{})
		if err != nil {
			writeErr(w, err)
			return
		}
		events := listEvents(r, cl, ns, "Pod", name)
		logs, _ := tailPodLogs(r, cl, ns, name)
		writeJSON(w, http.StatusOK, map[string]any{
			"kind":       "Pod",
			"name":       p.Name,
			"namespace":  p.Namespace,
			"context":    cl.Context,
			"age":        age(p.CreationTimestamp.Time),
			"phase":      string(p.Status.Phase),
			"ready":      podReady(p),
			"restarts":   restartCount(p),
			"node":       p.Spec.NodeName,
			"labels":     p.Labels,
			"conditions": podConditions(p),
			"events":     events,
			"logs":       logs,
		})
	}
}

func listEvents(r *http.Request, cl *kube.Clients, ns, kind, name string) []map[string]string {
	field := fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s", kind, name)
	list, err := cl.Clientset.CoreV1().Events(ns).List(r.Context(), metav1.ListOptions{
		FieldSelector: field,
		Limit:         100,
	})
	if err != nil {
		return []map[string]string{{"type": "Error", "reason": "ListFailed", "message": err.Error()}}
	}
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].LastTimestamp.Time.After(list.Items[j].LastTimestamp.Time)
	})
	out := make([]map[string]string, 0, maxEvents)
	for i, ev := range list.Items {
		if i >= maxEvents {
			break
		}
		msg := strings.TrimSpace(ev.Message)
		if len(msg) > 400 {
			msg = msg[:400] + "…"
		}
		out = append(out, map[string]string{
			"type":    ev.Type,
			"reason":  ev.Reason,
			"message": msg,
			"age":     age(ev.LastTimestamp.Time),
		})
	}
	return out
}

func pickDeploymentPod(r *http.Request, cl *kube.Clients, d *appsv1.Deployment) (string, error) {
	sel, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return "", err
	}
	list, err := cl.Clientset.CoreV1().Pods(d.Namespace).List(r.Context(), metav1.ListOptions{
		LabelSelector: sel.String(),
		Limit:         20,
	})
	if err != nil {
		return "", err
	}
	if len(list.Items) == 0 {
		return "", nil
	}
	// Prefer not-ready / high-restart pods for useful logs.
	best := list.Items[0]
	bestScore := podTroubleScore(&best)
	for i := 1; i < len(list.Items); i++ {
		p := list.Items[i]
		if s := podTroubleScore(&p); s > bestScore {
			best = p
			bestScore = s
		}
	}
	return best.Name, nil
}

func podTroubleScore(p *corev1.Pod) int {
	score := int(restartCount(p)) * 10
	if p.Status.Phase != corev1.PodRunning {
		score += 50
	}
	for _, c := range p.Status.ContainerStatuses {
		if !c.Ready {
			score += 20
		}
	}
	return score
}

func tailPodLogs(r *http.Request, cl *kube.Clients, ns, podName string) (string, error) {
	tail := logTailLines
	opts := &corev1.PodLogOptions{TailLines: &tail}
	stream, err := cl.Clientset.CoreV1().Pods(ns).GetLogs(podName, opts).Stream(r.Context())
	if err != nil {
		return "", err
	}
	defer stream.Close()
	data, err := io.ReadAll(io.LimitReader(stream, maxLogBytes))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func deployConditions(d *appsv1.Deployment) []map[string]string {
	out := make([]map[string]string, 0, len(d.Status.Conditions))
	for _, c := range d.Status.Conditions {
		out = append(out, map[string]string{
			"type":    string(c.Type),
			"status":  string(c.Status),
			"reason":  c.Reason,
			"message": truncate(c.Message, 240),
		})
	}
	return out
}

func podConditions(p *corev1.Pod) []map[string]string {
	out := make([]map[string]string, 0, len(p.Status.Conditions))
	for _, c := range p.Status.Conditions {
		out = append(out, map[string]string{
			"type":    string(c.Type),
			"status":  string(c.Status),
			"reason":  c.Reason,
			"message": truncate(c.Message, 240),
		})
	}
	return out
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
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
