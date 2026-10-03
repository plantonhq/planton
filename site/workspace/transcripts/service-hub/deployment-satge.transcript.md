Serviceub will essentially kick off the deployment of a microser Kubernetes by creating a micros service on Kubernetes catalog kind with by extracting the configuration from _kustomize folder within the check uh co-loccated alongside the service code which I'll cover in a few minutes in the next video but the point is as part of service of build and deployment pipelines once the build is completed and the image is pushed to the corresponding configured docker registry or container registry. Um the we tecton has a step which runs build uh customized build and uh the customized folders consists of uh the yl manifests of the catalog kinds uh of planton cloud and if if the developer wants to deploy the created container image from the source code onto micros service as a micros service on Kubernetes then the developer needs to add a customized folder with base and overlays um as uh let's see an example For example, this web console which is an XJS app is is configured to deploy to Kubernetes. So the way it is done is there should be a folder called_ customize and a base folder and overlays folder. Each overlay folder folder represents an environment to which it is deployed and there will always be two files in each of these folders. One is called customization.l and another one is called service.l. The name of the file other than customization.l doesn't matter it can be anything but uh for consistency we simply put it as service.l and the contents of this service.l can represent uh a few varieties of catalog kinds on plant and clothes infraub. Um it can be a micros service kubernetes, it can be ECS service, it can be digital ocean app platform service, it can be even lambda function very soon. Um so the idea is u and the overlayless service should also have the same exact kind. Um and when we run customized build the we leverage the open source customize tool and uh uh the the overlays each of the overlay directory will produce one full catalog kind manifest of microser Kubernetes and uh uh that generated uh uh environment overlay environment to the generated uh catalog kind YAML manifest map is stored as a config map on Kubernetes and that is all part of Tecton's uh customized build step uh and And in um Plantum Cloud's pipeline orchestration logic once the build stage is completed uh because we know where the image is pushed or the commit tag of the container image uh is pushed. We know that because it is uh predictable because it is always the commit sha. Uh the only thing that we know need to know is uh we need to understand uh what to which platform the developer wants. to deploy the service. And we determine that by looking up the config map created as part of the customized build step and uh as part of PIP pipeline orchestration uh logic on planton cloud in within the temporal workflow. Again none of the internal details the like the temporal and stuff needs to be exposed to the end users. This is only for uh the AI um tools that help us gen the documentation so that they better understand the context. So the temporal orchestration logic will look up um uh there is an activity that looks up the config map um from the kubernetes uh that runs tecton and it looks up the overlays uh environment to service uh yl map and based on the contents of the service.yamel map for each environment the pipeline orchestration will create uh one deployment uh task for each uh overlay environment. So the deployment stage will come contain arbitrary number of uh deployment tasks. Each task represents one overlay directory which uh result produces one uh infra component Lego EML block and uh plan the orchestration logic will simply apply that manifest which results in either create or updating the cloud uh resource catalog kind and as mentioned previously uh any creation or updation of an infra component catalog kind produces an infra job. So the deployment stage of uh the pipeline orchestration is nothing but uh a collection of infra jobs mapped to each uh created for each uh overlay uh directory. So we didn't reinvent uh anything related to deployment the service of nicely ties into um uh the service of nicely ties into infraub and leverages infra jobs uh thereby creating a nice cohesive experience. So when when the developers are looking at uh the uh logs of deployment stage they're essentially looking at uh the infra job executed to take that catalog kind manifest and uh run pulley or terraform plan and supply which produces a micros service on Kubernetes. So that explains how the services go from build to deployments.


❯ tree _kustomize
_kustomize
├── base
│   ├── kustomization.yaml
│   └── service.yaml
└── overlays
    ├── local
    │   ├── kustomization.yaml
    │   └── service.yaml
    └── prod
        ├── kustomization.yaml
        └── service.yaml

5 directories, 6 files

## ~/scm/github.com/plantonhq/planton/backend/services/copilot-agent/_kustomize/base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - service.yaml

## ~/scm/github.com/plantonhq/planton/backend/services/copilot-agent/_kustomize/base/service.yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: copilot-agent
  org: planton-cloud
spec:
  container:
    app:
      env:
        variables:
          PORT: "8080"
          LOG_LEVEL: info
          POSTGRES_DB_NAME: $variables-group/postgres-copilot-agent/db.copilot-agent.name
          POSTGRES_USERNAME: $variables-group/postgres-copilot-agent/username
          REDIS_CLUSTER_PORT: $variables-group/redis/port
          ELASTICSEARCH_USERNAME: $variables-group/elasticsearch/username
          ELASTICSEARCH_VECTOR_CLOUD_INFRA_API_RESOURCE_YAML: cloud_infra_api_resource_yaml
          ELASTICSEARCH_VECTOR_CREDENTIAL_API_RESOURCE_YAML: credential_api_resource_yaml
          ELASTICSEARCH_VECTOR_PLANTON_DOCUMENTATION: planton_documentation
          ELASTICSEARCH_LOAD_VECTOR: "false"
          ELASTICSEARCH_PORT: $variables-group/elasticsearch/port
          LANGCHAIN_ENDPOINT: $variables-group/langchain/api-endpoint
          PLANTON_CLOUD_APIS_REPO_OWNER: $variables-group/planton-cloud/apis.repo-owner
          PLANTON_CLOUD_APIS_REPO: $variables-group/planton-cloud/apis.repo
          PLANTON_CLOUD_APIS_GITHUB_TAG: $variables-group/planton-cloud/apis.github-tag
          PLANTON_CLOUD_APIS_APIRESOURCEKIND_ENUM_PROTO_FILE_PATH: $variables-group/planton-cloud/apis.api-resource-kind-enum-proto-file-path
          PROJECT_PLANTON_REPO_OWNER: $variables-group/openmcf/repo-owner
          PROJECT_PLANTON_REPO: $variables-group/openmcf/repo
          PROJECT_PLANTON_VERSION: $variables-group/openmcf/version
          PROJECT_PLANTON_SHARED_OPTIONS_PROTO_FILE_PATH: $variables-group/openmcf/shared-options-proto-file-path
        secrets:
          GITHUB_TOKEN: $secrets-group/github/token
          BUF_TOKEN: $secrets-group/buf/token
      resources:
        limits:
          cpu: 500m
          memory: 500Mi
        requests:
          cpu: 50m
          memory: 100Mi
      ports:
        - name: grpc-api
          appProtocol: grpc
          networkProtocol: TCP
          servicePort: 80
          containerPort: 8080
          isIngressPort: true
  ingress:
    enabled: true
    dnsDomain: planton.live
  availability:
    minReplicas: 1
    horizontalPodAutoscaling:
      isEnabled: false
  version: main

## ~/scm/github.com/plantonhq/planton/backend/services/copilot-agent/_kustomize/overlays/local/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../prod

patches:
  - path: service.yaml

## ~/scm/github.com/plantonhq/planton/backend/services/copilot-agent/_kustomize/overlays/local/service.yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: copilot-agent
spec:
  container:
    app:
      env:
        variables:
          PORT: "8003"
          GRPC_SERVICE_HUB: $variables-group/grpc-endpoints/prod.service-hub.external-endpoint
          GRPC_INFRA_HUB: $variables-group/grpc-endpoints/prod.infra-hub.external-endpoint
          GRPC_CONNECT: $variables-group/grpc-endpoints/prod.connect.external-endpoint
          GRPC_IAM: $variables-group/grpc-endpoints/prod.iam.external-endpoint
          GRPC_KUBE_OPS: $variables-group/grpc-endpoints/prod.kube-ops.external-endpoint
          GRPC_RESOURCE_MANAGER: $variables-group/grpc-endpoints/prod.resource-manager.external-endpoint
          GRPC_SEARCH: $variables-group/grpc-endpoints/prod.search.external-endpoint
          GRPC_COPILOT_CHAT: localhost:8092
          POSTGRES_HOST: $variables-group/postgres-copilot-agent/prod.external-endpoint
          ELASTICSEARCH_URL: $variables-group/elasticsearch/prod.external-endpoint
          LANGSMITH_PROJECT: "planton-cloud-local"
          ELASTICSEARCH_LOAD_VECTOR: false
          COPILOT_AGENT_RUNTIME_API_TOKEN: pak_fVNVV-Gu1Cvqc
          COPILOT_AGENT_RUNTIME: "test"
          COPILOT_AGENT_RUNTIME_ORG: "planton-cloud"

## ~/scm/github.com/plantonhq/planton/backend/services/copilot-agent/_kustomize/overlays/prod/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: service.yaml

## ~/scm/github.com/plantonhq/planton/backend/services/copilot-agent/_kustomize/overlays/prod/service.yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: copilot-agent
  env: app-prod
spec:
  container:
    app:
      resources:
        limits:
          cpu: "2"
          memory: 2Gi
        requests:
          cpu: 60m
          memory: 100Mi
      env:
        variables:
          ELASTICSEARCH_URL: $variables-group/elasticsearch/prod.external-endpoint
          GRPC_CONNECT: $variables-group/grpc-endpoints/prod.connect.kube-endpoint
          GRPC_IAM: $variables-group/grpc-endpoints/prod.iam.kube-endpoint
          GRPC_KUBE_OPS: $variables-group/grpc-endpoints/prod.kube-ops.kube-endpoint
          GRPC_INFRA_HUB: $variables-group/grpc-endpoints/prod.infra-hub.kube-endpoint
          GRPC_RESOURCE_MANAGER: $variables-group/grpc-endpoints/prod.resource-manager.kube-endpoint
          GRPC_SEARCH: $variables-group/grpc-endpoints/prod.search.kube-endpoint
          GRPC_COPILOT_CHAT: $variables-group/grpc-endpoints/prod.copilot-chat.kube-endpoint
          GRPC_SERVICE_HUB: $variables-group/grpc-endpoints/prod.service-hub.kube-endpoint
          LANGSMITH_PROJECT: $variables-group/langchain/project
          LANGSMITH_TRACING: "true"
          OPEN_AI_EMBEDDINGS_MODEl_NAME: $variables-group/openai/embeddings-model
          OPEN_AI_LLM_MODEl_NAME: $variables-group/openai/reasoning-model
          POSTGRES_HOST: $variables-group/postgres-copilot-agent/prod.external-endpoint
          WEB_CONSOLE_INFRA_JOB_ID_PATH: $variables-group/web-console/prod.iac-infra-job-id-path
        secrets:
          OPENAI_API_KEY: $secrets-group/openai/api-key
          POSTGRES_PASSWORD: $secrets-group/postgres-copilot-agent-main/password
          ELASTICSEARCH_PASSWORD: $secrets-group/elasticsearch-prod/password
          LANGSMITH_API_KEY: $secrets-group/langchain/api-key

----

local overlay folder is excluded

 ~/scm/github.com/plantonhq/planton/backend/services/copilot-agent  main                                                                                                                    11:39:18
❯ kustomize build _kustomize/overlays/prod
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  env: app-prod
  name: copilot-agent
  org: planton-cloud
spec:
  availability:
    horizontalPodAutoscaling:
      isEnabled: false
    minReplicas: 1
  container:
    app:
      env:
        secrets:
          BUF_TOKEN: $secrets-group/buf/token
          ELASTICSEARCH_PASSWORD: $secrets-group/elasticsearch-prod/password
          GITHUB_TOKEN: $secrets-group/github/token
          LANGSMITH_API_KEY: $secrets-group/langchain/api-key
          OPENAI_API_KEY: $secrets-group/openai/api-key
          POSTGRES_PASSWORD: $secrets-group/postgres-copilot-agent-main/password
        variables:
          ELASTICSEARCH_LOAD_VECTOR: "false"
          ELASTICSEARCH_PORT: $variables-group/elasticsearch/port
          ELASTICSEARCH_URL: $variables-group/elasticsearch/prod.external-endpoint
          ELASTICSEARCH_USERNAME: $variables-group/elasticsearch/username
          ELASTICSEARCH_VECTOR_CLOUD_INFRA_API_RESOURCE_YAML: cloud_infra_api_resource_yaml
          ELASTICSEARCH_VECTOR_CREDENTIAL_API_RESOURCE_YAML: credential_api_resource_yaml
          ELASTICSEARCH_VECTOR_PLANTON_DOCUMENTATION: planton_documentation
          GRPC_CONNECT: $variables-group/grpc-endpoints/prod.connect.kube-endpoint
          GRPC_COPILOT_CHAT: $variables-group/grpc-endpoints/prod.copilot-chat.kube-endpoint
          GRPC_IAM: $variables-group/grpc-endpoints/prod.iam.kube-endpoint
          GRPC_INFRA_HUB: $variables-group/grpc-endpoints/prod.infra-hub.kube-endpoint
          GRPC_KUBE_OPS: $variables-group/grpc-endpoints/prod.kube-ops.kube-endpoint
          GRPC_RESOURCE_MANAGER: $variables-group/grpc-endpoints/prod.resource-manager.kube-endpoint
          GRPC_SEARCH: $variables-group/grpc-endpoints/prod.search.kube-endpoint
          GRPC_SERVICE_HUB: $variables-group/grpc-endpoints/prod.service-hub.kube-endpoint
          LANGCHAIN_ENDPOINT: $variables-group/langchain/api-endpoint
          LANGSMITH_PROJECT: $variables-group/langchain/project
          LANGSMITH_TRACING: "true"
          LOG_LEVEL: info
          OPEN_AI_EMBEDDINGS_MODEl_NAME: $variables-group/openai/embeddings-model
          OPEN_AI_LLM_MODEl_NAME: $variables-group/openai/reasoning-model
          PLANTON_CLOUD_APIS_APIRESOURCEKIND_ENUM_PROTO_FILE_PATH: $variables-group/planton-cloud/apis.api-resource-kind-enum-proto-file-path
          PLANTON_CLOUD_APIS_GITHUB_TAG: $variables-group/planton-cloud/apis.github-tag
          PLANTON_CLOUD_APIS_REPO: $variables-group/planton-cloud/apis.repo
          PLANTON_CLOUD_APIS_REPO_OWNER: $variables-group/planton-cloud/apis.repo-owner
          PORT: "8080"
          POSTGRES_DB_NAME: $variables-group/postgres-copilot-agent/db.copilot-agent.name
          POSTGRES_HOST: $variables-group/postgres-copilot-agent/prod.external-endpoint
          POSTGRES_USERNAME: $variables-group/postgres-copilot-agent/username
          PROJECT_PLANTON_REPO: $variables-group/openmcf/repo
          PROJECT_PLANTON_REPO_OWNER: $variables-group/openmcf/repo-owner
          PROJECT_PLANTON_SHARED_OPTIONS_PROTO_FILE_PATH: $variables-group/openmcf/shared-options-proto-file-path
          PROJECT_PLANTON_VERSION: $variables-group/openmcf/version
          REDIS_CLUSTER_PORT: $variables-group/redis/port
          WEB_CONSOLE_INFRA_JOB_ID_PATH: $variables-group/web-console/prod.iac-infra-job-id-path
      ports:
      - appProtocol: grpc
        containerPort: 8080
        isIngressPort: true
        name: grpc-api
        networkProtocol: TCP
        servicePort: 80
      resources:
        limits:
          cpu: "2"
          memory: 2Gi
        requests:
          cpu: 60m
          memory: 100Mi
  ingress:
    dnsDomain: planton.live
    enabled: true
  version: main
  