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

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sv1 "github.com/Humalect/humalect-core/api/v1"
	constants "github.com/Humalect/humalect-core/internal/controller/constants"
	helpers "github.com/Humalect/humalect-core/internal/controller/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8s.humalect.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.humalect.com,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.humalect.com,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile

func (r *ApplicationReconciler) setNamespaceOwnerReference(application *k8sv1.Application, namespace *corev1.Namespace) error {
	controllerRef := metav1.NewControllerRef(application, k8sv1.GroupVersion.WithKind("Application"))
	namespace.OwnerReferences = append(namespace.OwnerReferences, *controllerRef)
	return nil
}

const (
	applicationFinalizerName = "finalizers.humalect.com/application"
)

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	// Fetch the Application instance
	application := &k8sv1.Application{}
	err := r.Get(ctx, req.NamespacedName, application)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Application not found, ignoring reconcile")
			application.Spec.WebhookData = helpers.UpdateStatusData(application.Spec.WebhookData, constants.CreatedKubernetesResources, false)
			helpers.SendWebhook(application.Spec.WebhookEndpoint, application.Spec.WebhookData, false, constants.CreatedKubernetesResources)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Application")
		return ctrl.Result{}, err
	}

	fmt.Print("application.DeletionTimestamp: ", application.DeletionTimestamp)
	fmt.Print("application.DeletionTimestamp.IsZero(): ", application.DeletionTimestamp.IsZero())

	// Check if the Application is marked for deletion
	if !application.DeletionTimestamp.IsZero() && containsString(application.ObjectMeta.Finalizers, applicationFinalizerName) {
		fmt.Print("Deleting Application------------------ \n")
		return r.handleDeletion(ctx, application)
	} else {
		res, err := r.handleCreation(ctx, application, application.Spec.DeploymentYamlManifest, application.Spec.ServiceYamlManifest, application.Spec.IngressYamlManifest, application.Spec.Namespace)
		if err != nil {
			application.Spec.WebhookData = helpers.UpdateStatusData(application.Spec.WebhookData, constants.CreatedKubernetesResources, false)
			helpers.SendWebhook(application.Spec.WebhookEndpoint, application.Spec.WebhookData, false, constants.CreatedKubernetesResources)
		}
		application.Spec.WebhookData = helpers.UpdateStatusData(application.Spec.WebhookData, constants.CreatedKubernetesResources, true)
		application.Spec.WebhookData = helpers.UpdateStatusData(application.Spec.WebhookData, constants.DeploymentCompleted, true)
		helpers.SendWebhook(application.Spec.WebhookEndpoint, application.Spec.WebhookData, true, constants.DeploymentCompleted)
		return res, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1.Application{}).
		Complete(r)
}
