# Jenkins Service for Keptn
This is a Sandbox Keptn Service integrating [Jenkins](https://jenkins.io/). This service allows you to have Keptn trigger any Jenkins job (parameterized or not) for tasks such as deployment, testing or promotion of artifacts. This keptn service will also wait until the Jenkins job is done executing and can send a Keptn event based on the outcome of that Jenkins job run in order to continue the Keptn pipeline orchestration!

**ATTENTION: THIS REPO IS CURRENTLY UNDER DEVELOPMENT. Expecting a first version soon!**

## Compatibility Matrix

| Keptn Version    | [Jenkins Service for Keptn] |
|:----------------:|:----------------------------------------:|
|       0.6.1      | grabner/jenkins-service:0.1.0 |

## Installation

The *jenkins-service* can be installed as a part of [Keptn's uniform](https://keptn.sh).

### Deploy in your Kubernetes cluster

To deploy the current version of the *jenkins-service* in your Keptn Kubernetes cluster, apply the [`deploy/service.yaml`](deploy/service.yaml) file:

```console
kubectl apply -f deploy/service.yaml
```

This should install the `jenkins-service` together with a Keptn `distributor` into the `keptn` namespace, which you can verify using

```console
kubectl -n keptn get deployment jenkins-service -o wide
kubectl -n keptn get pods -l run=jenkins-service
```

### Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by the `$VERSION` placeholder):

```console
kubectl -n keptn set image deployment/jenkins-service jenkins-service=your-username/jenkins-service:$VERSION --record
```

### Uninstall

To delete a deployed *jenkins-service*, use the file `deploy/*.yaml` files from this repository and delete the Kubernetes resources:

```console
kubectl delete -f deploy/service.yaml
```

## Usage

The **jenkins-service** can be used to execute a Jenkins Job upon receival of a specific Keptn event, e.g: Configuration.Change can trigger a Jenkins Job to deploy that change
The **jenkins-service** can also send a Keptn Event once the executed Jenkins Job is done, e.g: Deployment.Finished in case Jenkins Job deployed a new application

I think the most common use cases are the following as described in this table:
| Incoming Event    | Role of Jenkins Job  | Outgoing Event |
|:----------------:|:----------------------------------------:|:-----------------------------:|
| configuration.change | Deployment of that configuration change, e.g: deploy a new container, Java App, ... | deployment.finished event |
| deployment.finished | Execute Automated Functional Tests against that app, e.g: Selenium Tests  | tests.finished event |

In order to use the **jenkins-service** you need to create a jenkins.conf.yaml file and upload it to your Keptn Configuration Rep for your service in the jenkins subfolder. Here is an example:
```
keptn add-resource --project=PROJECTNAME --stage=STAGE --service=SERVICENAME --resource=jenkins/jenkins.conf.yaml
```

The **jenkins.conf.yaml** has 3 major sections:
1. Event Mapping: Defines which Keptn Event should execute which action and what to do with the response
2. Action Mapping: Defines which Jenkins Job should be executed with which parameters
3. Jenkins Servers: A List of Jenkins Servers that are accessible via the Jenkins REST API

There is a sample jenkins.conf.yaml in this repo. You will be able to see how to define your mappings.


## License

Please find more information in the [LICENSE](LICENSE) file.