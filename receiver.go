// Receiver for flynn/gitreceive-next
// receiver-script SSH_USER REPO KEYNAME FINGERPRINT
package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	dcli "github.com/fsouza/go-dockerclient"
	"github.com/rochacon/cargo/nginx"
	"github.com/rochacon/cargo/slug"
	"log"
	"os"
)

const IMAGE_CACHE = "/tmp/app-cache"

func main() {
	aws_key := flag.String("aws-key", "", "AWS access key")
	aws_secret := flag.String("aws-secret", "", "AWS secret key")
	base_domain := flag.String("d", "localhost", "Base domain")
	bucket_name := flag.String("bucket", "", "AWS S3 bucket name")
	runner_endpoint := flag.String("runner", "http://127.0.0.1:4243", "Runner Docker API endpoint")
	s3_endpoint := flag.String("s3-endpoint", "https://s3.amazonaws.com", "AWS S3 API endpoint")
	flag.Parse()

	slug.AWS_ACCESS_KEY_ID = *aws_key
	slug.AWS_SECRET_ACCESS_KEY = *aws_secret
	slug.BUCKET_NAME = *bucket_name
	slug.RUNNER_ENDPOINT = *runner_endpoint
	slug.S3_ENDPOINT = *s3_endpoint

	if len(flag.Args()) < 4 {
		flag.Usage()
		os.Exit(1)
	}

	app_name := flag.Arg(1)

	// Build slug
	slug_url, err := slug.Build(app_name, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	// TODO read Procfile and get processes
	// processes := []string{"web"}
	// fmt.Println("-----> Starting containers:", strings.Join(processes, ", "))

	// FIXME run all processes
	// for _, process := range processes {
	//   go docker.Run(app_name, slug_url, process)
	// }

	container, err := slug.Run(app_name, slug_url, "web")
	if err != nil {
		log.Fatal(err)
	}

	// TODO for web processes
	// TODO inspect container and retrieve IP (no need to expose container port)
	var app_name_sha1 = sha1.Sum([]byte(app_name))
	var app_name_for_url = fmt.Sprintf("%x", app_name_sha1)[:10]
	var hostname = fmt.Sprintf("%s.%s", app_name_for_url, *base_domain)

	// add container as a server to local NGINX
	port := getPort(container.NetworkSettings.Ports)
	err = nginx.AddServer(
		app_name,
		[]string{fmt.Sprintf("%s:%s", container.NetworkSettings.IPAddress, port.Port())},
		hostname,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = nginx.Reload(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("-----> Application deployed")
	fmt.Println("       http://" + hostname)
}

// getPort naively retrieve the first port number of a
// map[dcli.Port][]dcli.PortBinding
func getPort(ports map[dcli.Port][]dcli.PortBinding) *dcli.Port {
	for port, _ := range ports {
		return &port
	}
	return nil
}
