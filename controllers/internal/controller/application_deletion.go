package controller

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	k8sv1 "github.com/Humalect/humalect-core/api/v1"
)

func (r *ApplicationReconciler) handleDeletion(ctx context.Context, application *k8sv1.Application) (ctrl.Result, error) {
	// log := log.FromContext(ctx)
	// log.Info("Handling deletion", "namespace", commitId)

	// namespace := &corev1.Namespace{}
	// fmt.Print("namespace is ------------------ ", namespace)
	// err := r.Get(ctx, client.ObjectKey{Name: commitId}, namespace)
	// if err != nil {
	// 	if !errors.IsNotFound(err) {
	// 		log.Error(err, "Failed to get namespace", "namespace", commitId)
	// 		return ctrl.Result{}, err
	// 	}
	// } else {
	// 	// If the namespace exists, delete it
	// 	// if !containsString(namespace.ObjectMeta.Finalizers, finalizerName) {
	// 	err := r.Delete(ctx, namespace)
	// 	if err != nil {

	// 		log.Error(err, "Failed to delete namespace", "namespace", commitId)
	// 		return ctrl.Result{}, err
	// 	}

	// 	log.Info("Namespace marked for deletion", "namespace", commitId)
	// 	// Requeue the request to ensure the namespace is deleted before removing the finalizer
	// 	return ctrl.Result{}, nil
	// 	// }

	// }

	// Remove the finalizer and update the Application
	finalizerName := "finalizers.humalect.com/application"
	if containsString(application.ObjectMeta.Finalizers, finalizerName) {
		application.ObjectMeta.Finalizers = removeString(application.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, application); err != nil {
			fmt.Println("There is some error ", err)
			return ctrl.Result{}, err
		}
	}
	fmt.Println("There is no error ")

	return ctrl.Result{}, nil
}
