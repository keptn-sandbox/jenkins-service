package main

import (
	"log"

	"github.com/bndr/gojenkins"
)

/**
 * Defines the Jenkins Configuration
 */
type JenkinsConfig struct {
	url   string
	user  string
	token string
	job   string
}

//
// Leveraging Jenkins Go Client Library from https://github.com/bndr/gojenkins
//

/**
 * Validates that the Jenkins API can be reached and that the Jenkins server has the minimum required version
 */
func validateJenkins(jenkinsConfig JenkinsConfig) (bool, error) {
	jenkins := gojenkins.CreateJenkins(nil, jenkinsConfig.url, jenkinsConfig.user, jenkinsConfig.token)
	// Provide CA certificate if server is using self-signed certificate
	// caCert, _ := ioutil.ReadFile("/tmp/ca.crt")
	// jenkins.Requester.CACert = caCert
	_, err := jenkins.Init()

	if err != nil {
		return false, err
	}

	log.Printf("Successful connection to Jenkins Version %s", jenkins.Version)

	// now lets validate that the requested job exists
	_, err = jenkins.GetJob(jenkinsConfig.job)
	if err != nil {
		log.Printf("Job %s does not exist!", jenkinsConfig.job)
		return false, err
	}

	log.Printf("Job %s does exist!", jenkinsConfig.job)

	return true, nil
}
