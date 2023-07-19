package constants

const (
	RegistryIdDockerhub               = "dockerhub"
	CloudIdAzure                      = "azure"
	CloudIdAWS                        = "aws"
	CloudIdCivo                       = "civo"
	DeploymentJobCreated              = "DEPLOYMENT_JOB_CREATED"
	WebhookTypeDeploymentStatusUpdate = "TYPE_DEPLOYMENT_STATUS_UPDATE"
	DeploymentFailed                  = "DEPLOYMENT_FAILED"
	CreatedKubernetesResources        = "CREATED_KUBERNETES_RESOURCES"
	DeploymentCompleted               = "DEPLOYMENT_COMPLETED"
)

type SecretConfig struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType,omitempty"`
}

// AzureResourcesName := os.Getenv("DATABASE_URL")
