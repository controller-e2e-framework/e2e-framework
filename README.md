# E2E Framework

This repository contains the testing infrastructure for the controllers.
It contains e2e tests that run scenarios that involve multiple controllers
working together. It's built around reusable steps to minimize the amount of
testing code that has to be written.

## Architecture

This framework is built around [Tilt](https://tilt.dev) and [e2e-framework](https://github.com/kubernetes-sigs/e2e-framework) from Kubernetes.

### Tilt

Why is Tilt involved in all of this? Tilt was added to greatly simplify setting
up a dev environment. For a longer explanation read [this](https://skarlso.github.io/2023/02/25/rapid-controller-development-with-tilt/) wiki post.

It's a convenient way to set up the controllers and their dependencies such as
RBAC, CRDs, Deployment, patches, etc.

Alternatives were considered, such as building test images and pushing to local
registry. This was found suboptimal in cases of using CI vs. a local environment.
The problem is to get the manifest files. If we download them, you might be testing
the incorrect files locally when changing them.

Using Tilt you can be assured that you are testing the correct files and gives
a lot of flexibility in configuring the controllers.

### e2e-framework

This framework provides capabilities to interact with a cluster including
creating and destroying them using `kind`. Each suite contains a `TestMain` which
sets up things that the tests themselves will require. Such as:

- what controllers should be started
- port forward the registry
- ...

The framework provides clear setup and teardown functions like this one:

```go
func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	kindClusterName = envconf.RandomName("component-version", 32)
	namespace = envconf.RandomName("testing", 32)

	testEnv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.CreateNamespace(namespace),
		shared.RunTiltForControllers("test-1-controller", "test-2-controller"),
		shared.ForwardRegistry(),
	)

	testEnv.Finish(
		shared.ShutdownPortForward(),
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyKindCluster(kindClusterName),
	)

	os.Exit(testEnv.Run(m))
}
```

Further Setup and Teardown functions can be added under `shared/`.

Details about writing tests can be read under [Implementation](#implementations).

## Running the tests

### Prerequisites and project structure

Any environment MUST have `tilt` and `kind` installed. To install them, run `make prepare`.

Another requirement is the controllers that are used in the test MUST be checked out next to
this project. The testing framework will gradually do `cd ..` to find any controllers it needs.

### Simply `go test`

To run all tests, simply run `make test`. To run an individual suite, run:

`go test -v -count=1 ./test/component_version`

For a little while, it can seem like the test is hanging. This is the setup phase. You should
see activity once the test enters the running phase. In the meantime, you can monitor progress
with `kind get clusters`. `e2e-framework` sets up the kubernetes context to the kind cluster
so, you should also be able to see pods by running `kubectl get pods -A`.

### Parallel running

For now, this framework doesn't support running the suites in parallel and neither does e2e-framework.
It uses the kube context to set up context to the current cluster. There are some workarounds, but
doing that was not priority at the time of this writing. Tilt also can be told to use a different
kube config and to run on a different port. This is future work.

## Implementations

### Writing tests in a declarative manner

The tests should try to follow a narrative. At the time of this writing, only two, very basic
tests are available.

TODO: Fill this out with some concrete examples once we have them.

### Using shared steps

There a number of steps like, setup and assesses that can be used to construct a
test. If a step is not found under `shared/steps/*` consider adding it if you
think that it might be reused.

### Waiting for objects or conditions

To wait for an object to has certain property:

```go
wait.For(conditions.New(config.Client().Resources()).ResourceScaled(deployment, func(object k8s.Object) int32 {
    return object.(*appsv1.Deployment).Status.ReadyReplicas
}, 2))
```

To wait on a condition to be fulfilled:

```go
// wait for component version to be reconciled
err = wait.For(conditions.New(client.Resources()).ResourceMatch(cv, func(object k8s.Object) bool {
    cvObj := object.(*v1alpha1.ComponentVersion)
    return fconditions.IsTrue(cvObj, meta.ReadyCondition)
}), wait.WithTimeout(time.Minute*2))
```

### Adding Setup functions

### Using Context to share values between steps

The context in the `Assess` and `Setup` can be used to pass around certain values and
objects which the next Assess can use. For example, passing around a whole object:

```go
	createDeployment := features.New("Create Deployment").
		Assess("Create Nginx Deployment 1", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			deployment := newDeployment(namespace, "deployment-1", 2)
			err := config.Client().Resources().Create(ctx, deployment)
			if err != nil {
				t.Error("failed to create test pod for deployment-1")
			}
			ctx = context.WithValue(ctx, "DEPLOYMENT", deployment)
			return ctx
		}).
		Assess("Wait for Nginx Deployment 1 to be scaled up", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			deployment := ctx.Value("DEPLOYMENT").(*appsv1.Deployment)
			err := wait.For(conditions.New(config.Client().Resources()).ResourceScaled(deployment, func(object k8s.Object) int32 {
				return object.(*appsv1.Deployment).Status.ReadyReplicas
			}, 2))
			if err != nil {
				t.Error("failed waiting for deployment to be scaled up")
			}
			return ctx
		}).Feature()
```

Then, this feature can be executed together with others.

```go
	testEnv.TestInParallel(t, createDeployment, checkDeployment, deleteDeployment)
```
