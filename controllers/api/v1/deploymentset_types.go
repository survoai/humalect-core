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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeploymentSetSpec defines the desired state of DeploymentSet

type EcrCredentials struct {
	AccessKey   string `json:"accessKey,omitempty"`
	SecretKey   string `json:"secretKey,omitempty"`
	RegistryUrl string `json:"registryUrl,omitempty"`
}

type DockerHubCredentials struct {
	Username   string `json:"username,omitempty"`
	SecretName string `json:"secretName,omitempty"`
}

type AcrCredentials struct {
	ManagementScopeToken string `json:"managementScopeToken,omitempty"`
	RegistryName         string `json:"registryName,omitempty"`
	SubscriptionId       string `json:"subscriptionId,omitempty"`
	ResourceGroupName    string `json:"resourceGroupName,omitempty"`
}

type AwsSecretCredentials struct {
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
	Region    string `json:"region,omitempty"`
}

type AzureVaultCredentials struct {
	Token string `json:"token,omitempty"`
	Name  string `json:"name,omitempty"`
}

type DeploymentSetSpec struct {
	ArtifactsRegistryProvider string                     `json:"artifactsRegistryProvider,omitempty"`
	SecretsProvider           string                     `json:"secretsProvider,omitempty"`
	EcrCredentials            EcrCredentials             `json:"ecrCredentials,omitempty"`
	DockerHubCredentials      DockerHubCredentials       `json:"dockerHubCredentials,omitempty"`
	AcrCredentials            AcrCredentials             `json:"acrCredentials,omitempty"`
	AwsSecretCredentials      AwsSecretCredentials       `json:"awsSecretCredentials,omitempty"`
	AzureVaultCredentials     AzureVaultCredentials      `json:"azureVaultCredentials,omitempty"`
	CommitId                  string                     `json:"commitId,omitempty"`
	SourceCodeToken           string                     `json:"sourceCodeToken,omitempty"`
	CloudRegion               string                     `json:"cloudRegion,omitempty"`
	SourceCodeProvider        string                     `json:"sourceCodeProvider,omitempty"`
	ArtifactsRepositoryName   string                     `json:"artifactsRepositoryName,omitempty"`
	CloudProvider             string                     `json:"cloudProvider,omitempty"`
	SourceCodeOrgName         string                     `json:"sourceCodeOrgName,omitempty"`
	SourceCodeRepositoryName  string                     `json:"sourceCodeRepositoryName,omitempty"`
	K8sResourcesIdentifier    string                     `json:"k8sResourcesIdentifier,omitempty"`
	DeploymentYamlManifest    DeploymentYamlManifestType `json:"deploymentYamlManifest"`
	ServiceYamlManifest       ServiceYamlManifestType    `json:"serviceYamlManifest"`
	IngressYamlManifest       IngressYamlManifestType    `json:"ingressYamlManifest"`
	DockerManifest            []string                   `json:"dockerManifest,omitempty"`
	SecretManagerName         string                     `json:"secretManagerName,omitempty"`
	ManagedBy                 string                     `json:"managedBy,omitempty"`
	AzureVaultToken           string                     `json:"azureVaultToken,omitempty"`
	AzureVaultName            string                     `json:"azureVaultName,omitempty"`
	AzureResourceGroupName    string                     `json:"azureResourceGroupName,omitempty"`
	AzureSubscriptionId       string                     `json:"azureSubscriptionId,omitempty"`
	AzureAcrRegistryName      string                     `json:"azureAcrRegistryName,omitempty"`
	AzureManagementScopeToken string                     `json:"azureManagementScopeToken,omitempty"`
	AwsEcrUserName            string                     `json:"awsEcrUserName,omitempty"`
	UseDockerFromCodeFlag     bool                       `json:"useDockerFromCodeFlag,omitempty"`
	AwsEcrRegistryUrl         string                     `json:"awsEcrRegistryUrl,omitempty"`
	JobName                   string                     `json:"jobName"`
	K8sAppName                string                     `json:"k8sAppName"`
	Namespace                 string                     `json:"namespace"`
	DeploymentId              string                     `json:"deploymentId"`
	WebhookData               string                     `json:"webhookData"`
	WebhookEndpoint           string                     `json:"webhookEndpoint"`
}

// DeploymentSetStatus defines the observed state of DeploymentSet
type DeploymentSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DeploymentSet is the Schema for the deploymentsets API
type DeploymentSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSetSpec   `json:"spec,omitempty"`
	Status DeploymentSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeploymentSetList contains a list of DeploymentSet
type DeploymentSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeploymentSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeploymentSet{}, &DeploymentSetList{})
}
