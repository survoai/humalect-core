package services

import (
	"context"
	"fmt"
	"log"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WatchJobEvents(namespace, jobName string) bool {
	config := GetK8sConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}
	for {
		watcher, err := clientset.BatchV1().Jobs(namespace).Watch(context.TODO(), metav1.SingleObject(metav1.ObjectMeta{Name: jobName}))
		if err != nil {
			return false
		}

		ch := watcher.ResultChan()
		for event := range ch {
			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				panic("unexpected type")
			}

			succeeded := job.Status.Succeeded
			failed := job.Status.Failed

			if succeeded > 0 {
				fmt.Println("Job succeeded")
				return true
			} else if failed > 0 {
				fmt.Println("Job failed")
				return false
			} else {
				fmt.Println("Job status:", job.Status)
			}
		}
	}
}
