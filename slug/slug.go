package slug

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	dcli "github.com/fsouza/go-dockerclient"
	"io"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"os"
	"os/exec"
)

var AWS_ACCESS_KEY_ID string
var AWS_SECRET_ACCESS_KEY string
var BUCKET_NAME string
var RUNNER_URL = "http://127.0.0.1:4243"

const S3_ENDPOINT = "https://s3.amazonaws.com"

// Build builds the slug with the received tar as content and upload it to S3
func Build(name string, tar io.Reader) (string, error) {
	slug := bytes.NewBuffer([]byte{})

	appBuildCache := fmt.Sprintf("/tmp/app-cache/%s", name)
	os.MkdirAll(appBuildCache, 0700)

	builder := exec.Command("docker", "run",
		"-i",
		"-a", "stdin",
		"-a", "stdout",
		"-a", "stderr",
		"-v", fmt.Sprintf("%s:/tmp/cache:rw", appBuildCache),
		"flynn/slugbuilder",
		"-")
	builder.Stderr = os.Stdout
	builder.Stdout = slug
	builder.Stdin = tar
	if err := builder.Run(); err != nil {
		return "", err
	}

	var auth = aws.Auth{AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY}
	var s3conn = s3.New(auth, aws.Region{S3Endpoint: S3_ENDPOINT})
	var bucket = s3conn.Bucket(BUCKET_NAME)
	var name_sha1 = sha1.Sum([]byte(name))
	var key = fmt.Sprintf("%x.tgz", name_sha1[:10])
	err := bucket.PutReader(key, slug, int64(slug.Len()), "application/tar", "private")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", S3_ENDPOINT, BUCKET_NAME, key), nil
}

// Run runs a slug process
func Run(name string, slugUrl string, process string, port string) (*dcli.Container, error) {
	docker, err := dcli.NewClient(RUNNER_URL)
	if err != nil {
		return nil, err
	}

	opts := dcli.CreateContainerOptions{
		"",
		&dcli.Config{
			Cmd:       []string{"start", process},
			Env:       []string{"SLUG_URL=" + slugUrl, "PORT=" + port},
			PortSpecs: []string{port + "/tcp"},
			Image:     "flynn/slugrunner",
			Tty:       true,
		},
	}

	container, err := docker.CreateContainer(opts)
	if err != nil {
		return nil, err
	}

	err = docker.StartContainer(container.ID, &dcli.HostConfig{})
	if err != nil {
		return nil, err
	}

	return docker.InspectContainer(container.ID)
}
