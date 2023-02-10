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

package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var customcrdlog = logf.Log.WithName("customcrd-resource")

func (r *CustomCrd) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-myapi-ishtiaq-com-v1-customcrd,mutating=true,failurePolicy=fail,sideEffects=None,groups=myapi.ishtiaq.com,resources=customcrds,verbs=create;update,versions=v1,name=mcustomcrd.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CustomCrd{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CustomCrd) Default() {
	customcrdlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.DeploymentName == "" {
		r.Spec.DeploymentName = r.Name + "-deploy"
	}
	if r.Spec.Replicas == nil {
		//r.Spec.Replicas = new(int)
		*r.Spec.Replicas = 1
	}
	if r.Spec.Service.ServiceName == "" {
		r.Spec.Service.ServiceName = r.Name + "-service"
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-myapi-ishtiaq-com-v1-customcrd,mutating=false,failurePolicy=fail,sideEffects=None,groups=myapi.ishtiaq.com,resources=customcrds,verbs=create;update,versions=v1,name=vcustomcrd.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CustomCrd{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomCrd) ValidateCreate() error {
	customcrdlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validationCustomCrd()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomCrd) ValidateUpdate(old runtime.Object) error {
	customcrdlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.validationCustomCrd()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CustomCrd) ValidateDelete() error {
	customcrdlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *CustomCrd) validationCustomCrd() error {
	var allErrs field.ErrorList
	if err := r.validateCustomCrdName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "myapi.ishtiaq.com", Kind: "CustomCrd"},
		r.Name, allErrs)
}

/*
Validating the length of a string field can be done declaratively by
the validation schema.
But the `ObjectMeta.Name` field is defined in a shared package under
the apimachinery repo, so we can't declaratively validate it using
the validation schema.
*/
func (r *CustomCrd) validateCustomCrdName() *field.Error {
	if len(r.ObjectMeta.Name) > validationutils.DNS1035LabelMaxLength-11 {
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 52 characters")
	}
	return nil
}
