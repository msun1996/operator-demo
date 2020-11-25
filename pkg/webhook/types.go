package webhook

import (
	"fmt"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	admits  = map[string]WebhookFunc{}
	mutates = map[string]WebhookFunc{}
)

type patchAction string

func (p patchAction) string() string {
	return string(p)
}

const (
	PatchForAdd     patchAction = "add"
	PatchForReplace patchAction = "replace"
)

type PatchOperation struct {
	Op    patchAction `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

//WebhookHandler handles admission
type WebhookHandler func(v1.AdmissionReview) *v1.AdmissionResponse

//WebhookFunc admission and mutating handle func type,
//input args is current object and old object
//return patch operations if type is mutating else is nil
type WebhookFunc func([]byte, []byte, v1.Operation, client.Client, logr.Logger) ([]PatchOperation, error)

func (w Webhook) RegisterAdmit(gvk metav1.GroupVersionKind, f WebhookFunc) {
	admits[getKeyByGVK(gvk)] = f
}

func (w Webhook) RegisterMutate(gvk metav1.GroupVersionKind, f WebhookFunc) {
	w.Log.Info("RegisterMutate mutate")
	mutates[getKeyByGVK(gvk)] = f
}

func getKeyByGVK(gvk metav1.GroupVersionKind) string {
	return fmt.Sprintf("%v-%v", gvk.Kind, gvk.Group)
}
