package sentry

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// service for the sentry web process
func (r *ReconcileSentry) serviceForSentryWebUI(name string) *corev1.Service {
	labels := map[string]string{"app": name}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      name,
			Namespace: r.sentry.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "sentry-http",
					Port:     9000,
					Protocol: "TCP",
				},
			},
		},
	}

	return svc
}
