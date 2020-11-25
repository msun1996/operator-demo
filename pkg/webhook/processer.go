package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
)

// WebhookProcessor handles the http portion of a request prior to handing to an admit
func (wh Webhook) webhookProcessor(w http.ResponseWriter, r *http.Request, h WebhookHandler) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		wh.Log.Error(fmt.Errorf("contentType=%s, expect application/json", contentType), "content type error")
		return
	}

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1.AdmissionReview{}
	responseAdmissionReview.Kind = "AdmissionReview"
	responseAdmissionReview.APIVersion = "admission.k8s.io/v1"

	deserializer := scheme.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		wh.Log.Error(err, "deserializer error")
		responseAdmissionReview.Response = toAdmissionErrResponse(err)
	} else {
		wh.Log.V(1).Info(fmt.Sprintf("process request: %v, op:%v", requestedAdmissionReview.Request.Object.Raw, requestedAdmissionReview.Request.Operation))
		// pass to admitFunc
		responseAdmissionReview.Response = h(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	wh.Log.Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		wh.Log.Error(err, "marshal error")
	}
	if _, err := w.Write(respBytes); err != nil {
		wh.Log.Error(err, "write data error")
	}
}

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
func toAdmissionErrResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func (w Webhook) admitFuncProcessor(ar v1.AdmissionReview) *v1.AdmissionResponse {
	if ar.Request.SubResource != "" {
		return &v1.AdmissionResponse{Allowed: true}
	}

	var wf WebhookFunc
	if f, ok := admits[getKeyByGVK(ar.Request.Kind)]; ok {
		wf = f
	} else {
		return &v1.AdmissionResponse{Allowed: true}
	}

	w.Log.Info(fmt.Sprintf("admitFuncProcessor handling request: %s", ar.Request.Kind))

	p, err := w.Scheme.New(schema.GroupVersionKind{
		Group:   ar.Request.Kind.Group,
		Version: ar.Request.Kind.Version,
		Kind:    ar.Request.Kind.Kind,
	})
	if err != nil {
		toAdmissionErrResponse(err)
	}
	err = json.Unmarshal(ar.Request.Object.Raw, &p)
	if err != nil {
		toAdmissionErrResponse(err)
	}

	w.Log.V(1).Info("admit func processor", "obj", ar.Request.Object.Raw)
	if _, err := wf(ar.Request.Object.Raw, ar.Request.OldObject.Raw, ar.Request.Operation, w.Client, w.Log); err != nil {
		return toAdmissionErrResponse(err)
	}

	return &v1.AdmissionResponse{
		Allowed: true,
	}
}

func (w Webhook) mutateFuncProcessor(ar v1.AdmissionReview) *v1.AdmissionResponse {
	if ar.Request.SubResource != "" {
		return &v1.AdmissionResponse{Allowed: true}
	}

	if ar.Request.Operation == v1.Delete {
		return &v1.AdmissionResponse{Allowed: true}
	}

	var wf WebhookFunc
	if f, ok := mutates[getKeyByGVK(ar.Request.Kind)]; ok {
		wf = f
	} else {
		return &v1.AdmissionResponse{Allowed: true}
	}

	w.Log.Info(fmt.Sprintf("mutateFuncProcessor handling request: %s", ar.Request.Kind))

	var patchs []PatchOperation
	var err error
	if patchs, err = wf(ar.Request.Object.Raw, ar.Request.OldObject.Raw, ar.Request.Operation, w.Client, w.Log); err != nil {
		return &v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	patchBytes, err := json.Marshal(patchs)
	w.Log.Info("apply patchs", "patchs", string(patchBytes))
	if err != nil {
		return &v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &v1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}

}
