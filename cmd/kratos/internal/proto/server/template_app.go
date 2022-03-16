package server

import (
	"bytes"
	"html/template"
)

var appTemplate = `
{{- /* delete empty line */ -}}
package app

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	pb "{{ .Package }}"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	"github.com/go-kratos/kratos/v2/log"
	// "{{ .ModulePackage }}/{{ .ServiceLower }}/service"
	"{{ .InternalPackage }}/domain"
)

type {{ .Service }}App struct {
	// pb.Unimplemented{{ .Service }}Server
	logger *log.Helper
	service    domain.I{{ .Service }}Service
}

func New{{ .Service }}App(logger log.Logger, service domain.I{{ .Service }}Service) pb.{{ .Service }}HTTPServer {
	return &{{ .Service }}App{
		logger:  log.NewHelper(logger),
		service:    service,
	}
}

{{- $s1 := "google.protobuf.Empty" }}
{{ range .Methods }}
{{- if eq .Type 1 }}
func (s *{{ .Service }}App) {{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*pb.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Reply }}{{ end }}, error) {
	return {{ if eq .Reply $s1 }}&emptypb.Empty{}{{ else }}s.service.{{ .Name }}(ctx,req){{ end }}
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

func (s *Service) executeApp() ([]byte, error) {
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
	tmpl, err := template.New("app").Parse(appTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
