package server

var repoTemplateGrpc = `
{{- /* delete empty line */ -}}
package repo

import (
	{{- if .UseContext }}
	"context"
	"time"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	{{ .GrpcPbName }} "{{ .GrpcPackage }}"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	"{{ .InternalPackage }}/conf"
	"{{ .InternalPackage }}/data"
	"{{ .InternalPackage }}/domain"
)

type {{ .Service }}Repo struct {
	// pb.Unimplemented{{ .Service }}Server
	data *data.Data
	bootstrap *conf.Bootstrap
}

func New{{ .Service }}Repo(data *data.Data, bootstrap *conf.Bootstrap) domain.I{{ .Service }}Repo {
	return &{{ .Service }}Repo{
		data: data,
		bootstrap: bootstrap,
	}
}

{{- $s1 := "google.protobuf.Empty" }}
{{ range .Methods }}
{{- if eq .Type 1 }}
func (s *{{ .Service }}Repo) {{ .Name }}(ctx context.Context{{ if eq .Request $s1 }}{{ else }}, req *{{ .GrpcPbName }}.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}{{ else }}*{{ .GrpcPbName }}.{{ .Reply }},{{ end }} error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(s.bootstrap.Data.Mysql.Timeout))
	defer cancel()

	return {{ if eq .Reply $s1 }}{{ else }}&{{ .GrpcPbName }}.{{ .Reply }}{}, {{ end }}nil
	// return {{ if eq .Reply $s1 }}{{ else }}&{{ .GrpcPbName }}.{{ .Reply }}{},{{ end }} nil
}

{{- else if eq .Type 2 }}
func (s *{{ .Service }}Repo) {{ .Name }}(conn {{ .GrpcPbName }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		
		err = conn.Send(&{{ .GrpcPbName }}.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 3 }}
func (s *{{ .Service }}Repo) {{ .Name }}(conn {{ .GrpcPbName }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return conn.SendAndClose(&{{ .GrpcPbName }}.{{ .Reply }}{})
		}
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 4 }}
func (s *{{ .Service }}Repo) {{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*{{ .GrpcPbName }}.{{ .Request }}{{ end }}, conn {{ .GrpcPbName }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		err := conn.Send(&{{ .GrpcPbName }}.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- end }}
{{- end }}
`
