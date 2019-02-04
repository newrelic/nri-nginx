package main

import (
	"bufio"
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"strings"
	"testing"
)

var (
	testNginxConf = `
# This is a comment that should be ignored

user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log;
pid /run/nginx.pid;

events {
  worker_connections 1024;
}

http {
  log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';
  access_log  /var/log/nginx/access.log  main;

  sendfile            on;
  tcp_nopush          on;
  tcp_nodelay         on;
  keepalive_timeout   65;
  types_hash_max_size 2048;

  default_type        application/octet-stream;

  server {
    listen       80 default_server;
    listen       [::]:80 default_server;
    server_name  www.example.com;
    root         /usr/share/nginx/html;

    location / {
    }
    error_page 404 /404.html;
      location = /40x.html {
    }
    error_page 500 502 503 504 /50x.html;
      location = /50x.html {
    }

    location /status {
      stub_status on;
      access_log off;
      allow 192.168.100.0/24;
      deny all;
    }
  }
}`
)

func TestParseNginxConf(t *testing.T) {

	i := inventory.New()

	err := populateInventory(bufio.NewReader(strings.NewReader(testNginxConf)), i)

	if err != nil {
		t.Fatal()
	}

	if i.Items()["pid"]["value"] != "/run/nginx.pid" {
		t.Error()
	}
	if i.Items()["events/worker_connections"]["value"] != "1024" {
		t.Error()
	}
	if i.Items()["http/server/server_name"]["value"] != "www.example.com" {
		t.Error()
	}
	if i.Items()["http/server/location::status/allow"]["value"] != "192.168.100.0/24" {
		t.Error()
	}

}
