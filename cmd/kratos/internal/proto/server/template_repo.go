package server

import (
	"bytes"
	"html/template"
)

var repoTemplate = `
{{- /* delete empty line */ -}}
package repo

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}

	{{ .GrpcPbName }} "{{ .GrpcPackage }}"
	conf "git.hiscene.net/hiar_mozi/server/mozi-common/model/internal_conf"  {{- if eq 2 1 }} 注释,此行不需要跟随项目变动, common里的pb定义,包括err {{ end }}
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
func (s *{{ .Service }}Repo) {{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*{{ .GrpcPbName }}.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*{{ .GrpcPbName }}.{{ .Reply }}{{ end }}, error) {
	
	return {{ if eq .Reply $s1 }}&emptypb.Empty{}, nil{{ else }}s.data.{{ .Service }}Grpc().{{ .Name }}(ctx, req){{ end }}
	// return {{ if eq .Reply $s1 }}&emptypb.Empty{}{{ else }}&{{ .GrpcPbName }}.{{ .Reply }}{}{{ end }}, nil
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

func (s *Service) executeRepo() ([]byte, error) {
	const empty = "google.protobuf.Empty"
	buf := new(bytes.Buffer)
	for _, method := range s.Methods {
		if (method.Type == unaryType && (method.Request == empty || method.Reply == empty)) ||
			(method.Type == returnsStreamsType && method.Request == empty) {
			s.GoogleEmpty = true
		}
		if method.Type == twoWayStreamsType || method.Type == requestStreamsType {
			s.UseIO = true
		}
		if method.Type == unaryType {
			s.UseContext = true
		}
	}
	tmpl, err := template.New("repo").Parse(repoTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
