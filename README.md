# humalect-core

## Description

**humalect-core** is a CI/CD solution purposed to enhance the process of Docker-based application deployment within a Kubernetes environment. This open-source tool builds, pushes, and deploys Docker images into Kubernetes clusters by harnessing the capabilities of Kubernetes Custom Resource Definitions (CRDs) and a Kubernetes controller.

The primary goal of **humalect-core** is to simplify the complexities of integrating the CI/CD process within Kubernetes, thereby providing development teams with an easier and more efficient way to manage Docker applications from the stage of coding to deployment.

## Features

- **Seamless Integration:** Effortlessly integrates with Kubernetes using custom resource definitions and a purpose-built controller.
- **Automated Build Process:** Fetches your code from the repository, compiles, and wraps it into a Docker image automatically.
- **Docker Image Push:** Pushes Docker image to a Docker registry of your choice (either private or public).
- **Automated Deployment:** Deploys the Docker image to the Kubernetes cluster, offering a complete and autonomous CI/CD pipeline.
- **Scalability and Resilience:** Leverages the scalability and resilience of Kubernetes, ensuring your CI/CD pipeline is as robust as your production environment.
- **Intuitive Setup:** Enables easy setup and configuration using Kubernetes-native yaml files.
