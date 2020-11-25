package pod

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	coreV1 "k8s.io/api/core/v1"
	"operator-demo/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PodMutate(curr []byte, old []byte, op v1.Operation, c client.Client, log logr.Logger) ([]webhook.PatchOperation, error) {
	if op == v1.Create || op == v1.Update {
		currPod := &coreV1.Pod{}
		err := json.Unmarshal(curr, currPod)
		if err != nil {
			return nil, fmt.Errorf("kind not match %v", err.Error())
		}
		fmt.Println(currPod)
	}
	return nil, nil
}
