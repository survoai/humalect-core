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

## Repository Structure

The **humalect-core** repository is organized into two primary directories:

- **agent:** This directory contains the source code for a Dockerized Golang application responsible for the build, push, and deployment processes. It essentially forms the engine of our CI/CD pipeline.

- **controllers:** This directory hosts the code for a Kubernetes Custom Resource Definition (CRD) controller. The CRD controller is designed to detect and respond to changes in the custom resources related to build and deployment processes.

The repository structure is as follows:
```
humalect-core
│
├── agent
│ ├── [files and folders related to Dockerized Golang application]
│
└── controllers
| ├── [files and folders related to the CRD controller]
```
