package fixtures

import (
	"bytes"
	"io"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Every fixture must decode into the typed objects it claims to be, carry
// the names this package publishes, and pin its images: these are the facts
// a lane relies on at minute ninety of a run, checked here in a second.

// decodeAll splits a multi-document manifest and decodes each document into
// a typed object by kind, failing the test on anything malformed or unknown.
func decodeAll(t *testing.T, manifest []byte) []runtime.Object {
	t.Helper()
	var objects []runtime.Object
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
	for {
		var u unstructured.Unstructured
		if err := decoder.Decode(&u); err != nil {
			if err == io.EOF {
				return objects
			}
			t.Fatalf("decoding a fixture document: %v", err)
		}
		if len(u.Object) == 0 {
			continue
		}
		var typed runtime.Object
		switch u.GetKind() {
		case "Namespace":
			typed = &corev1.Namespace{}
		case "Deployment":
			typed = &appsv1.Deployment{}
		case "Service":
			typed = &corev1.Service{}
		case "Job":
			typed = &batchv1.Job{}
		default:
			t.Fatalf("fixture carries a kind the lanes do not expect: %s", u.GetKind())
		}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, typed); err != nil {
			t.Fatalf("decoding %s %s into its typed object: %v", u.GetKind(), u.GetName(), err)
		}
		objects = append(objects, typed)
	}
}

// images collects every container image a fixture runs.
func images(objects []runtime.Object) []string {
	var out []string
	for _, o := range objects {
		switch typed := o.(type) {
		case *appsv1.Deployment:
			for _, c := range typed.Spec.Template.Spec.Containers {
				out = append(out, c.Image)
			}
		case *batchv1.Job:
			for _, c := range typed.Spec.Template.Spec.Containers {
				out = append(out, c.Image)
			}
		}
	}
	return out
}

func requirePinned(t *testing.T, image string) {
	t.Helper()
	tag := image[strings.LastIndex(image, ":")+1:]
	if !strings.Contains(image, ":") || tag == "latest" || tag == "" {
		t.Errorf("fixture image must be pinned to a tag, got %q", image)
	}
}

func TestMinIOFixture(t *testing.T) {
	objects := decodeAll(t, MinIO())
	var ns, deploy, svc, job bool
	for _, o := range objects {
		switch typed := o.(type) {
		case *corev1.Namespace:
			ns = typed.Name == MinIONamespace
		case *appsv1.Deployment:
			deploy = typed.Namespace == MinIONamespace
			env := map[string]string{}
			for _, e := range typed.Spec.Template.Spec.Containers[0].Env {
				env[e.Name] = e.Value
			}
			if env["MINIO_ROOT_USER"] != MinIOAccessKey || env["MINIO_ROOT_PASSWORD"] != MinIOSecretKey {
				t.Errorf("the store's credentials must be the ones this package publishes, got %v", env)
			}
		case *corev1.Service:
			svc = typed.Namespace == MinIONamespace && typed.Name == "minio" && typed.Spec.Ports[0].Port == 9000
		case *batchv1.Job:
			job = typed.Namespace == MinIONamespace && typed.Name == MinIOBucketJob
			args := strings.Join(typed.Spec.Template.Spec.Containers[0].Args, " ")
			for _, want := range []string{MinIOEndpoint, MinIOBucket, MinIOAccessKey, MinIOSecretKey} {
				if !strings.Contains(args, want) {
					t.Errorf("the bucket Job must use %q, got %q", want, args)
				}
			}
		}
	}
	if !ns || !deploy || !svc || !job {
		t.Fatalf("the MinIO fixture must carry its Namespace, Deployment, Service, and bucket Job under %s "+
			"(ns=%v deploy=%v svc=%v job=%v)", MinIONamespace, ns, deploy, svc, job)
	}
	for _, image := range images(objects) {
		requirePinned(t, image)
	}
}

func TestKeyHolderFixture(t *testing.T) {
	objects := decodeAll(t, KeyHolder())
	var ns, deploy, svc bool
	for _, o := range objects {
		switch typed := o.(type) {
		case *corev1.Namespace:
			ns = typed.Name == KeyHolderNamespace
		case *appsv1.Deployment:
			deploy = typed.Namespace == KeyHolderNamespace &&
				typed.Spec.Template.Labels["app"] == strings.TrimPrefix(KeyHolderPodSelector, "app=")
			container := typed.Spec.Template.Spec.Containers[0]
			if !strings.HasPrefix(container.Image, "quay.io/openbao/openbao:") {
				t.Errorf("the key holder must run the OpenBao image line the operator's chart runs, got %q", container.Image)
			}
			var root string
			for _, e := range container.Env {
				if e.Name == "BAO_DEV_ROOT_TOKEN_ID" {
					root = e.Value
				}
			}
			if root != KeyHolderRootToken {
				t.Errorf("the key holder's root token must be the one this package publishes, got %q", root)
			}
		case *corev1.Service:
			svc = typed.Namespace == KeyHolderNamespace && typed.Name == "key-holder" && typed.Spec.Ports[0].Port == 8200
		}
	}
	if !ns || !deploy || !svc {
		t.Fatalf("the key holder fixture must carry its Namespace, Deployment, and Service under %s (ns=%v deploy=%v svc=%v)",
			KeyHolderNamespace, ns, deploy, svc)
	}
	for _, image := range images(objects) {
		requirePinned(t, image)
	}
}
