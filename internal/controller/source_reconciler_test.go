package controller

import (
	"testing"

	pulseticv1 "github.com/clevyr/pulsetic-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFindMonitorsScopedToNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := pulseticv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	newMonitor := func(namespace string) *pulseticv1.Monitor {
		return &pulseticv1.Monitor{
			ObjectMeta: metav1.ObjectMeta{Name: "wordpress", Namespace: namespace},
			Status: pulseticv1.MonitorStatus{
				SourceRef: &corev1.TypedLocalObjectReference{Kind: "HTTPRoute", Name: "wordpress"},
			},
		}
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithIndex(&pulseticv1.Monitor{}, SourceRefIndex, indexSourceRef).
		WithObjects(newMonitor("a"), newMonitor("b")).
		WithStatusSubresource(&pulseticv1.Monitor{}).
		Build()
	r := &SourceReconciler{Client: c}

	for _, namespace := range []string{"a", "b"} {
		list, err := r.findMonitors(t.Context(), "HTTPRoute", namespace, "wordpress")
		if err != nil {
			t.Fatal(err)
		}
		if len(list.Items) != 1 || list.Items[0].Namespace != namespace {
			t.Fatalf("namespace %q: got %d monitors, want only the one in %q", namespace, len(list.Items), namespace)
		}
	}

	list, err := r.findMonitors(t.Context(), "HTTPRoute", "c", "wordpress")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 0 {
		t.Fatalf("namespace %q: got %d monitors, want none", "c", len(list.Items))
	}
}
