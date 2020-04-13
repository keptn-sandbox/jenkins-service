package main

import "time"

const JenkinsConfigFilename = "jenkins/jenkins.conf.yaml"
const JenkinsConfigFilenameLOCAL = "jenkins/_jenkins.conf.yaml"
const DEFAULT_TIMEOUT = 600   // 600seconds == 10 minutes
const MAX_TIMEOUT = 3600      // 3600seconds == 1h
const DEFAULT_WAIT_RETRY = 10 // 10 seconds

const DEFAULT_INVOKE_RETRY_COUNT = 3 // Try Job Invoke 3 times
const DEFAULT_INVOKE_RETRY_WAIT = 2  // 2 Seconds wait between retries

type JenkinsConfigFile struct {
	SpecVersion    string                 `json:"spec_version" yaml:"spec_version"`
	JenkinsServers []*JenkinsServerConfig `json:"jenkinsservers" yaml:"jenkinsservers"`
	Actions        []*ActionConfig        `json:"actions" yaml:"actions"`
	Events         []*EventMappingConfig  `json:"events" yaml:"events"`
}

/**
 * Defines the Jenkins Configuration
 */
type JenkinsServerConfig struct {
	Name  string `json:"name" yaml:"name"`
	Url   string `json:"url" yaml:"url"`
	User  string `json:"user" yaml:"user"`
	Token string `json:"token" yaml:"token"`
}

type ActionConfig struct {
	Name            string            `json:"name" yaml:"name"`
	Server          string            `json:"server" yaml:"server"`
	JenkinsJob      string            `json:"jenkinsjob" yaml:"jenkinsjob"`
	BuildParameters map[string]string `json:"buildparameters" yaml:"buildparameters"`
}

type EventMappingConfig struct {
	Event      string            `json:"event" yaml:"event"`
	Action     string            `json:"action" yaml:"action"`
	Timeout    int               `json:"timeout" yaml:"timoue"`
	OnSuccess  map[string]string `json:"onsuccess" yaml:"onsuccess"`
	OnFailure  map[string]string `json:"onfailure" yaml:"onfailure"`
	startedAt  time.Time
	finishedAt time.Time
}
