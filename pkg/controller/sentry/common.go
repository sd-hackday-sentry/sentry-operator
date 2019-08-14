package sentry

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type templateOpts struct {
	Name           string
	Args           []string
	ExtraEnv       []corev1.EnvVar
	RestartPolicy  *corev1.RestartPolicy
	ContainerPorts []corev1.ContainerPort
	LivenessProbe  *corev1.Probe
}

// returns a common pod template for the various jobs/deployments
func (r *ReconcileSentry) getCommonPodTemplate(opts templateOpts) corev1.PodTemplateSpec {
	labels := map[string]string{"app": opts.Name}
	env := []corev1.EnvVar{
		{
			Name:  "SENTRY_ENVIRONMENT",
			Value: r.sentry.Spec.SentryEnvironment,
		},
		{
			Name: "SENTRY_SECRET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.sentry.Spec.SentrySecret,
					},
					Key: r.sentry.Spec.SentrySecretKeyKey,
				},
			},
		},
		{
			Name:  "SENTRY_POSTGRES_HOST",
			Value: r.sentry.Spec.PostgresHost,
		},
		{
			Name:  "SENTRY_POSTGRES_PORT",
			Value: fmt.Sprintf("%d", r.sentry.Spec.PostgresPort),
		},
		{
			Name:  "SENTRY_DB_NAME",
			Value: r.sentry.Spec.PostgresDB,
		},
		{
			Name:  "SENTRY_DB_USER",
			Value: r.sentry.Spec.PostgresUser,
		},
		{
			Name: "SENTRY_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.sentry.Spec.SentrySecret,
					},
					Key: r.sentry.Spec.PostgresPasswordKey,
				},
			},
		},
		{
			Name:  "SENTRY_REDIS_HOST",
			Value: r.sentry.Spec.RedisHost,
		},
		{
			Name:  "SENTRY_REDIS_PORT",
			Value: fmt.Sprintf("%d", r.sentry.Spec.RedisPort),
		},
		{
			Name:  "SENTRY_REDIS_DB",
			Value: r.sentry.Spec.RedisDB,
		},
		{
			Name:  "C_FORCE_ROOT",
			Value: "true",
		},
	}
	if len(opts.ExtraEnv) > 0 {
		env = append(env, opts.ExtraEnv...)
	}
	restartPolicy := corev1.RestartPolicyAlways
	if opts.RestartPolicy != nil {
		restartPolicy = *opts.RestartPolicy
	}
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image:           r.sentry.Spec.SentryImage,
				Name:            opts.Name,
				Args:            opts.Args,
				Env:             env,
				ImagePullPolicy: corev1.PullAlways,
				Ports:           opts.ContainerPorts,
				LivenessProbe:   opts.LivenessProbe,
			}},
			RestartPolicy: restartPolicy,
		},
	}
	return podTemplate
}
