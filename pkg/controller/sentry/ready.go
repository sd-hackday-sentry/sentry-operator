package sentry

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileSentry) checkIfJobIsCompleted(name string) func() (bool, error) {
	return func() (bool, error) {
		r.logger.Info(fmt.Sprintf("checking if job '%s' has completed", name))
		ns := r.sentry.Namespace
		job := &batchv1.Job{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, job)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, fmt.Errorf("the job '%s' was not found in namespace '%s'", name, ns)
			}
			return false, err
		}
		for _, c := range job.Status.Conditions {
			if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	}
}
