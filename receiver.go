// Receiver for flynn/gitreceive-next
// receiver-script SSH_USER REPO KEYNAME FINGERPRINT
package main

import (
    "crypto/sha1"
	"flag"
	"fmt"
    "github.com/rochacon/cargo/nginx"
    "log"
	"os"
    "os/exec"
	"math/rand"
	"strings"
	"time"
)

const IMAGE_CACHE = "/tmp/app-cache"

func getRandomPort() (port int) {
	rand.Seed(time.Now().UnixNano())
	for ;port <= 1024; {
		port = rand.Intn(65534)
	}
	return
}

func main() {
    base_domain := flag.String("d", "localhost", "Base domain")
    bucket_name := flag.String("bucket", "", "AWS S3 bucket name")
	flag.Parse()

	if len(flag.Args()) < 4 {
		 flag.Usage()
		 os.Exit(1)
	}

    app_name := flag.Arg(1)
    app_name_sha1 := sha1.Sum([]byte(app_name))
    // log.Println("app_name_sha1", fmt.Sprintf("%x", app_name_sha1))
    app_name_for_url := fmt.Sprintf("%x", app_name_sha1)[:10]
    // log.Println("app_name_for_url", app_name_for_url)

	// XXX change this to your bucket URL and allow access for GET and PUT keys to your server
	var bucket = "https://s3.amazonaws.com/" + *bucket_name
	var image_name = fmt.Sprintf("%s/%s.tgz", bucket, app_name_for_url)

	var container_port = getRandomPort()

	fmt.Println("-----> Builing image")
    // TODO use slugbuilder cache
    slugbuilder := exec.Command("docker", strings.Split("run -i -a stdin -a stdout flynn/slugbuilder " + image_name, " ")...)
    slugbuilder.Stdout = os.Stdout
    slugbuilder.Stdin = os.Stdin
    if err := slugbuilder.Run(); err != nil {
        log.Fatal(err)
    }

	// TODO read Procfile and get processes
	// processes := []string{"web", "worker"}
	// fmt.Println("-----> Starting containers:", strings.Join(processes, ", "))

	// FIXME run all processes
	// for _, process := range processes {
	//   go docker.Run(-d -i -e SLUG_URL="$IMAGE" -e PORT=8000 -p $PORT:8000 flynn/slugrunner start web)
	// }
    runner_opts := strings.Split(fmt.Sprintf("run -d -i -e SLUG_URL=%s -e PORT=8000 -p %d:8000 flynn/slugrunner start web", image_name, container_port), " ")
    // log.Println("main", "slugrunner", "runner_opts", runner_opts)
    slugrunner := exec.Command("docker", runner_opts...)
    slugrunner.Stdout = os.Stdout
    slugrunner.Stdin = os.Stdin
    if err := slugrunner.Run(); err != nil {
        log.Fatal(err)
    }

	// TODO for web processes
	// TODO inspect container and retrieve IP (no need to expose container port)
	var hostname = fmt.Sprintf("%s.%s", app_name_for_url, *base_domain)
    var container_ip = "127.0.0.1"

    // log.Println("main", "container_ip", container_ip, "container_port", container_port)

	// add container as a server to local NGINX
    nginx.AddServer(
        app_name,
        []string{fmt.Sprintf("%s:%d", container_ip, container_port)},
        hostname,
    )
    nginx.Reload()

    fmt.Println("-----> http://" + hostname)
}
