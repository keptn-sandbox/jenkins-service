package main

import (
	"fmt"

	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
)

/**
* Here are all the handler functions for the individual events
  -> "sh.keptn.event.configuration.change"
  -> "sh.keptn.events.deployment-finished"
  -> "sh.keptn.events.tests-finished"
  -> "sh.keptn.event.start-evaluation"
  -> "sh.keptn.events.evaluation-done"
  -> "sh.keptn.event.problem.open"
  -> "sh.keptn.events.problem"
*/

/**
* Will lookup the events.event==eventname entry - if there is a match will execute that Jenkins Job and send the configured event back to Keptn
	eventname := "configuration.change"
*/
func handleEvent(eventname string, event cloudevents.Event, keptnEvent baseKeptnEvent, logger *keptnutils.Logger) error {
	jenkinsConfigFile, err := getJenkinsConfiguration(keptnEvent, logger)
	if err != nil {
		return err
	}

	logger.Info("Loaded Jenkins Config File")
	eventMapConfig, err := getEventMappingConfig(jenkinsConfigFile, eventname)
	if err != nil {
		logger.Info(fmt.Sprintf("No event mapping found for %s. Therefore executing no Jenkins Job", eventname))
		return nil
	}

	logger.Info(fmt.Sprintf("Event Mapping %s points to action %s", eventname, eventMapConfig.Action))

	var actionConfig *ActionConfig
	actionConfig, err = getActionConfig(jenkinsConfigFile, eventMapConfig.Action)
	if err != nil {
		return fmt.Errorf("No Action Config found for: %s", eventMapConfig.Action)
	}

	var serverConfig *JenkinsServerConfig
	serverConfig, err = getJenkinsServerConfig(jenkinsConfigFile, actionConfig.Server)
	if err != nil {
		return fmt.Errorf("No Server Config found for: %s", actionConfig.Server)
	}

	var success bool
	var keptnResult *KeptnResultArtifact
	success, keptnResult, err = executeJenkinsJobAndWaitForCompletion(eventMapConfig, actionConfig, serverConfig)
	if err != nil {
		return fmt.Errorf("Error executing: %s", err.Error())
	}

	// Now lets send back the Keptn Event if one is configured for OnSuccess or OnFailure
	success, err = sendKeptnEventForEventConfig(
		&keptnEvent, &event,
		eventMapConfig,
		success, keptnResult,
		logger)

	return err
}

//
// Handles ConfigurationChangeEventType = "sh.keptn.event.configuration.change"
// TODO: add in your handler code
//
func handleConfigurationChangeEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.ConfigurationChangeEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Configuration Changed Event: %s", event.Context.GetID()))

	return handleEvent("configuration.change", event, keptnEvent, logger)
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func handleDeploymentFinishedEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.DeploymentFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Deployment Finished Event: %s", event.Context.GetID()))

	return handleEvent("deployment.finished", event, keptnEvent, logger)
}

//
// Handles TestsFinishedEventType = "sh.keptn.events.tests-finished"
// TODO: add in your handler code
//
func handleTestsFinishedEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.TestsFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Tests Finished Event: %s", event.Context.GetID()))

	return handleEvent("tests.finished", event, keptnEvent, logger)
}

//
// Handles EvaluationDoneEventType = "sh.keptn.events.start-evaluation"
// TODO: add in your handler code
//
func handleStartEvaluationEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.StartEvaluationEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Start Evaluation Event: %s", event.Context.GetID()))

	return handleEvent("start.evaluation", event, keptnEvent, logger)
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.evaluation-done"
// TODO: add in your handler code
//
func handleEvaluationDoneEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.EvaluationDoneEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Evaluation Done Event: %s", event.Context.GetID()))

	return nil
}

//
// Handles ProblemOpenEventType = "sh.keptn.event.problem.open"
// Handles ProblemEventType = "sh.keptn.events.problem"
// TODO: add in your handler code
//
func handleProblemEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.ProblemEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Problem Event: %s", event.Context.GetID()))

	return nil
}
