package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/ghodss/yaml"
	"github.com/keptn/go-utils/pkg/configuration-service/utils"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"

	"github.com/bndr/gojenkins"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

//
// Leveraging Jenkins Go Client Library from https://github.com/bndr/gojenkins
//

func makeJson(data interface{}) string {
	str, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(json.RawMessage(str))
}

//
// Loads jenkins.conf for the current service
//
func getJenkinsConfiguration(keptnEvent baseKeptnEvent, logger *keptnutils.Logger) (*JenkinsConfigFile, error) {
	// if we run in a runlocal mode we are just getting the file from the local disk
	var fileContent []byte
	var err error
	if runlocal {
		fileContent, err = ioutil.ReadFile(JenkinsConfigFilenameLOCAL)
		if err != nil {
			logMessage := fmt.Sprintf("No %s file found LOCALLY for service %s in stage %s in project %s", JenkinsConfigFilenameLOCAL, keptnEvent.service, keptnEvent.stage, keptnEvent.project)
			logger.Info(logMessage)
			return nil, errors.New(logMessage)
		}
	} else {
		resourceHandler := utils.NewResourceHandler("configuration-service:8080")
		keptnResourceContent, err := resourceHandler.GetServiceResource(keptnEvent.service, keptnEvent.stage, keptnEvent.project, JenkinsConfigFilename)
		if err != nil {
			logMessage := fmt.Sprintf("No %s file found for service %s in stage %s in project %s", JenkinsConfigFilename, keptnEvent.service, keptnEvent.stage, keptnEvent.project)
			logger.Info(logMessage)
			return nil, errors.New(logMessage)
		}
		fileContent = []byte(keptnResourceContent.ResourceContent)
	}

	// replace the placeholders
	fileAsString := string(fileContent)
	fileAsStringReplaced := replaceKeptnPlaceholders(fileAsString, keptnEvent)
	fileContent = []byte(fileAsStringReplaced)

	// Print file content
	// fmt.Println(fileAsStringReplaced)

	// unmarshal the file
	var jenkinsConfFile *JenkinsConfigFile
	jenkinsConfFile, err = parseJenkinsConfigFile(fileContent)

	if err != nil {
		logMessage := fmt.Sprintf("Couldn't parse %s file found for service %s in stage %s in project %s. Error: %s", JenkinsConfigFilename, keptnEvent.service, keptnEvent.stage, keptnEvent.project, err.Error())
		logger.Error(logMessage)
		return nil, errors.New(logMessage)
	}

	return jenkinsConfFile, nil
}

func parseJenkinsConfigFile(input []byte) (*JenkinsConfigFile, error) {
	jenkinsConfFile := &JenkinsConfigFile{}
	err := yaml.Unmarshal([]byte(input), &jenkinsConfFile)

	if err != nil {
		return nil, err
	}

	return jenkinsConfFile, nil
}

/**
 * Iterates through the JenkinsConfigFile and returns the jenkins server configuration by name
 */
func getJenkinsServerConfig(jenkinsConfigFile *JenkinsConfigFile, servername string) (*JenkinsServerConfig, error) {
	// get the entry for the passed name
	if jenkinsConfigFile != nil && jenkinsConfigFile.JenkinsServers != nil {
		for _, serverConfig := range jenkinsConfigFile.JenkinsServers {
			if serverConfig.Name == servername {
				return serverConfig, nil
			}
		}
	}

	return nil, errors.New("No Jenkins Server Configuration found for " + servername)
}

/**
 * Iterates through the JenkinsConfigFile and returns the action configuration by name
 */
func getActionConfig(jenkinsConfigFile *JenkinsConfigFile, actionname string) (*ActionConfig, error) {
	// get the entry for the passed name
	if jenkinsConfigFile != nil && jenkinsConfigFile.Actions != nil {
		for _, actionConfig := range jenkinsConfigFile.Actions {
			if actionConfig.Name == actionname {
				return actionConfig, nil
			}
		}
	}

	return nil, errors.New("No Jenkins Job Configuration found for " + actionname)
}

/**
 * Iterates through the events and returns the first matching the event name
 */
func getEventMappingConfig(jenkinsConfigFile *JenkinsConfigFile, eventname string) (*EventMappingConfig, error) {
	// get the entry for the passed name
	if jenkinsConfigFile != nil && jenkinsConfigFile.Events != nil {
		for _, eventMapConfig := range jenkinsConfigFile.Events {
			if eventMapConfig.Event == eventname {
				return eventMapConfig, nil
			}
		}
	}

	return nil, errors.New("No Event Map Configuration found for " + eventname)
}

/**
 * Retrieves the Jenkins Job
 */
func getJenkinsJob(actionConfig *ActionConfig, jenkinsServerConfig *JenkinsServerConfig) (*gojenkins.Job, error) {
	jenkins := gojenkins.CreateJenkins(nil, jenkinsServerConfig.Url, jenkinsServerConfig.User, jenkinsServerConfig.Token)
	// Provide CA certificate if server is using self-signed certificate
	// caCert, _ := ioutil.ReadFile("/tmp/ca.crt")
	// jenkins.Requester.CACert = caCert
	_, err := jenkins.Init()

	if err != nil {
		return nil, err
	}

	log.Printf("Successful connection to Jenkins Version %s", jenkins.Version)

	// now lets validate that the requested job exists
	var job *gojenkins.Job
	job, err = jenkins.GetJob(actionConfig.JenkinsJob)
	if err != nil {
		log.Printf("Job %s does not exist!", actionConfig.JenkinsJob)
		return nil, err
	}

	log.Printf("Job %s does exist!", actionConfig.JenkinsJob)

	return job, nil
}

/**
 * takes the map of parameters from the action config and converts it into a name/value map
 */
/*func getJenkinsParameterMap(actionConfig *ActionConfig) map[string]string {
	// m := make(map[string]string)
}*/

func executeJenkinsJob(actionConfig *ActionConfig, jenkinsServerConfig *JenkinsServerConfig) (*gojenkins.Job, int64, error) {
	// now lets validate that the requested job exists
	var buildNumber int64
	buildNumber = 0
	job, err := getJenkinsJob(actionConfig, jenkinsServerConfig)
	if err != nil {
		log.Printf("Job %s does not exist!", actionConfig.JenkinsJob)
		return nil, buildNumber, err
	}

	log.Printf("Executing %s on Jenkins Server %s with params: %s", actionConfig.JenkinsJob, jenkinsServerConfig.Name, actionConfig.BuildParameters)

	// had to build in some retries as it turns out the Jenkins API often returns 403 errors for API attempts!
	retryCount := DEFAULT_INVOKE_RETRY_COUNT
	for retryCount > 0 {
		buildNumber, err = job.InvokeSimple(actionConfig.BuildParameters)
		if err == nil {
			log.Printf("Job %s running with Build Number %d", actionConfig.JenkinsJob, buildNumber)
			return job, buildNumber, err
		} else {
			log.Printf("Job %s invocation failed. Retry Count Left(%d): %s", actionConfig.JenkinsJob, retryCount, err.Error())
			time.Sleep(DEFAULT_INVOKE_RETRY_WAIT * time.Second)
		}
	}

	return nil, buildNumber, err
}

/**
 * Executes the job and waits until completition if specified in the configuration
 */
func executeJenkinsJobAndWaitForCompletion(eventMapConfig *EventMappingConfig, actionConfig *ActionConfig, jenkinsServerConfig *JenkinsServerConfig) (bool, error) {

	// before we execute the job we save current time in the eventMap
	eventMapConfig.startedAt = time.Now()
	eventMapConfig.finishedAt = time.Now()

	job, buildNumber, err := executeJenkinsJob(actionConfig, jenkinsServerConfig)
	if err != nil {
		return false, nil
	}

	// lets see if we have to wait for the completion. If not we just return true!
	if (len(eventMapConfig.OnSuccess) == 0) && (len(eventMapConfig.OnFailure) == 0) {
		return true, nil
	}

	// TODO - make sure we are polling the currently started job that might still be in queue -> https://github.com/bndr/gojenkins/issues/161
	log.Printf("Waiting until %d is finished", buildNumber)

	// now - lets check correct timeout values
	timeout := eventMapConfig.Timeout
	if timeout <= 0 {
		timeout = DEFAULT_TIMEOUT
	}
	if timeout > MAX_TIMEOUT {
		timeout = MAX_TIMEOUT
	}

	// query the job state until its done or until we run into our max wait
	var lastBuild *gojenkins.Build = nil
	timeleft := timeout
	for timeleft > 0 {
		start := time.Now()

		// first we sleep before we try to fetch the job state
		time.Sleep(DEFAULT_WAIT_RETRY * time.Second)

		// first we need to poll the job to get the latest data
		job.Poll()

		// then we query last build object
		lastBuild, err = job.GetLastBuild()
		if err != nil {
			log.Printf("Couldnt retrieve last build from job %s. Error: %s", actionConfig.JenkinsJob, err.Error())
			return false, err
		}

		// now lets check the status
		if !lastBuild.IsRunning() {
			buildResult := lastBuild.GetResult()
			log.Printf("Build %d finished with status: %s", lastBuild.GetBuildNumber(), buildResult)
			eventMapConfig.finishedAt = time.Now()
			return buildResult == "SUCCESS", nil
		}

		log.Printf("Build %d still running. Checking again in %ds", lastBuild.GetBuildNumber(), DEFAULT_WAIT_RETRY)

		// adjust our timeout with the time this iteration took
		t := time.Now()
		elapsed := t.Sub(start)
		timeleft = timeleft - int(elapsed.Seconds())
	}

	logMessage := fmt.Sprintf("Job %s did not finish within %d seconds", actionConfig.JenkinsJob, timeout)
	log.Printf(logMessage)
	eventMapConfig.finishedAt = time.Now()
	return false, errors.New(logMessage)
}

/**
 * Based on the actionSuccess sends the onSuccess or onFailure Event definition
 */
func sendKeptnEventForEventConfig(incomingBaseEvent *baseKeptnEvent, incomingEvent *cloudevents.Event, eventMappingConfig *EventMappingConfig, actionSuccess bool, logger *keptnutils.Logger) (bool, error) {

	// first lets get the correct OnXX data set
	var eventData map[string]string
	if actionSuccess {
		eventData = eventMappingConfig.OnSuccess
	} else {
		eventData = eventMappingConfig.OnFailure
	}

	// double check if we have to do anything at all, e.g: if there is no OnXX data set we are done
	if len(eventData) == 0 {
		return true, nil
	}

	// we have to have at least the event name
	eventName, eventNameExists := eventData["event"]
	if eventNameExists == false {
		return false, errors.New("No event was specified to send back to Keptn")
	} else {
		log.Println(fmt.Sprintf("Processing OnXX configuration with eventData: %s", eventData))
	}

	// now lets validate that we have at least the event name, e.g: deployment.finished
	switch eventName {
	case "deployment.finished":
		deploymentURILocal, _ := eventData["deploymentURILocal"]
		deploymentURIPublic, _ := eventData["deploymentURIPublic"]

		log.Println(deploymentURILocal, deploymentURIPublic)

		sendDeploymentFinishedEvent(incomingBaseEvent.context, incomingEvent, incomingBaseEvent.project, incomingBaseEvent.service, incomingBaseEvent.stage, incomingBaseEvent.testStrategy, incomingBaseEvent.deployment, "", "",
			deploymentURILocal,
			deploymentURIPublic,
			incomingBaseEvent.labels, logger)
	case "tests.finished":
		result, _ := eventData["result"]
		sendTestsFinishedEvent(incomingBaseEvent.context, incomingEvent, incomingBaseEvent.project, incomingBaseEvent.service, incomingBaseEvent.stage, incomingBaseEvent.testStrategy, incomingBaseEvent.deployment,
			eventMappingConfig.startedAt,
			eventMappingConfig.finishedAt,
			result,
			incomingBaseEvent.labels, logger)
	default:
		return false, errors.New("Event not supported: " + eventName)
	}

	return true, nil
}
