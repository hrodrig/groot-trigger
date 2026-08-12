package jobs

import (
	"context"
	"testing"

	"github.com/hrodrig/groot-trigger/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateAndBusy(t *testing.T) {
	cs := fake.NewSimpleClientset()
	s := &K8sStarter{
		Client: cs,
		Cfg: config.Config{
			GrootImage:     "ghcr.io/hrodrig/groot:v1.1.1",
			GrootConfigMap: "groot-config",
			GrootConfigKey: "groot.yml",
			GrootJobSA:     "groot",
			JobTTLSeconds:  3600,
		},
		NS: "groot",
	}
	ctx := context.Background()
	res, err := s.Create(ctx, "abcdef12-3456", "note")
	if err != nil {
		t.Fatal(err)
	}
	if res.JobName == "" || res.RunID == "" {
		t.Fatalf("result: %+v", res)
	}
	_, err = s.Create(ctx, "zzzzzzzz-9999", "")
	if err == nil {
		t.Fatal("expected busy")
	}
	if _, ok := err.(*ErrBusy); !ok {
		t.Fatalf("want ErrBusy, got %T %v", err, err)
	}
}

func TestActiveIgnoresSucceeded(t *testing.T) {
	cs := fake.NewSimpleClientset(&batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done",
			Namespace: "groot",
			Labels:    map[string]string{LabelName: NameValue},
		},
		Status: batchv1.JobStatus{Succeeded: 1},
	})
	s := &K8sStarter{Client: cs, Cfg: config.Config{JobTTLSeconds: 1}, NS: "groot"}
	_, busy, err := s.ActiveJob(context.Background())
	if err != nil || busy {
		t.Fatalf("busy=%v err=%v", busy, err)
	}
}

func TestDNS1123(t *testing.T) {
	if got := dns1123("Groot_Collect.ABC"); got != "groot-collect-abc" {
		t.Fatal(got)
	}
}
