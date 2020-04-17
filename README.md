# jenkins-pipeline

Execute groovy pipeline code in remote jenkins server and monitor the output in client console(I prefer [Atom](https://atom.io/) editor).

![jenkins-pipeline](https://github.com/gitrootid/jenkins-pipeline-go/blob/master/help/atom-preview.png?raw=true)

Install Example:

  *  download tar package file from release
  *  tar xf jenkins-pipeline-go-linux-bin.tgz
  *  mv jenkins-pipeline-go-linux-bin/jenkins-pipeline-go /usr/local/bin
  *  jenkins-pipeline-go -h

Create pipeline job in jenkins

![create-pipeline-job](https://github.com/gitrootid/jenkins-pipeline-go/blob/master/help/pipeline-job.png?raw=true)

Command

    jenkins-pipeline-go -file <path to groovy file> -url <http://jenkins.host:port> -job <path-to-pipeline-job> -username <jenkins-username> -api-token <api-token of the username> -trigger-token <any string,keep default value is fine>

Example command line

    jenkins-pipeline-go -file ~/pipeline_demo.groovy -job /job/test-pipeline -template ~/config.xml.template -url http://localhost:8080 -username admin -api-token 11111460a1115de06456a83ed16822c8eb  

## Atom editor configuration

To configure this command in [Atom Editor](https://atom.io/), make sure you have [build](https://atom.io/packages/build) package installed in your atom editor and add `.atom-build.yml` file in project folder with below yml code. For more information on how to build, please check this [link](https://atom.io/packages/build)

Save this as `.atom-build.yml`, and build(<kbd>ctrl</kbd>+<kbd>Alt</kbd>+<kbd>b</kbd>) your groovy file.

    cmd: "jenkins-pipeline-go"
    args:
      - "-file {FILE_ACTIVE}"
      - "-url http://localhost:8080"
      - "-job /job/test-pipeline"
      - "-username admin"
      - "-api-token 11111460a1115de06456a83ed16822c8eb"
    sh: true

## compile

Before compile,please make sure [Go](https://github.com/golang/go) environment installed
for linux plateform:

    make build4linux

for windows plateform:

    make build4windows

## Note

Groovy pipeline code should be saved as with `.groovy` extension

## Idea source

    https://github.com/jainath/jenkins-pipeline
