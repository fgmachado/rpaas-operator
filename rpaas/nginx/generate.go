package nginx

import (
	"bytes"
	"text/template"

	"github.com/tsuru/rpaas-operator/pkg/apis/extensions/v1alpha1"
)

type ConfigurationRenderer interface {
	Render(ConfigurationData) (string, error)
}

type ConfigurationData struct {
	Config *v1alpha1.NginxConfig
}

type rpaasConfigurationRenderer struct {
	t *template.Template
}

func (r *rpaasConfigurationRenderer) Render(c ConfigurationData) (string, error) {
	buffer := &bytes.Buffer{}
	err := r.t.Execute(buffer, c)
	return buffer.String(), err
}

func NewRpaasConfigurationRenderer() ConfigurationRenderer {
	return &rpaasConfigurationRenderer{
		t: template.Must(template.New("rpaas-configuration-template").Parse(rawNginxConfiguration)),
	}
}

var rawNginxConfiguration = `
# This file was generated by RPaaS (https://github.com/tsuru/rpaas-operator.git)
# Do not modify this file, any change will be lost.

user {{with .Config.User}}{{.}}{{else}}nginx{{end}};
worker_process {{with .Config.WorkerProcesses}}{{.}}{{else}}1{{end}};

include modules/*.conf;

events {
    worker_connections {{with .Config.WorkerConnections}}{{.}}{{else}}1024{{end}};
}

http {
    include       mime.types;
    default_type  application/octet-stream;
    server_tokens off;

    sendfile          on;
    keepalive_timeout 65;

{{if .Config.RequestIDEnabled}}
    uuid4 $request_id_uuid;
    map $http_x_request_id $request_id_final {
        default $request_id_uuid;
        "~."    $http_x_request_id;
    }
{{end}}

    map $http_x_real_ip $real_ip_final {
        default $remote_addr;
        "~."    $http_x_real_ip;
    }

    map $http_x_forwarded_proto $forwarded_proto_final {
        default $scheme;
        "~."    $http_x_forwarded_proto;
    }

    map $http_x_forwarded_host $forwarded_host_final {
        default $host;
        "~." $http_x_forwarded_host;
    }

    log_format rpaas_combined
        '${remote_addr}\t${host}\t${request_method}\t${request_uri}\t${server_protocol}\t'
        '${http_referer}\t${http_x_mobile_group}\t'
        'Local:\t${status}\t*${connection}\t${body_bytes_sent}\t${request_time}\t'
        'Proxy:\t${upstream_addr}\t${upstream_status}\t${upstream_cache_status}\t'
        '${upstream_response_length}\t${upstream_response_time}\t${request_uri}\t'
{{if .Config.RequestIDEnabled}}
        'Agent:\t${http_user_agent}\t$request_id_final\t'
{{else}}
        'Agent:\t${http_user_agent}\t'
{{end}}
        'Fwd:\t${http_x_forwarded_for}';

{{if .Config.SyslogEnabled}}
    access_log syslog:server={{.Config.SyslogServerAddress}},facility={{with .Config.SyslogFacility}}{{.}}{{else}}local6{{end}},tag={{if .Config.SyslogTag}}{{.Config.SyslogTag}}{{else}}rpaas{{end}} rpaas_combined;
    error_log syslog:server={{.Config.SyslogServerAddress}},facility={{with .Config.SyslogFacility}}{{.}}{{else}}local6{{end}},tag={{if .Config.SyslogTag}}{{.Config.SyslogTag}}{{else}}rpaas{{end}};
{{else}}
    access_log /dev/stdout rpaas_combined;
    error_log  /dev/stderr;
{{end}}

{{if .Config.CacheEnabled}}
    proxy_cache_path {{.Config.CachePath}}/nginx levels=1:2 keys_zone=rpaas:{{.Config.CacheZoneSize}} inactive={{.Config.CacheInactive}} max_size={{.Config.CacheSize}} loader_files={{.Config.CacheLoaderFiles}};
    proxy_temp_path  {{.Config.CachePath}}/nginx_temp 1 2;
{{end}}

    gzip                on;
    gzip_buffers        128 4k;
    gzip_comp_level     5;
    gzip_http_version   1.0;
    gzip_min_length     20;
    gzip_proxied        any;
    gzip_vary           on;
    gzip_types          application/atom+xml application/javascript
                        application/json application/rss+xml
                        application/xml application/x-javascript
                        text/css text/javascript text/plain text/xml;

{{if .Config.VTSEnabled}}
    vhost_traffic_status_zone
{{end}}

    server {
        listen {{with .Config.HTTPPort}}{{.}}{{else}}80{{end}} default_server{{with .Config.HTTPListenOptions}} {{.}}{{end}};
    }
}
`