package nginx

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

const tmpl = `
upstream {{.UpstreamName}} {
    {{range .Servers}}
    server {{.}} ;
    {{end}}
}

server {
    listen 80;
    server_name {{.Hostname}};

    location / {
        proxy_pass http://{{.UpstreamName}};
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Port $http_port;
    }
}
`

type Upstream struct {
	UpstreamName string
	Servers      []string
	Hostname     string
}

func AddServer(upstreamname string, servers []string, hostname string) error {
	t := template.Must(template.New("upstream").Parse(tmpl))

	serverfile := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", hostname)
	fp, err := os.Create(serverfile)
	if err != nil {
		return err
	}
	defer fp.Close()

	u := Upstream{
		upstreamname,
		servers,
		hostname,
	}
	err = t.Execute(fp, u)
	if err != nil {
		return err
	}
	return nil
}

func Reload() error {
	cmd := exec.Command("sudo", "/usr/sbin/nginx", "-s", "reload")
	return cmd.Run()
}
