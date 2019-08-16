package sentry

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// all jobs should be restarted on failure only
var onFailure corev1.RestartPolicy = corev1.RestartPolicyOnFailure

// job for the sentry upgrade process
func (r *ReconcileSentry) jobForSentryUpgrader() *batchv1.Job {
	name := "sentry-upgrader"
	opts := templateOpts{
		Name: name,
		Args: []string{
			"upgrade",
			"--noinput",
		},
		RestartPolicy: &onFailure,
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.sentry.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: r.getCommonPodTemplate(opts),
		},
	}

	controllerutil.SetControllerReference(r.sentry, job, r.scheme)
	return job
}

// job for the sentry createuser process
func (r *ReconcileSentry) jobForSentryCreateUser() *batchv1.Job {
	name := "sentry-createuser"
	opts := templateOpts{
		Name: name,
		Args: []string{
			"createuser",
			"--no-input",
			"--superuser",
			"--email",
			"$(SENTRY_SU_EMAIL)",
			"--password",
			"$(SENTRY_SU_PASSWORD)",
		},
		ExtraEnv: []corev1.EnvVar{
			{
				Name: "SENTRY_SU_EMAIL",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: r.sentry.Spec.SentrySecret,
						},
						Key: r.sentry.Spec.SentrySuperUserEmailKey,
					},
				},
			},
			{
				Name: "SENTRY_SU_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: r.sentry.Spec.SentrySecret,
						},
						Key: r.sentry.Spec.SentrySuperUserPasswordKey,
					},
				},
			},
		},
		RestartPolicy: &onFailure,
	}
	jobSpec := r.getCommonPodTemplate(opts)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.sentry.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: jobSpec,
		},
	}

	controllerutil.SetControllerReference(r.sentry, job, r.scheme)
	return job
}
