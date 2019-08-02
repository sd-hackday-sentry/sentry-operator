package sentry

import (
	"context"
	"fmt"

	v1alpha1 "github.com/sd-hackday-sentry/sentry-operator/pkg/apis/sentry/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_sentry")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Sentry Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSentry{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sentry-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Sentry
	err = c.Watch(&source.Kind{Type: &v1alpha1.Sentry{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Sentry
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Sentry{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSentry implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSentry{}

// ReconcileSentry reconciles a Sentry object
type ReconcileSentry struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Sentry object and makes changes based on the state read
// and what is in the Sentry.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSentry) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Sentry")

	// Fetch the Sentry instance
	sentry := &v1alpha1.Sentry{}
	err := r.client.Get(context.TODO(), request.NamespacedName, sentry)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	requeue := false

	var allDeployments = map[string]func(*v1alpha1.Sentry, string) *appsv1.Deployment{
		"sentry-web-ui": r.deploymentForSentryWebUI,
		"sentry-worker": r.deploymentForSentryWorker,
		"sentry-cron":   r.deploymentForSentryCron,
	}

	for name, f := range allDeployments {
		// Check if the deployment already exists, if not create a new one
		found := &appsv1.Deployment{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: sentry.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			requeue = true
			var dep *appsv1.Deployment = f(sentry, name)
			reqLogger.Info("Creating a new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.client.Create(context.TODO(), dep)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return reconcile.Result{}, err
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to get Deployment.", "Deployment.Name", name)
			return reconcile.Result{}, err
		} else {
			reqLogger.Info("Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		}
	}

	return reconcile.Result{Requeue: requeue}, nil
}

func (r *ReconcileSentry) deploymentForSentryWebUI(m *v1alpha1.Sentry, name string) *appsv1.Deployment {
	labels := map[string]string{"app": name}
	var replicas int32 = 1
	var terminationGracePeriodSeconds int64 = 30

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: fmt.Sprintf("sentry:%s", m.Spec.SentryVersion),
						Name:  name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9000,
							Name:          name,
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "SENTRY_SECRET_KEY",
								Value: "my_secret_here_some_random_hash", // TODO get from config or generate
							},
							{
								Name:  "SENTRY_POSTGRES_HOST",
								Value: m.Spec.PostgresHost,
							},
							{
								Name:  "SENTRY_POSTGRES_PORT",
								Value: string(m.Spec.PostgresPort),
							},
							{
								Name:  "SENTRY_DB_NAME",
								Value: m.Spec.PostgresDB,
							},
							{
								Name:  "SENTRY_DB_USER",
								Value: m.Spec.PostgresUser,
							},
							{
								Name:  "SENTRY_DB_PASSWORD",
								Value: m.Spec.PostgresPassword,
							},
							{
								Name:  "SENTRY_REDIS_HOST",
								Value: m.Spec.RedisHost,
							},
							{
								Name:  "SENTRY_REDIS_PORT",
								Value: string(m.Spec.RedisPort),
							},
							{
								Name:  "SENTRY_REDIS_DB",
								Value: m.Spec.RedisDB,
							},
							{
								Name:  "SENTRY_USE_SSL",
								Value: "true", // TODO get from config?
							},
						},
						ImagePullPolicy: corev1.PullAlways,
					}},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					DNSPolicy:                     corev1.DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

func (r *ReconcileSentry) deploymentForSentryWorker(m *v1alpha1.Sentry, name string) *appsv1.Deployment {
	labels := map[string]string{"app": name}
	var replicas int32 = 1
	var terminationGracePeriodSeconds int64 = 30

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "sentry:latest",
						Name:  name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9000,
							Name:          name,
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "SENTRY_SECRET_KEY",
								Value: "my_secret_here_some_random_hash", // TODO get from config or generate
							},
							{
								Name:  "SENTRY_POSTGRES_HOST",
								Value: m.Spec.PostgresHost,
							},
							{
								Name:  "SENTRY_POSTGRES_PORT",
								Value: string(m.Spec.PostgresPort),
							},
							{
								Name:  "SENTRY_DB_NAME",
								Value: m.Spec.PostgresDB,
							},
							{
								Name:  "SENTRY_DB_USER",
								Value: m.Spec.PostgresUser,
							},
							{
								Name:  "SENTRY_DB_PASSWORD",
								Value: m.Spec.PostgresPassword,
							},
							{
								Name:  "SENTRY_REDIS_HOST",
								Value: m.Spec.RedisHost,
							},
							{
								Name:  "SENTRY_REDIS_PORT",
								Value: string(m.Spec.RedisPort),
							},
							{
								Name:  "SENTRY_REDIS_DB",
								Value: m.Spec.RedisDB,
							},
							{
								Name:  "C_FORCE_ROOT",
								Value: "true", // TODO get from config
							},
						},
						ImagePullPolicy: corev1.PullAlways,
					}},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					DNSPolicy:                     corev1.DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

func (r *ReconcileSentry) deploymentForSentryCron(m *v1alpha1.Sentry, name string) *appsv1.Deployment {
	labels := map[string]string{"app": name}
	var replicas int32 = 1
	var terminationGracePeriodSeconds int64 = 30

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "sentry:latest",
						Name:  name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9000,
							Name:          name,
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "SENTRY_SECRET_KEY",
								Value: "my_secret_here_some_random_hash", // TODO get from config or generate
							},
							{
								Name:  "SENTRY_POSTGRES_HOST",
								Value: m.Spec.PostgresHost,
							},
							{
								Name:  "SENTRY_POSTGRES_PORT",
								Value: string(m.Spec.PostgresPort),
							},
							{
								Name:  "SENTRY_DB_NAME",
								Value: m.Spec.PostgresDB,
							},
							{
								Name:  "SENTRY_DB_USER",
								Value: m.Spec.PostgresUser,
							},
							{
								Name:  "SENTRY_DB_PASSWORD",
								Value: m.Spec.PostgresPassword,
							},
							{
								Name:  "SENTRY_REDIS_HOST",
								Value: m.Spec.RedisHost,
							},
							{
								Name:  "SENTRY_REDIS_PORT",
								Value: string(m.Spec.RedisPort),
							},
							{
								Name:  "SENTRY_REDIS_DB",
								Value: m.Spec.RedisDB,
							},
							{
								Name:  "C_FORCE_ROOT",
								Value: "true", // TODO get from config
							},
						},
						ImagePullPolicy: corev1.PullAlways,
					}},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					DNSPolicy:                     corev1.DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}
