package ca_bundle

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const configOutputPath = "/opt/aws/amazon-cloudwatch-agent/bin/config.json"
const commonConfigOutputPath = "/opt/aws/amazon-cloudwatch-agent/etc/common-config.toml"
const configJSON = "/config.json"
const commonConfigTOML = "/common-config.toml"
const outputLog = "/opt/aws/amazon-cloudwatch-agent/logs/amazon-cloudwatch-agent.log"
const targetString = "x509: certificate signed by unknown authority"
//Let the agent run for 1 minutes. This will give agent enough time to call server
const agentRuntime = 60000


type input struct {
	findTarget bool
	dataInput string
}

//Must run this test with parallel 1 since this will fail if more than one test is running at the same time
func TestBundle(t *testing.T) {

	parameters := []input{
		{dataInput: "./resources/integration/ssl/with/combine/bundle", findTarget: false},
		{dataInput: "./resources/integration/ssl/with/original/bundle", findTarget: false},
		{dataInput: "./resources/integration/ssl/with/without/bundle", findTarget: false},
	}

	for _, parameter := range parameters {
		//before test run
		log.Printf("resource file location %s find target %t", parameter.dataInput, parameter.findTarget)
		clearLogFile()
		t.Run(fmt.Sprintf("resource file location %s find target %t", parameter.dataInput, parameter.findTarget), func(t *testing.T) {
			copyFile(parameter.dataInput +configJSON, configOutputPath)
			copyFile(parameter.dataInput +commonConfigTOML, commonConfigOutputPath)
			startTheAgent();
			time.Sleep(agentRuntime * time.Second);
			log.Printf("Agent has been running for : %d", agentRuntime);
			stopTheAgent();
			containsTarget := readTheOutputLog()
			if (parameter.findTarget && !containsTarget) || (!parameter.findTarget && containsTarget) {
				t.Errorf("Find target is %t contains target is %t", parameter.findTarget, containsTarget)
			}
		})
	}
}

func clearLogFile() {
	cmd := exec.Command("sudo rm " + outputLog + " && sudo cat > " + outputLog)

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Cleared log file : %s", outputLog);
}

func copyFile(pathIn string, pathOut string) {
	pathInAbs, err := filepath.Abs(pathIn)
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("sudo cp " + pathInAbs + " " + pathOut)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("File : %s copied to : %s", pathIn, pathOut);
}

func startTheAgent() {
	cmd := exec.Command("sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a " +
		"fetch-config -m ec2 -s -c file:" +
		configOutputPath)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Agent has started");
}

func stopTheAgent() {
	cmd := exec.Command("sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a stop")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Agent is stopped");
}

func readTheOutputLog() bool {
	logFile, err := ioutil.ReadFile(outputLog)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Log file %s", string(logFile))
	log.Printf("Finished reading log file")
	contains := strings.Contains(string(logFile), targetString)
	log.Printf("Log file contains target string %t", contains)
	return contains
}
