package sentryoperator

import (
	"context"

	sentryoperatorv1 "github.com/sd-hackday-sentry/sentry-operator/pkg/apis/sentryoperator/v1"
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

var log = logf.Log.WithName("controller_sentryoperator")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new SentryOperator Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSentryOperator{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sentryoperator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SentryOperator
	err = c.Watch(&source.Kind{Type: &sentryoperatorv1.SentryOperator{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner SentryOperator
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &sentryoperatorv1.SentryOperator{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSentryOperator implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSentryOperator{}

// ReconcileSentryOperator reconciles a SentryOperator object
type ReconcileSentryOperator struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SentryOperator object and makes changes based on the state read
// and what is in the SentryOperator.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSentryOperator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SentryOperator")

	// Fetch the SentryOperator instance
	sentry := &sentryoperatorv1.SentryOperator{}
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

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sentry.Name, Namespace: sentry.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.newDeployment(sentry)
		reqLogger.Info("Creating a new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment.")
		return reconcile.Result{}, err
	}

	reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	return reconcile.Result{}, nil
}

// newDeployment returns a Sentry Deployment object
func (r *ReconcileSentryOperator) newDeployment(sentry *sentryoperatorv1.SentryOperator) *appsv1.Deployment {
	dep := &appsv1.Deployment{}
	// Set Sentry instance as the owner and controller
	controllerutil.SetControllerReference(sentry, dep, r.scheme)
	return dep
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
// TODO this shoudl be modified to provide a sentry depoyment
// we'll probably need one for each deployment in sentry-worker sentry-ui sentry-cron
func newPodForCR(cr *sentryoperatorv1.SentryOperator) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func (r *ReconcileSentryOperator) deploymentForSentryWebUI(m *sentryoperatorv1.Sentry) *appsv1.Deployment {
	name := "sentry-web-ui"

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: 1,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: name,
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
							corev1.EnvVar{
								Name:  "SENTRY_POSTGRES_HOST",
								Value: "db", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_SECRET_KEY",
								Value: "my_secret_here_some_random_hash", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_DB_USER",
								Value: "my_user", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_DB_PASSWORD",
								Value: "my_password", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_REDIS_HOST",
								Value: "my_ip", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_USE_SSL",
								Value: true, // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_EMAIL_HOST",
								Value: "smtp.example.com", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_SERVER_EMAIL",
								Value: "noreply@example.com", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_ACCESS",
								Value: "minio", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_BUCKET",
								Value: "sentry", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_HOST",
								Value: "'http://minio:9000'", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_SECRET",
								Value: "minio123", // TODO get from config
							},
						},
						PullPolicy: PullAlways,
					}},
					RestartPolicy:                 RestartPolicyAlways,
					TerminationGracePeriodSeconds: 30,
					DNSPolicy:                     DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

func (r *ReconcileSentryOperator) deploymentForSentryWorker(m *sentryoperatorv1.Sentry) *appsv1.Deployment {
	name := "sentry-worker"

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: 1,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: name,
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
							corev1.EnvVar{
								Name:  "SENTRY_POSTGRES_HOST",
								Value: "db", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_SECRET_KEY",
								Value: "my_secret_here_some_random_hash", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "C_FORCE_ROOT",
								Value: true, // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_DB_USER",
								Value: "my_user", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_DB_PASSWORD",
								Value: "my_password", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_SERVER_EMAIL",
								Value: "noreply@example.com", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_ACCESS",
								Value: "minio", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_BUCKET",
								Value: "sentry", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_HOST",
								Value: "'http://minio:9000'", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_SECRET",
								Value: "minio123", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_EMAIL_HOST",
								Value: "smtp.example.com", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_REDIS_HOST",
								Value: "my_ip", // TODO get from config
							},
						},
						PullPolicy: PullAlways,
					}},
					RestartPolicy:                 RestartPolicyAlways,
					TerminationGracePeriodSeconds: 30,
					DNSPolicy:                     DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

func (r *ReconcileSentryOperator) deploymentForSentryCron(m *sentryoperatorv1.Sentry) *appsv1.Deployment {
	name := "sentry-cron"

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: 1,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: name,
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
							corev1.EnvVar{
								Name:  "SENTRY_POSTGRES_HOST",
								Value: "db", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_SECRET_KEY",
								Value: "my_secret_here_some_random_hash", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_DB_USER",
								Value: "my_user",
							},
							corev1.EnvVar{
								Name:  "SENTRY_DB_PASSWORD",
								Value: "my_password", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_EMAIL_HOST",
								Value: "smtp.example.com", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_SERVER_EMAIL",
								Value: "noreply@example.com", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_ACCESS",
								Value: "minio", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_BUCKET",
								Value: "sentry", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_HOST",
								Value: "'http://minio:9000'", // TODO get from config
							},
							corev1.EnvVar{
								Name:  "SENTRY_FILE_SECRET",
								Value: "minio123", // TODO get from config
							},
						},
						PullPolicy: PullAlways,
					}},
					RestartPolicy:                 RestartPolicyAlways,
					TerminationGracePeriodSeconds: 30,
					DNSPolicy:                     DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}
