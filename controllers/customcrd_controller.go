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
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	myapiv1 "ishtiaq.com/mycrd/api/v1"
)

// CustomCrdReconciler reconciles a CustomCrd object
type CustomCrdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=myapi.ishtiaq.com,resources=customcrds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=myapi.ishtiaq.com,resources=customcrds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=myapi.ishtiaq.com,resources=customcrds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomCrd object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CustomCrdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.WithValues("ReqName", req.Name, "ReqNameSpace", req.Namespace)

	// TODO(user): your logic here

	/*
		### 1: Load the CustomCrd by name

		We'll fetch the CustomCrd using our client.  All client methods take a
		context (to allow for cancellation) as their first argument, and the object
		in question as their last.  Get is a bit special, in that it takes a
		[`NamespacedName`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client?tab=doc#ObjectKey)
		as the middle argument (most don't have a middle argument, as we'll see
		below).

		Many client methods also take variadic options at the end.
	*/
	var customCrd myapiv1.CustomCrd
	if err := r.Get(ctx, req.NamespacedName, &customCrd); err != nil {
		log.Error(err, "unable to fetch customcrd")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/*
		### 2: List all active deployments, and update the status

		To fully update our status, we'll need to list all child deployments in this namespace that belong to this CustomCrd.
		Similarly to Get, we can use the List method to list the child deployments.  Notice that we use variadic options to
		set the namespace and field match (which is actually an index lookup that we set up below).
	*/
	var childDeploys appsv1.DeploymentList
	if err := r.List(ctx, &childDeploys, client.InNamespace(req.Namespace), client.MatchingFields{deployOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child Deployments")
		return ctrl.Result{}, err
	}

	log.WithValues("ChildDeploy", len(childDeploys.Items))

	/*
		### Note: What is this index about?

		The reconciler fetches all deployments owned by the CustomCrd for the status. As our number of customcrds increases,
		looking these up can become quite slow as we have to filter through all of them. For a more efficient lookup,
		these deployments will be indexed locally on the controller's name. A deployOwnerKey field is added to the
		cached deployment objects. This key references the owning controller and functions as the index. Later in this
		document we will configure the manager to actually index this field.
	*/

	// if no ChildDeployment found, or found deployments are not owned by CustomCrd then create one on the cluster
	trimAppName := func(s string) string {
		arr := strings.Split(s, "/")
		if len(arr) == 1 {
			return arr[0]
		}
		return arr[1]
	}

	setName := func(s, suf string) string {
		fmt.Println("Name:", s, customCrd.Name)
		if s == "" {
			s = customCrd.Name + suf
		}
		return s
	}

	newDeployment := func(customCrd *myapiv1.CustomCrd) *appsv1.Deployment {
		log.Info("New Deployment is called")

		labels := map[string]string{
			"app":        trimAppName(customCrd.Spec.Container.Image),
			"controller": customCrd.Name,
		}

		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				//Name:      setName(customCrd.Spec.DeploymentName, "-deploy"),
				Namespace: customCrd.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(customCrd, myapiv1.GroupVersion.WithKind(ourKind)),
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: customCrd.Spec.Replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "my-container",
								Image: customCrd.Spec.Container.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: customCrd.Spec.Container.Port,
									},
								},
							},
						},
					},
				},
			},
		}
	}

	if len(childDeploys.Items) == 0 || !findDeploysOwnedByCustomCrd(&childDeploys) {
		deploy := newDeployment(&customCrd)
		if err := r.Create(ctx, deploy); err != nil {
			log.Error(err, "unable to create deployment for CustomCrd", "Deploy", deploy)
			return ctrl.Result{}, err
		} else {
			log.WithValues("created childDeploys", len(childDeploys.Items))
		}
	}

	// same for service
	var childServices corev1.ServiceList
	if err := r.List(ctx, &childServices, client.InNamespace(req.Namespace), client.MatchingFields{svcOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child services")
		return ctrl.Result{}, err
	}

	log.WithValues("ChildDeploy", len(childServices.Items))

	// if no childService found or the default 'kubernetes' service found, create one on the cluster
	getServiceType := func(s string) corev1.ServiceType {
		if s == "NodePort" {
			return corev1.ServiceTypeNodePort
		} else {
			return corev1.ServiceTypeClusterIP
		}
	}

	newService := func(customCrd *myapiv1.CustomCrd) *corev1.Service {
		log.Info("New Service is called")
		labels := map[string]string{
			"app":        trimAppName(customCrd.Spec.Container.Image),
			"controller": customCrd.Name,
		}

		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      setName(customCrd.Spec.Service.ServiceName, "-service"),
				Namespace: customCrd.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(customCrd, myapiv1.GroupVersion.WithKind(ourKind)),
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: labels,
				Type:     getServiceType(customCrd.Spec.Service.ServiceType),
				Ports: []corev1.ServicePort{
					{
						NodePort:   customCrd.Spec.Service.ServiceNodePort,
						Port:       customCrd.Spec.Container.Port,
						TargetPort: intstr.FromInt(int(customCrd.Spec.Container.Port)),
					},
				},
			},
		}
	}

	if len(childServices.Items) == 0 || !findServicesOwnedByCustomCrd(&childServices) {
		svcObj := newService(&customCrd)
		if err := r.Create(ctx, svcObj); err != nil {
			log.Error(err, "unable to create service for CustomCrd", "service", svcObj)
			return ctrl.Result{}, err
		} else {
			log.WithValues("created childServices", len(childServices.Items))
		}
	}

	// your logic here
	fmt.Println("Reconcilier function has been called")

	return ctrl.Result{}, nil
}

// findDeploysOwnedByCustomCrd
func findDeploysOwnedByCustomCrd(deploys *appsv1.DeploymentList) bool {
	for i := 0; i < len(deploys.Items); i++ {
		ownRefs := deploys.Items[0].GetOwnerReferences()
		for j := 0; j < len(ownRefs); j++ {
			if (ownRefs[j].APIVersion == apiGVStr) && (ownRefs[j].Kind == ourKind) {
				return true
			}
		}
	}
	return false
}

// findServicesOwnedByCustomCrd
func findServicesOwnedByCustomCrd(services *corev1.ServiceList) bool {
	for i := 0; i < len(services.Items); i++ {
		ownRefs := services.Items[0].GetOwnerReferences()
		for j := 0; j < len(ownRefs); j++ {
			if (ownRefs[j].APIVersion == apiGVStr) && (ownRefs[j].Kind == ourKind) {
				return true
			}
		}
	}
	return false
}

/*
### Setup

Finally, we'll update our setup.  In order to allow our reconciler to quickly
look up deployments by their owner, we'll need an index.  We declare an index key that
we can later use with the client as a pseudo-field name, and then describe how to
extract the indexed value from the Job object.  The indexer will automatically take
care of namespaces for us, so we just have to extract the owner name if the deployment has
a CustomCrd owner.

Additionally, we'll inform the manager that this controller owns some deployments, so that it
will automatically call Reconcile on the underlying CustomCrd when a Deployment changes, deleted,
etc.
*/
var (
	deployOwnerKey = ".metadata.controller"
	svcOwnerKey    = ".metadata.controller"
	apiGVStr       = myapiv1.GroupVersion.String()
	ourKind        = "CustomCrd"
)

// SetupWithManager sets up the controller with the Manager.
func (r *CustomCrdReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, deployOwnerKey, func(rawObj client.Object) []string {
		// grab the deployment object, extract the owner...
		deploy := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deploy)
		if owner == nil {
			return nil
		}
		// make sure its a CustomCrd...
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, svcOwnerKey, func(rawObj client.Object) []string {
		// grab the service object, extract the owner...
		svc := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(svc)
		if owner == nil {
			return nil
		}
		// make sure its a CustomCrd...
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	fmt.Println("SetupWithManager Successful.")
	return ctrl.NewControllerManagedBy(mgr).
		For(&myapiv1.CustomCrd{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
