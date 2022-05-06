package server

var appTemplateGrpc = `
{{- /* delete empty line */ -}}
package app

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	{{ .HttpPbName }} "{{ .Package }}"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	"github.com/go-kratos/kratos/v2/log"
	// "{{ .ModulePackage }}/{{ .ServiceLower }}/service"
	"{{ .InternalPackage }}/domain"
)

type {{ .Service }}App struct {
	pb.Unimplemented{{ .Service }}Server
	logger *log.Helper
	service    domain.I{{ .Service }}Service
}

func New{{ .Service }}App(logger log.Logger, service domain.I{{ .Service }}Service) pb.{{ .Service }}Server {
	return &{{ .Service }}App{
		logger:  log.NewHelper(logger),
		service:    service,
	}
}

{{- $s1 := "google.protobuf.Empty" }}
{{ range .Methods }}
{{- if eq .Type 1 }}
func (s *{{ .Service }}App) {{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Reply }}{{ end }}, error) {
	return s.service.{{ .Name }}(ctx,req)
}

{{- else if eq .Type 2 }}
func (s *{{ .Service }}App) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		
		err = conn.Send(&pb.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 3 }}
func (s *{{ .Service }}App) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return conn.SendAndClose(&pb.{{ .Reply }}{})
		}
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 4 }}
func (s *{{ .Service }}App) {{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*pb.{{ .Request }}{{ end }}, conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		err := conn.Send(&pb.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- end }}
{{- end }}
`
