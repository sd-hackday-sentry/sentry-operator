package sentry

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// deployment for the sentry web process
func (r *ReconcileSentry) deploymentForSentryWebUI() *appsv1.Deployment {
	name := "sentry-web-ui"
	replicas := int32(r.sentry.Spec.SentryWebReplicas)
	sentryPort := int32(9000)

	opts := templateOpts{
		Name: name,
		Args: []string{
			"run",
			"web",
		},
		ExtraEnv: []corev1.EnvVar{
			{
				Name:  "SENTRY_WEB_PORT",
				Value: fmt.Sprintf("%d", sentryPort),
			},
			{
				Name:  "SENTRY_WEB_HOST",
				Value: "0.0.0.0",
			},
		},
		ContainerPorts: []corev1.ContainerPort{
			{
				ContainerPort: sentryPort,
				Protocol:      "TCP",
			},
		},
		LivenessProbe: &corev1.Probe{
			InitialDelaySeconds: int32(3),
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/_health",
					Port: intstr.IntOrString{
						IntVal: sentryPort,
					},
					Scheme: "HTTP",
				},
			},
			PeriodSeconds: int32(3),
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.sentry.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: r.getCommonPodTemplate(opts),
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(r.sentry, dep, r.scheme)
	return dep
}

// deployment for the sentry worker process
func (r *ReconcileSentry) deploymentForSentryWorker() *appsv1.Deployment {
	name := "sentry-worker"
	replicas := int32(r.sentry.Spec.SentryWorkers)
	opts := templateOpts{
		Name: name,
		Args: []string{
			"run",
			"worker",
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.sentry.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: r.getCommonPodTemplate(opts),
		},
	}

	controllerutil.SetControllerReference(r.sentry, dep, r.scheme)
	return dep
}

// deployment for the sentry cron process
func (r *ReconcileSentry) deploymentForSentryCron() *appsv1.Deployment {
	name := "sentry-cron"
	replicas := int32(1)
	opts := templateOpts{
		Name: name,
		Args: []string{
			"run",
			"cron",
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.sentry.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: r.getCommonPodTemplate(opts),
		},
	}

	controllerutil.SetControllerReference(r.sentry, dep, r.scheme)
	return dep
}
