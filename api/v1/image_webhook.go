package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var imagelog = logf.Log.WithName("image-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *Image) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ai-example-com-v1-image,mutating=false,failurePolicy=fail,sideEffects=None,groups=ai.example.com,resources=images,verbs=create;update,versions=v1,name=vimage.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Image{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Image) ValidateCreate() (admission.Warnings, error) {
	imagelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Image) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	imagelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Image) ValidateDelete() (admission.Warnings, error) {
	imagelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
