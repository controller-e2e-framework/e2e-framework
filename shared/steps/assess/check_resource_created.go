package assess

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// ResourceWasCreated is an assess step to check if a given resource was created.
func ResourceWasCreated(name, namespace string, obj k8s.Object) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		t.Helper()
		t.Log("check if resources are created")

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}

		if err := r.Get(ctx, name, namespace, obj); err != nil {
			t.Fail()
		}

		t.Log("resource successfully created")

		return ctx
	}
}
