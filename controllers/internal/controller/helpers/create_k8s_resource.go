package helpers

import (
	"context"
	"reflect"

	k8sv1 "github.com/Humalect/humalect-core/api/v1"
	constants "github.com/Humalect/humalect-core/internal/controller/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}
type Object interface {
	metav1.Object
	runtime.Object
}

func createEmptyObject(obj Object) Object {
	objType := reflect.TypeOf(obj)
	emptyObj := reflect.New(objType.Elem()).Interface().(Object)
	return emptyObj
}

func CreateK8sResource(ctx context.Context, application *k8sv1.Application, namespace string, r *ApplicationReconciler, objs ...Object) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	for _, obj := range objs {
		log.Info("Create Resource - ", reflect.TypeOf(obj).String(), obj.GetObjectKind())
		emptyObj := createEmptyObject(obj)

		log.Info("Handling creation", reflect.TypeOf(obj).String(), obj.GetName())
		obj.SetNamespace(namespace)

		if err := r.Get(ctx, client.ObjectKey{Name: obj.GetName(), Namespace: namespace}, emptyObj); err != nil {
			if errors.IsNotFound(err) {

				controllerRef := metav1.NewControllerRef(application, k8sv1.GroupVersion.WithKind("Application"))
				obj.SetOwnerReferences(append(obj.GetOwnerReferences(), *controllerRef))

				if err := r.Create(ctx, obj); err != nil {
					log.Error(err, "Failed to create", reflect.TypeOf(obj).String(), obj.GetName())
					SendWebhook(application.Spec.WebhookEndpoint, application.Spec.WebhookData, false, constants.CreatedKubernetesResources)
					return ctrl.Result{}, err
				}
				log.Info("Created Resource", reflect.TypeOf(obj).String(), obj.GetName())
			} else {
				log.Error(err, "Failed to get", reflect.TypeOf(emptyObj).String(), obj.GetName())
				SendWebhook(application.Spec.WebhookEndpoint, application.Spec.WebhookData, false, constants.CreatedKubernetesResources)
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Resource already exists, skipping creation", reflect.TypeOf(obj).String(), obj.GetName())
			controllerRef := metav1.NewControllerRef(application, k8sv1.GroupVersion.WithKind("Application"))
			obj.SetOwnerReferences(append(obj.GetOwnerReferences(), *controllerRef))
			if err := r.Update(ctx, obj); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}
