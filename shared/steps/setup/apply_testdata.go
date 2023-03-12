package setup

import (
	"context"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// ApplyTestData takes a pattern and applies that from a testdata location.
func ApplyTestData(addToSchemeFunc func(s *runtime.Scheme) error, namespace, pattern string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		t.Helper()
		t.Log("in setup phase")

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}

		if err := decoder.DecodeEachFile(
			ctx, os.DirFS("./testdata"), pattern,
			decoder.CreateHandler(r),
			decoder.MutateNamespace(namespace),
		); err != nil {
			t.Fail()
		}

		t.Log("set up is done, component version should have been applied")

		return ctx
	}
}
