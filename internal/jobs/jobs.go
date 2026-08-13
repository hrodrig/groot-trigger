// Package jobs creates and inspects groot collect Jobs.
package jobs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hrodrig/groot-trigger/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	LabelName   = "app.kubernetes.io/name"
	LabelPartOf = "app.kubernetes.io/part-of"
	LabelRunID  = "groot-trigger/run_id"
	NameValue   = "groot-trigger-collect"
	PartOfValue = "groot-trigger"
)

// Result of a successful create.
type Result struct {
	RunID   string
	JobName string
}

// ErrBusy means a collect Job is already Pending/Running.
type ErrBusy struct {
	JobName string
}

func (e *ErrBusy) Error() string {
	return fmt.Sprintf("collect in progress: %s", e.JobName)
}

// Starter creates collect Jobs and detects busy state.
type Starter interface {
	ActiveJob(ctx context.Context) (jobName string, busy bool, err error)
	Create(ctx context.Context, runID, message string) (Result, error)
}

// K8sStarter uses client-go.
type K8sStarter struct {
	Client kubernetes.Interface
	Cfg    config.Config
	NS     string
}

// NewInCluster builds a K8sStarter from in-cluster config.
func NewInCluster(cfg config.Config) (*K8sStarter, error) {
	rc, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return nil, err
	}
	ns := cfg.GrootNamespace
	if ns == "" {
		ns, err = currentNamespace()
		if err != nil {
			return nil, err
		}
	}
	return &K8sStarter{Client: cs, Cfg: cfg, NS: ns}, nil
}

// readNamespaceFile is overridden in tests.
var readNamespaceFile = func() ([]byte, error) {
	return os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
}

func currentNamespace() (string, error) {
	data, err := readNamespaceFile()
	if err != nil {
		return "default", nil
	}
	ns := strings.TrimSpace(string(data))
	if ns == "" {
		return "default", nil
	}
	return ns, nil
}

// Unavailable returns a Starter that always errors (e.g. no in-cluster config).
func Unavailable(err error) Starter {
	return unavailableStarter{err: err}
}

type unavailableStarter struct{ err error }

func (u unavailableStarter) ActiveJob(context.Context) (string, bool, error) {
	return "", false, u.err
}

func (u unavailableStarter) Create(context.Context, string, string) (Result, error) {
	return Result{}, u.err
}

// ActiveJob lists collect Jobs that are not finished.
func (s *K8sStarter) ActiveJob(ctx context.Context) (string, bool, error) {
	list, err := s.Client.BatchV1().Jobs(s.NS).List(ctx, metav1.ListOptions{
		LabelSelector: LabelName + "=" + NameValue,
	})
	if err != nil {
		return "", false, err
	}
	for i := range list.Items {
		j := &list.Items[i]
		if jobActive(j) {
			return j.Name, true, nil
		}
	}
	return "", false, nil
}

func jobActive(j *batchv1.Job) bool {
	if j.Status.Succeeded > 0 || j.Status.Failed > 0 {
		return false
	}
	// Pending or Running (Active > 0) or just created.
	if j.Status.Active > 0 {
		return true
	}
	for _, c := range j.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return false
		}
	}
	// No terminal condition and not succeeded/failed counts → treat as active.
	return true
}

// Create builds and submits a collect Job.
func (s *K8sStarter) Create(ctx context.Context, runID, message string) (Result, error) {
	if name, busy, err := s.ActiveJob(ctx); err != nil {
		return Result{}, err
	} else if busy {
		return Result{}, &ErrBusy{JobName: name}
	}
	short := runID
	if len(short) > 8 {
		short = short[:8]
	}
	jobName := dns1123("groot-collect-" + short)
	ttl := s.Cfg.JobTTLSeconds
	backoff := int32(0)
	args := []string{"collect", "--config", "/config/" + s.Cfg.GrootConfigKey}
	args = append(args, s.Cfg.GrootExtraArgs...)

	nonroot := int64(65532)
	ro := true
	no := false
	container := corev1.Container{
		Name:            "groot",
		Image:           s.Cfg.GrootImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args:            args,
		VolumeMounts: []corev1.VolumeMount{
			{Name: "config", MountPath: "/config", ReadOnly: true},
			{Name: "out", MountPath: "/out"},
			{Name: "tmp", MountPath: "/tmp"},
		},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: &no,
			ReadOnlyRootFilesystem:   &ro,
			RunAsNonRoot:             &ro,
			RunAsUser:                &nonroot,
			RunAsGroup:               &nonroot,
			Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		},
	}
	if message != "" {
		container.Env = append(container.Env, corev1.EnvVar{Name: "GROOT_TRIGGER_MESSAGE", Value: message})
	}
	if s.Cfg.GrootEnvFromSecret != "" {
		container.EnvFrom = []corev1.EnvFromSource{{
			SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: s.Cfg.GrootEnvFromSecret}},
		}}
	}

	volumes := []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: s.Cfg.GrootConfigMap},
				},
			},
		},
	}
	if s.Cfg.GrootOutPVC != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "out",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: s.Cfg.GrootOutPVC},
			},
		})
	} else {
		volumes = append(volumes, corev1.Volume{
			Name:         "out",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}
	volumes = append(volumes, corev1.Volume{
		Name:         "tmp",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
			Labels: map[string]string{
				LabelName:   NameValue,
				LabelPartOf: PartOfValue,
				LabelRunID:  runID,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &backoff,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelName:   NameValue,
						LabelPartOf: PartOfValue,
						LabelRunID:  runID,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: s.Cfg.GrootJobSA,
					RestartPolicy:      corev1.RestartPolicyNever,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &ro,
						RunAsUser:    &nonroot,
						RunAsGroup:   &nonroot,
						FSGroup:      &nonroot,
					},
					Containers: []corev1.Container{container},
					Volumes:    volumes,
				},
			},
		},
	}
	created, err := s.Client.BatchV1().Jobs(s.NS).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return Result{}, err
	}
	return Result{RunID: runID, JobName: created.Name}, nil
}

func dns1123(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	out = strings.Trim(out, "-")
	if len(out) > 63 {
		out = out[:63]
		out = strings.Trim(out, "-")
	}
	if out == "" {
		return "groot-collect"
	}
	return out
}
