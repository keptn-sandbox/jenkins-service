# Jenkins Service for Keptn
This is a Sandbox Keptn Service integrating [Jenkins](https://jenkins.io/). This service allows you to have Keptn trigger any Jenkins pipeline (parameterized or not) for tasks such as deployment, testing or promotion of artifacts. This keptn service will also wait until the Jenkins pipeline is done executing and report back the status to Keptn to continue the Keptn pipeline orchestration!

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


## License

Please find more information in the [LICENSE](LICENSE) file.