// Receiver for flynn/gitreceive-next
// receiver-script SSH_USER REPO KEYNAME FINGERPRINT
package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"github.com/rochacon/cargo/nginx"
	"github.com/rochacon/cargo/slug"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const IMAGE_CACHE = "/tmp/app-cache"

func getRandomPort() (port int) {
	rand.Seed(time.Now().UnixNano())
	for port <= 1024 {
		port = rand.Intn(65534)
	}
	return
}

func main() {
	base_domain := flag.String("d", "localhost", "Base domain")
	bucket_name := flag.String("bucket", "", "AWS S3 bucket name")
	aws_key := flag.String("aws-key", "", "AWS access key")
	aws_secret := flag.String("aws-secret", "", "AWS secret key")
	flag.Parse()

	slug.AWS_ACCESS_KEY_ID = *aws_key
	slug.AWS_SECRET_ACCESS_KEY = *aws_secret
	slug.BUCKET_NAME = *bucket_name

	if len(flag.Args()) < 4 {
		flag.Usage()
		os.Exit(1)
	}

	app_name := flag.Arg(1)

	fmt.Println("-----> Building slug")

	slug_url, err := slug.Build(app_name, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	// TODO read Procfile and get processes
	// processes := []string{"web", "worker"}
	// fmt.Println("-----> Starting containers:", strings.Join(processes, ", "))

	// FIXME run all processes
	// for _, process := range processes {
	//   go docker.Run(-d -i -e SLUG_URL="$IMAGE" -e PORT=8000 -p $PORT:8000 flynn/slugrunner start web)
	// }

	var container_port = strconv.Itoa(getRandomPort())
	container, err := slug.Run(app_name, slug_url, "web", container_port)
	if err != nil {
		log.Fatal(err)
	}

	// TODO for web processes
	// TODO inspect container and retrieve IP (no need to expose container port)
	var app_name_sha1 = sha1.Sum([]byte(app_name))
	var app_name_for_url = fmt.Sprintf("%x", app_name_sha1)[:10]
	var hostname = fmt.Sprintf("%s.%s", app_name_for_url, *base_domain)

	// add container as a server to local NGINX
	err = nginx.AddServer(
		app_name,
		[]string{fmt.Sprintf("%s:%s", container.NetworkSettings.IPAddress, container_port)},
		hostname,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = nginx.Reload(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("-----> http://" + hostname)
}
