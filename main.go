package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

type DefaultCrumbIssuer struct {
	XMLName           xml.Name `xml:"defaultCrumbIssuer"`
	Text              string   `xml:",chardata"`
	Class             string   `xml:"_class,attr"`
	Crumb             string   `xml:"crumb"`
	CrumbRequestField string   `xml:"crumbRequestField"`
}

type HttpClient struct {
	Client *resty.Client
	Url    string
	Job    string
}

type JobApi struct {
	NextBuildNumber int `json:"nextBuildNumber"`
}

type ResultApi struct {
	Building bool `json:"building"`
}

var groovyFile, jenkinsUrl, job, userName, apiToken, triggerToken, template string

func init() {
	flag.StringVar(&groovyFile, "file", "", "path to groovy file,example:./pipeline_demo.groovy")
	flag.StringVar(&jenkinsUrl, "url", "", "jenkins http url,example:http://127.0.0.1:8080")
	flag.StringVar(&job, "job", "", "job uri,example:/job/pipeline_demo")
	flag.StringVar(&userName, "username", "", "user name")
	flag.StringVar(&apiToken, "api-token", "", "api token of the username")
	flag.StringVar(&triggerToken, "trigger-token", "DEFAULT_TRIGGER_TOKEN", "trigger token if need to replace default value in template")
	flag.StringVar(&template, "template", "", "path to config.xml.template,example:./config.xml.template")
	flag.Parse()
	var msg string
	switch {
	case jenkinsUrl == "":
		msg = "  -url\n"
		fallthrough
	case job == "":
		msg = msg + "  -job\n"
		fallthrough
	case userName == "":
		msg = msg + "  -username\n"
		fallthrough
	case apiToken == "":
		msg = msg + "  -api-token\n"
		fallthrough
	case groovyFile == "":
		msg = msg + "  -file\n"
		fallthrough
	case template == "":
		msg = msg + "  -template\n"
	}
	if msg != "" {
		flag.PrintDefaults()
		HandlerErr(errors.New(fmt.Sprintf("insufficiency parameter(s):\n%s", msg)))
	}
}

func main() {
	httpClient := resty.New().SetDisableWarn(true)
	httpClient.SetBasicAuth(userName, apiToken)
	var client = &HttpClient{
		Client: httpClient,
		Url:    jenkinsUrl,
		Job:    job,
	}
	code, filed := client.GetCrumbCode()
	log.Println(fmt.Sprintf("got %s:%s", filed, code))
	client.UpdateConfig(code, filed, groovyFile, triggerToken)
	currentBuild := client.ExecuteBuild(triggerToken)
	time.Sleep(2 * time.Second)
	client.GetBuildStatus(currentBuild)

}

func (h *HttpClient) GetCrumbCode() (code, filed string) {
	resp, err := h.Client.R().Get(h.Url + "/crumbIssuer/api/xml")
	HandlerErr(err)
	if HandlerStatusCode(resp.StatusCode()) {
		crumb := &DefaultCrumbIssuer{}
		err = xml.Unmarshal(resp.Body(), crumb)
		HandlerErr(err)
		code = crumb.Crumb
		filed = crumb.CrumbRequestField
		return
	}
	HandlerErr(errors.New(resp.String()))
	return
}

func (h *HttpClient) UpdateConfig(crumbField, crumb, scriptFile, triggerToken string) {
	if triggerToken == "" {
		triggerToken = "defaultTriggerToken"
	}
	xmlFileBytes, err := ioutil.ReadFile(template)
	HandlerErr(err)
	scriptFileBytes, err := ioutil.ReadFile(scriptFile)
	HandlerErr(err)
	fileContent := string(xmlFileBytes)
	fileContent = strings.Replace(fileContent, "####SCRIPT####", string(scriptFileBytes), 1)
	fileContent = strings.Replace(fileContent, "####DEFAULT_TRIGGER_TOKEN####", triggerToken, 1)
	h.Client.SetHeader(crumbField, crumb)
	_, err = h.Client.R().SetBody([]byte(fileContent)).Post(h.Url + h.Job + "/config.xml")
	HandlerErr(err)
	log.Println("Updated job config.xml file successfully")
}

func (h *HttpClient) ExecuteBuild(triggerToken string) (currentBuild string) {
	resp, err := h.Client.R().Post(h.Url + h.Job + "/api/json")
	HandlerErr(err)
	jobApi := &JobApi{}
	err = json.Unmarshal(resp.Body(), jobApi)
	HandlerErr(err)
	currentBuild = strconv.Itoa(jobApi.NextBuildNumber)
	resp, err = h.Client.R().Post(h.Url + h.Job + "/build?token=" + triggerToken)
	HandlerErr(err)
	if HandlerStatusCode(resp.StatusCode()) {
		log.Printf("Executed build #%s successfully\n", currentBuild)
		return
	}
	HandlerErr(errors.New(resp.String()))
	return
}

func (h *HttpClient) GetBuildStatus(currentBuildId string) {
	resultApi := &ResultApi{}
	waitingCount := 10
	var textSize int64
	log.Println("waiting for starting")
	for {
		resp, err := h.Client.R().Post(h.Url + h.Job + "/" + currentBuildId + "/api/json")
		HandlerErr(err)
		if HandlerStatusCode(resp.StatusCode()) {
			err = json.Unmarshal(resp.Body(), resultApi)
			HandlerErr(err)
			response, err := h.Client.R().Post(h.Url + h.Job + "/" + currentBuildId + "/consoleText")
			resp, err = response, err
			HandlerErr(err)
			if textSize == 0 {
				fmt.Print("\n")
				log.Print("output consoleText:")
			}
			fmt.Print(string(resp.Body()[textSize:]))
			textSize = resp.Size()
			if (resp.Size() == textSize) && (resultApi.Building == false) {
				break
			}
			time.Sleep(1 * time.Second)
			continue
		}
		fmt.Println(".")
		if waitingCount == 0 {
			HandlerErr(errors.New("something wrong,time out for waiting"))
		}
		waitingCount -= 1
		time.Sleep(2 * time.Second)
	}
}

func HandlerStatusCode(statusCode int) bool {
	if (statusCode == 200) || (statusCode == 201) {
		return true
	}
	return false
}

func HandlerErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
