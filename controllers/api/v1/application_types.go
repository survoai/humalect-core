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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationSpec defines the desired state of Application
type DeploymentYamlManifestType struct {
	Metadata metav1.ObjectMeta     `json:"metadata"`
	Spec     appsv1.DeploymentSpec `json:"spec"`
}

type ServiceYamlManifestType struct {
	Metadata metav1.ObjectMeta  `json:"metadata"`
	Spec     corev1.ServiceSpec `json:"spec"`
}
type IngressYamlManifestType struct {
	Metadata metav1.ObjectMeta        `json:"metadata"`
	Spec     networkingv1.IngressSpec `json:"spec"`
}

type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AwsSecretCredentials     AwsSecretCredentials       `json:"awsSecretCredentials,omitempty"`
	AzureVaultCredentials    AzureVaultCredentials      `json:"azureVaultCredentials,omitempty"`
	CloudRegion              string                     `json:"cloudRegion,omitempty"`
	SecretsProvider          string                     `json:"secretsProvider,omitempty"`
	CloudProvider            string                     `json:"cloudProvider,omitempty"`
	K8sResourcesIdentifier   string                     `json:"k8sResourcesIdentifier,omitempty"`
	DeploymentYamlManifest   DeploymentYamlManifestType `json:"deploymentYamlManifest"`
	ServiceYamlManifest      ServiceYamlManifestType    `json:"serviceYamlManifest"`
	IngressYamlManifest      IngressYamlManifestType    `json:"ingressYamlManifest"`
	BuildSecretsConfig       []SecretConfig             `json:"buildSecretsConfig,omitempty"`
	ApplicationSecretsConfig []SecretConfig             `json:"applicationSecretsConfig,omitempty"`
	ManagedBy                string                     `json:"managedBy,omitempty"`
	Namespace                string                     `json:"namespace"`
	WebhookEndpoint          string                     `json:"webhookEndpoint"`
	WebhookData              string                     `json:"webhookData"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
