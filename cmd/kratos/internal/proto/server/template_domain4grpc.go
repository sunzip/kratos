package server

var domainTemplateGrpc = `
{{- /* delete empty line */ -}}
// Code generated by  {{ .ToolName }}. (仅初始化, 已存在则不覆盖).
// versions:
// 	v0.1
// source: {{ .SourceProto }}

package domain

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}
	// {{ .GrpcPbName }} "{{ .GrpcPackage }}"
	{{ .HttpPbName }} "{{ .Package }}"

)

// 参数和类型可改
type I{{ .Service }}Repo interface {
	{{- $s1 := "google.protobuf.Empty" }}	
	{{- /* delete empty line */ -}}
	{{ range .Methods }}
	{{- if eq .Type 1 }}
	{{ .Name }}(ctx context.Context{{ if eq .Request $s1 }}{{ else }}, req *{{ .GrpcPbName }}.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}{{ else }}*{{ .GrpcPbName }}.{{ .Reply }}, {{ end }}error) 

	{{- else if eq .Type 2 }}
	{{ .Name }}(conn {{ .GrpcPbName }}.{{ .Service }}_{{ .Name }}Server) error 

	{{- else if eq .Type 3 }}
	{{ .Name }}(conn {{ .GrpcPbName }}.{{ .Service }}_{{ .Name }}Server) error

	{{- else if eq .Type 4 }}
	{{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
	{{ else }}*{{ .GrpcPbName }}.{{ .Request }}{{ end }}, conn {{ .GrpcPbName }}.{{ .Service }}_{{ .Name }}Server) error

	{{- end }}
	{{- end }}
}

// 参数和类型可改
type I{{ .Service }}Service interface {
	{{- $s1 := "google.protobuf.Empty" }}	
	{{- /* delete empty line */ -}}
	{{ range .Methods }}
	{{- if eq .Type 1 }}
	{{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty
	{{ else }}*{{ .HttpPbName }}.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*{{ .HttpPbName }}.{{ .Reply }}{{ end }}, error) 

	{{- else if eq .Type 2 }}
	{{ .Name }}(conn {{ .HttpPbName }}.{{ .Service }}_{{ .Name }}Server) error 

	{{- else if eq .Type 3 }}
	{{ .Name }}(conn {{ .HttpPbName }}.{{ .Service }}_{{ .Name }}Server) error

	{{- else if eq .Type 4 }}
	{{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
	{{ else }}*{{ .HttpPbName }}.{{ .Request }}{{ end }}, conn {{ .HttpPbName }}.{{ .Service }}_{{ .Name }}Server) error

	{{- end }}
	{{- end }}
}

// 生成实体
type {{ .Service }} struct{

}
`