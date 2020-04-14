# Jenkins Service for Keptn
This is a Sandbox [Keptn](https://www.keptn.sh) Service integrating [Jenkins](https://jenkins.io/). This service allows you to have Keptn trigger any Jenkins job (parameterized or not) for tasks such as deployment, testing or promotion of artifacts. This keptn service will also wait until the Jenkins job is done executing and can send a Keptn event based on the outcome of that Jenkins job run in order to continue the Keptn pipeline orchestration!

**ATTENTION: THIS REPO IS CURRENTLY UNDER DEVELOPMENT. Expecting a first version soon!**

## Compatibility Matrix

| Keptn Version    | Jenkins Service for Keptn                | Description
|:----------------:|:----------------------------------------:| --------- |
|       0.6.1      | grabner/jenkins-service:0.1.0            | Initial Version |

## Installation

The *jenkins-service* can be installed as a part of [Keptn's uniform](https://keptn.sh).

### Deploy in your Kubernetes cluster

To deploy the current version of the *jenkins-service* in your Keptn Kubernetes cluster, apply the [`deploy/service.yaml`](deploy/service.yaml) file.
Before you do edit service.yaml and add any custom enviornment variable to it, e.g: your JENKINS_URL, JENKINS_USER, JENKINS_PASSWORD

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
1. **Event Mapping**: Defines which Keptn Event should execute which action and what to do with the response
2. **Action Mapping**: Defines which Jenkins Job should be executed with which parameters
3. **Jenkins Servers**: A List of Jenkins Servers that are accessible via the Jenkins REST API

There is a sample [jenkins.conf.yaml](jenkins/jenkins.conf.yaml) in this repo.
```
spec_version: '0.1.0'
jenkinsservers:
  - name: MyJenkinsServer
    url: $ENV.JENKINS_URL
    user: $ENV.JENKINS_USER
    token: $ENV.JENKINS_PASSWORD
actions:
  - name: MyKeptnJenkinsJob
    server: MyJenkinsServer
    jenkinsjob: TestPipelineWithParams
    buildparameters:
      Message: My message from Kept Project $PROJECT-$STAGE-$SERVICE for event $EVENT
      WaitTime: 2
      Result: SUCCESS
  - name: MyKeptnJenkinsJobWithFailure
    server: MyJenkinsServer
    jenkinsjob: TestPipelineWithParams
    buildparameters:
      Message: My message from Kept Project $PROJECT-$STAGE-$SERVICE for event $EVENT
      WaitTime: 2
      Result: FAILURE
events:
  - event: configuration.change
    action: MyKeptnJenkinsJob
    timeout: 60
    onsuccess:
      event: deployment.finished
      deploymentURIPublic: http://yourdeployedapp.yourdomain
      result: pass
    onfailure:
      event: deployment.finished
      result: failed
```

Let me quickly explain what it does (from bottom to top):
1. If the **jenkins-service** receives a configuration.change event
2. It will execute the action that is defined in **MyKeptnJenkinsJob** against the defined Jenkin Server
3. **onsuccess** it will send a deployment.finished event with the passed deploymentURIPublic
4. **onfailure** it will send a deployment.finished event without a URI and a result of failed

**Events: Mapping a Keptn Event to an Action**
You can define any number of event mappings but only the first that matches the incoming Keptn Cloud Event (configuration.change, deployment.finished, tests.finished, ...) will trigger as the **jenkins-service** currently only supports executing a single Job for an incoming event.
There are two options for **jenkins-service** to execute the job
1. If you specifyy onsuccess and/or onfailure **jenkins-service** will wait for the Jenkins Job to finish execution unless it runs into the specified timeout. In this case it will send a new Keptn event based on the event and parameters specified
2. If you DO NOT specify onsuccess or onfailure then **jenkins-service** only triggers the Jenkins Job and doesnt wait for any completion.

**Actions: Defining the details for a Jenkins Job**
As you can see in the example you can have multiple actions and each action has a logical name, e.g: MyKeptnJenkinsJob or MyJenkinsJenkinsJobWithFailure.
An action always maps to a Jenkins Job that will be executed on a Jenkins Server and can have parameters. In the example above both actions are calling the same Jenkins Job called **TestPipelineWithParams*, but - each Job is called with slightly different parameters.

**Jenkinsservers: Defining the credentials for the Jenkins API**
The ** jenkins-service** uses the Jenkins REST API by leveraging the [Jenkins Go Client Library](https://github.com/bndr/gojenkins). What you need to specify is your Jenkins URL, username and password. In order to not have this information in clear text I suggest you do it like shown in the example: Use the ENV.placeholder capability and pass this confidential information as part of your deployment definition in service.yaml

**Placeholders**
I've implemented the same placeholders as in the [Generic Executor Service](https://github.com/keptn-sandbox/generic-executor-service). Here is the full overview:
```
// Event Context
$CONTEXT,$EVENT,$SOURCE,$TIMESTRING,$TIMEUTCSTRING,$TIMEUTCMS

// Project Context
$PROJECT,$STAGE,$SERVICE,$DEPLOYMENT,$TESTSTRATEGY
    
// Deployment Finished specific
$DEPLOYMENTURILOCAL,$DEPLOYMENTURIPUBLIC

// Labels will be made available with a $LABEL. prefix, e.g.:
$LABEL.gitcommit,$LABEL.anotherlabel,$LABEL.xxxx

// Environment variables you pass to the generic-executor-service container in the service.yaml will be available with $ENV. prefix
$ENV.YOURCUSTOMENV,$ENV.KEPTN_API_TOKEN,$ENV.KEPTN_ENDPOINT,...
```


## Jenkins Sample Job

To build this **jenkins-service** I used a very simply Jenkins Job Pipeline to test this out. It is a parameterized job with 3 parameters
1. Message: Just a simple message that is echoed. This allows me to test the Keptn Placeholders, e.g: $SERVICE ...
2. SleepTime: The value is used in a sleep statement - this allows me to control the length of the job run
3. Result: The value will be used to set the build status. Therefore allows me to simulate a SUCCESS or FAILURE status
```
node {
   properties([
        parameters([
         string(defaultValue: 'This is the default message', description: '', name: 'Message', trim: false), 
         string(defaultValue: '5', description: '', name: 'SleepTime', trim: false), 
         string(defaultValue: 'SUCCESS', description: 'Use SUCCESS or FAILURE', name: 'Result', trim: false)
        ])
    ])
   stage('Preparation') {
           echo "${params.Message}"
   }
   stage('Doing') {
           sleep "${params.SleepTime}"
   }
   stage('Results') {
       currentBuild.result = "${params.Result}"
   }
}
```

## License

Please find more information in the [LICENSE](LICENSE) file.