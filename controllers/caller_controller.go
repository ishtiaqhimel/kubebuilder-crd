/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1 "ishtiaq.com/caller/api/v1"
)

// CallerReconciler reconciles a Caller object
type CallerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.ishtiaq.com,resources=callers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.ishtiaq.com,resources=callers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.ishtiaq.com,resources=callers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Caller object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CallerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("From Reconciler...", "Name", req.Name)

	caller := apiv1.Caller{}
	if err := r.Get(ctx, req.NamespacedName, &caller); err != nil {
		log.Error(err, "Unable to fetch caller")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cronJobList := batchv1.CronJobList{}
	if err := r.List(ctx, &cronJobList, client.InNamespace(req.Namespace), client.MatchingFields{cronJobOwnerKey: req.Name}); err != nil {
		log.Error(err, "Unable to list Child CronJob")
		return ctrl.Result{}, err
	}

	log.Info("Number of Running CronJobs", "Running", len(cronJobList.Items))

	newCronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caller.Name + "-cronjob",
			Namespace: caller.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&caller, apiv1.GroupVersion.WithKind(ourKind)),
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: caller.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            caller.Spec.Container.Image + caller.Name,
									Image:           caller.Spec.Container.Image,
									ImagePullPolicy: corev1.PullIfNotPresent,
									Command:         caller.Spec.Container.Command,
								},
							},
							RestartPolicy: corev1.RestartPolicy(caller.Spec.Container.RestartPolicy),
						},
					},
				},
			},
		},
	}

	if len(cronJobList.Items) == 0 {
		if err := r.Create(ctx, newCronJob); err != nil {
			log.Error(err, "Unable to create CronJob for Caller")
			return ctrl.Result{}, err
		} else {
			log.Info("CronJob Created")
		}
	} else {
		key := types.NamespacedName{
			Namespace: req.Namespace,
			Name:      newCronJob.Name,
		}
		cronJob := batchv1.CronJob{}
		if err := r.Get(ctx, key, &cronJob); err != nil {
			log.Info("No CronJob Found")
			return ctrl.Result{Requeue: true}, nil
		}
		jobs := batchv1.JobList{}
		opts := client.ListOptions{
			Namespace: cronJob.Namespace,
		}
		if err := r.List(ctx, &jobs, &opts); err != nil {
			log.Error(err, "Unable to fetch Jobs")
			return ctrl.Result{}, err
		}
		var success int32
		for _, job := range jobs.Items {
			if job.Status.Succeeded == 1 {
				success++
			}
		}
		caller.Status.CompletedJob = int32ptr(success)
		caller.Status.LastScheduleTime = cronJob.Status.LastScheduleTime
		caller.Status.LastSuccessfulTime = cronJob.Status.LastSuccessfulTime
		r.Status().Update(ctx, &caller)
	}

	log.Info("Exiting Reconcile...")

	return ctrl.Result{}, nil
}

func int32ptr(a int32) *int32 {
	return &a
}

var (
	cronJobOwnerKey = ".metadata.controller"
	ourKind         = "Caller"
	apiGVStr        = apiv1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *CallerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.CronJob{}, cronJobOwnerKey, func(o client.Object) []string {
		cronJob := o.(*batchv1.CronJob)
		owner := metav1.GetControllerOf(cronJob)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Caller{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
