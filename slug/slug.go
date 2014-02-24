package slug

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	dcli "github.com/fsouza/go-dockerclient"
	"io"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var AWS_ACCESS_KEY_ID string
var AWS_SECRET_ACCESS_KEY string
var BUCKET_NAME string
var RUNNER_ENDPOINT string
var S3_ENDPOINT string

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
	return bucket.SignedURL(key, time.Now().Add(10*time.Second)), nil
}

// Run runs a slug process
func Run(slugUrl string, cmd ...string) (*dcli.Container, error) {
	docker, err := dcli.NewClient(RUNNER_ENDPOINT)
	if err != nil {
		return nil, err
	}

	port := strconv.Itoa(getRandomPort())

	opts := dcli.CreateContainerOptions{
		"",
		&dcli.Config{
			Cmd:       cmd,
			Env:       []string{"PORT=" + port, "SLUG_URL=" + slugUrl},
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

// RemoveOthers remove other containers running same slug
func RemoveOthers(container dcli.Container, slugUrl, process string) error {
	slugUrl = cleanURL(slugUrl)

	docker, err := dcli.NewClient(RUNNER_ENDPOINT)
	if err != nil {
		return err
	}

	containers, err := docker.ListContainers(dcli.ListContainersOptions{})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, c := range containers {
		if c.ID == container.ID {
			continue
		}

		wg.Add(1)
		go func(id, slugUrl string) {
			defer wg.Done()

			c, err := docker.InspectContainer(id)
			if err != nil {
				return
			}

			mustDie := false

			for _, env := range c.Config.Env {
				if !strings.HasPrefix(env, "SLUG_URL=") {
					continue
				}

				envSlug := strings.Split(env, "=")[1]
				if cleanURL(envSlug) == slugUrl {
					mustDie = true
				}
			}

			if mustDie {
				docker.StopContainer(c.ID, 15)
			}
		}(c.ID, slugUrl)
	}
	wg.Wait()

	return nil
}

// getRandomPort generates a random int to be used at the TCP port for
// container communication
func getRandomPort() (port int) {
	rand.Seed(time.Now().UnixNano())
	for port <= 1024 {
		port = rand.Intn(65534)
	}
	return
}

// cleanURL removes query string and fragments from an URL
func cleanURL(address string) string {
	u, _ := url.Parse(address)
	return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
}
