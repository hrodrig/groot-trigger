package jobs

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/hrodrig/groot-trigger/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	j, err := cs.BatchV1().Jobs("groot").Get(ctx, res.JobName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := j.Spec.Template.Spec.Containers[0].Args
	if !containsPair(args, "--message", "note") {
		t.Fatalf("args missing --message note: %v", args)
	}
	for _, e := range j.Spec.Template.Spec.Containers[0].Env {
		if e.Name == "GROOT_TRIGGER_MESSAGE" {
			t.Fatal("must not set unused GROOT_TRIGGER_MESSAGE")
		}
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
	if got := dns1123("!!!"); got != "groot-collect" {
		t.Fatal(got)
	}
	long := strings.Repeat("a", 80)
	if len(dns1123(long)) > 63 {
		t.Fatal("too long")
	}
}

func TestErrBusyError(t *testing.T) {
	e := &ErrBusy{JobName: "j1"}
	if e.Error() == "" {
		t.Fatal("empty")
	}
}

func TestJobActiveVariants(t *testing.T) {
	if jobActive(&batchv1.Job{Status: batchv1.JobStatus{Failed: 1}}) {
		t.Fatal("failed")
	}
	if !jobActive(&batchv1.Job{Status: batchv1.JobStatus{Active: 2}}) {
		t.Fatal("active")
	}
	complete := &batchv1.Job{Status: batchv1.JobStatus{
		Conditions: []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}},
	}}
	if jobActive(complete) {
		t.Fatal("complete cond")
	}
	pending := &batchv1.Job{Status: batchv1.JobStatus{}}
	if !jobActive(pending) {
		t.Fatal("pending should be active")
	}
}

func TestCurrentNamespace(t *testing.T) {
	old := readNamespaceFile
	t.Cleanup(func() { readNamespaceFile = old })
	readNamespaceFile = func() ([]byte, error) { return nil, os.ErrNotExist }
	ns, err := currentNamespace()
	if err != nil || ns != "default" {
		t.Fatalf("%q %v", ns, err)
	}
	readNamespaceFile = func() ([]byte, error) { return []byte("  groot\n"), nil }
	ns, err = currentNamespace()
	if err != nil || ns != "groot" {
		t.Fatalf("%q %v", ns, err)
	}
	readNamespaceFile = func() ([]byte, error) { return []byte("   "), nil }
	ns, err = currentNamespace()
	if err != nil || ns != "default" {
		t.Fatalf("%q %v", ns, err)
	}
}

func TestUnavailable(t *testing.T) {
	u := Unavailable(errors.New("no k8s"))
	if _, _, err := u.ActiveJob(context.Background()); err == nil {
		t.Fatal("active")
	}
	if _, err := u.Create(context.Background(), "id", ""); err == nil {
		t.Fatal("create")
	}
}

func TestCreateWithPVCAndSecret(t *testing.T) {
	cs := fake.NewSimpleClientset()
	s := &K8sStarter{
		Client: cs,
		Cfg: config.Config{
			GrootImage:         "ghcr.io/hrodrig/groot:v1.1.1",
			GrootConfigMap:     "groot-config",
			GrootConfigKey:     "groot.yml",
			GrootJobSA:         "groot",
			GrootOutPVC:        "groot-out",
			GrootEnvFromSecret: "groot-s3",
			GrootExtraArgs:     []string{"--verbose"},
			JobTTLSeconds:      60,
		},
		NS: "groot",
	}
	res, err := s.Create(context.Background(), "aabbccdd", "m")
	if err != nil {
		t.Fatal(err)
	}
	j, err := cs.BatchV1().Jobs("groot").Get(context.Background(), res.JobName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	c := j.Spec.Template.Spec.Containers[0]
	if len(c.EnvFrom) != 1 || c.EnvFrom[0].SecretRef.Name != "groot-s3" {
		t.Fatalf("envFrom: %+v", c.EnvFrom)
	}
	if !containsPair(c.Args, "--message", "m") {
		t.Fatalf("args: %v", c.Args)
	}
	foundVerbose := false
	for _, a := range c.Args {
		if a == "--verbose" {
			foundVerbose = true
		}
	}
	if !foundVerbose {
		t.Fatalf("missing --verbose in %v", c.Args)
	}
	if j.Spec.Template.Spec.Volumes[1].PersistentVolumeClaim == nil {
		t.Fatal("expected pvc volume")
	}
}

func TestCreateJobReadOnlyRootfs(t *testing.T) {
	cs := fake.NewSimpleClientset()
	s := &K8sStarter{
		Client: cs,
		Cfg: config.Config{
			GrootImage:     "ghcr.io/hrodrig/groot:v1.1.1",
			GrootConfigMap: "groot-config",
			GrootConfigKey: "groot.yml",
			GrootJobSA:     "groot",
			JobTTLSeconds:  60,
		},
		NS: "groot",
	}
	res, err := s.Create(context.Background(), "aabbccdd", "")
	if err != nil {
		t.Fatal(err)
	}
	j, err := cs.BatchV1().Jobs("groot").Get(context.Background(), res.JobName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	c := j.Spec.Template.Spec.Containers[0]
	sc := c.SecurityContext
	if sc == nil || sc.ReadOnlyRootFilesystem == nil || !*sc.ReadOnlyRootFilesystem {
		t.Fatal("expected readOnlyRootFilesystem")
	}
	if sc.RunAsUser == nil || *sc.RunAsUser != 65532 {
		t.Fatalf("runAsUser: %+v", sc.RunAsUser)
	}
	foundTmp := false
	for _, m := range c.VolumeMounts {
		if m.Name == "tmp" && m.MountPath == "/tmp" {
			foundTmp = true
		}
	}
	if !foundTmp {
		t.Fatal("expected /tmp emptyDir mount")
	}
}

func TestNewInClusterFailsOutside(t *testing.T) {
	_, err := NewInCluster(config.Config{})
	if err == nil {
		t.Fatal("expected in-cluster config error outside cluster")
	}
}

func TestCreateOmitsEmptyMessage(t *testing.T) {
	cs := fake.NewSimpleClientset()
	s := &K8sStarter{
		Client: cs,
		Cfg: config.Config{
			GrootImage:     "ghcr.io/hrodrig/groot:v1.1.1",
			GrootConfigMap: "groot-config",
			GrootConfigKey: "groot.yml",
			GrootJobSA:     "groot",
			JobTTLSeconds:  60,
		},
		NS: "groot",
	}
	res, err := s.Create(context.Background(), "aabbccdd", "")
	if err != nil {
		t.Fatal(err)
	}
	j, err := cs.BatchV1().Jobs("groot").Get(context.Background(), res.JobName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range j.Spec.Template.Spec.Containers[0].Args {
		if a == "--message" {
			t.Fatal("empty message must not add --message")
		}
	}
}

func containsPair(args []string, flag, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}
