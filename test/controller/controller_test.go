package controller

import (
	"context"
	"os"
	"testing"
	"time"

	"k8s.io/api/node/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	cv1 "github.com/controller-e2e-framework/test-1-controller/api/v1alpha1"
	rv1 "github.com/controller-e2e-framework/test-2-controller/api/v1alpha1"

	"github.com/controller-e2e-framework/e2e-framework/shared/steps/assess"
	"github.com/controller-e2e-framework/e2e-framework/shared/steps/setup"
)

func TestControllerApply(t *testing.T) {
	t.Log("running component version apply")

	feature := features.New("Custom Controller").
		Setup(setup.AddSchemeAndNamespace(cv1.AddToScheme, namespace)).
		Setup(setup.AddSchemeAndNamespace(rv1.AddToScheme, namespace)).
		Setup(setup.ApplyTestData(v1alpha1.AddToScheme, namespace, "*")).
		Assess("check if controller was created", assess.ResourceWasCreated("controller-sample", namespace, &cv1.Controller{})).
		Assess("check if responder was created", assess.ResourceWasCreated("responder-sample", namespace, &rv1.Responder{})).
		Assess("wait for responder condition to be true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			cv := &rv1.Responder{
				ObjectMeta: metav1.ObjectMeta{Name: "responder-sample", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(cv, func(object k8s.Object) bool {
				obj, ok := object.(*rv1.Responder)
				if !ok {
					return false
				}

				labels := obj.GetLabels()

				if v, ok := labels["controller-e2e-framework.controlled"]; ok && v == "true" {
					return true
				}

				return false
			}), wait.WithTimeout(time.Minute*2))

			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("wait for responder acquired condition to be true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			client, err := cfg.NewClient()
			if err != nil {
				t.Fail()
			}

			cv := &rv1.Responder{
				ObjectMeta: metav1.ObjectMeta{Name: "responder-sample", Namespace: cfg.Namespace()},
			}

			// wait for component version to be reconciled
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(cv, func(object k8s.Object) bool {
				obj, ok := object.(*rv1.Responder)
				if !ok {
					return false
				}

				return obj.Status.Acquired
			}), wait.WithTimeout(time.Minute*2))

			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Helper()
			t.Log("teardown")

			// remove test resources before exiting
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			if err := decoder.DecodeEachFile(ctx, os.DirFS("./testdata"), "*",
				decoder.DeleteHandler(r),           // try to DELETE objects after decoding
				decoder.MutateNamespace(namespace), // inject a namespace into decoded objects, before calling DeleteHandler
			); err != nil {
				t.Fatal(err)
			}

			t.Log("teardown done")

			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
