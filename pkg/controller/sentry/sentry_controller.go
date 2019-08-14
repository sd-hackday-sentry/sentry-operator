package sentry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	v1alpha1 "github.com/sd-hackday-sentry/sentry-operator/pkg/apis/sentry/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("sentry")

// Add creates a new Sentry Controller and adds it to the Manager. The Manager
// will set fields on the Controller and Start it when the Manager is Started.
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
	client  client.Client
	scheme  *runtime.Scheme
	logger  logr.Logger
	secrets map[string]string
	sentry  *v1alpha1.Sentry
}

func (r *ReconcileSentry) validateSecrets() error {
	secretName := r.sentry.Spec.SentrySecret
	ns := r.sentry.ObjectMeta.Namespace
	secret := &corev1.Secret{}

	r.logger.Info(fmt.Sprintf("loading secrets from '%s'", secretName))

	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: secretName}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("the provided secret '%s' was not found in namespace '%s'", secretName, ns)
		}
		return err
	}

	// load and validate required secrets
	errors := []string{}
	required := []string{
		r.sentry.Spec.SentrySecretKeyKey,
		r.sentry.Spec.PostgresPasswordKey,
		r.sentry.Spec.SentrySuperUserEmailKey,
		r.sentry.Spec.SentrySuperUserPasswordKey,
	}
	for _, secretKey := range required {
		if _, ok := secret.Data[secretKey]; !ok {
			errors = append(errors, fmt.Sprintf("key '%s' is missing", secretKey))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors found when loading values from secret '%s': %s", secretName, strings.Join(errors, ", "))
	}

	return nil
}

// Reconcile reads that state of the cluster for a Sentry object and makes changes based on the state read
// and what is in the Sentry.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSentry) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.logger = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.logger.Info("Reconciling Sentry")

	// Fetch the Sentry instance
	r.sentry = &v1alpha1.Sentry{}
	err := r.client.Get(context.TODO(), request.NamespacedName, r.sentry)

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

	r.sentry.SetDefaults()
	if err := r.validateSecrets(); err != nil {
		return reconcile.Result{}, err
	}

	requeue := false

	var allJobs = map[string]func(string) *batchv1.Job{
		"sentry-upgrader":   r.jobForSentryUpgrader,
		"sentry-createuser": r.jobForSentryCreateUser,
	}

	for name, f := range allJobs {
		job := &batchv1.Job{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: r.sentry.Namespace}, job)
		if err != nil && errors.IsNotFound(err) {
			requeue = true
			job = f(name)
			r.logger.Info("Creating a new Job.", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			err = r.client.Create(context.TODO(), job)
			if err != nil {
				r.logger.Error(err, "Failed to create new Job.", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
				return reconcile.Result{}, err
			}
			// we want to wait until the job has run before proceeding
			err = wait.PollUntil(5*time.Second, r.checkIfJobIsCompleted(name), context.TODO().Done())
			if err != nil {
				r.logger.Error(err, "Timed out waiting for job.", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
				return reconcile.Result{}, err
			}
		} else if err != nil {
			r.logger.Error(err, "Failed to get Job.", "Job.Name", name)
			return reconcile.Result{}, err
		} else {
			// TODO: logic to run the upgrader job between versions should be here
			r.logger.Info("Job already exists, not running again", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		}
	}

	var allDeployments = map[string]func(string) *appsv1.Deployment{
		"sentry-web-ui": r.deploymentForSentryWebUI,
		"sentry-worker": r.deploymentForSentryWorker,
		"sentry-cron":   r.deploymentForSentryCron,
	}

	for name, f := range allDeployments {
		// Check if the deployment already exists, if not create a new one
		found := &appsv1.Deployment{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: r.sentry.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			requeue = true
			var dep *appsv1.Deployment = f(name)
			r.logger.Info("Creating a new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.client.Create(context.TODO(), dep)
			if err != nil {
				r.logger.Error(err, "Failed to create new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return reconcile.Result{}, err
			}
		} else if err != nil {
			r.logger.Error(err, "Failed to get Deployment.", "Deployment.Name", name)
			return reconcile.Result{}, err
		} else {
			r.logger.Info("Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		}
	}

	//expose the web service
	found := &corev1.Service{}
	name := "sentry-web-ui"
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: r.sentry.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		requeue = true
		svc := r.serviceForSentryWebUI(name)
		r.logger.Info("Creating a new Service.", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			r.logger.Error(err, "Failed to create new Service.", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		r.logger.Error(err, "Failed to get Service.", "Service.Name", name)
		return reconcile.Result{}, err
	} else {
		r.logger.Info("Service already exists", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
	}

	return reconcile.Result{Requeue: requeue}, nil
}
