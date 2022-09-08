package server

// repo层只判断或者返回指定的error, service记录err的log
var repoTemplateGrpc = `
{{- /* delete empty line */ -}}
package repo

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	commonPb "git.hiscene.net/hiar_mozi/server/mozi-common/api/mozi/common/v1"  {{- if eq 2 1 }} 注释,此行不需要跟随项目变动, common里的pb定义,包括err {{ end }}
	{{ .GrpcPbName }} "{{ .Package }}" {{- if eq 2 1 }} 注释, pb {{ end }}
	"{{ .InternalPackage }}/conf"
	"{{ .InternalPackage }}/data"
	"{{ .InternalPackage }}/domain"
	mozitools "git.hiscene.net/hiar_mozi/server/mozi-common/tools" {{- if eq 2 1 }} 注释,此行不需要跟随项目变动 {{ end }}
	"{{ .InternalPackage }}/pkg/middleware" {{- if eq 2 1 }} 注释,todo 可以改为不跟随项目变动 {{ end }}
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
	{{- if eq 2 1 }}注释 ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(s.bootstrap.Data.Mysql.Timeout))
	defer cancel()
	// 请移除下行代码, 换成实际使用, 否则删除上面的语句
	ctxTimeout=ctxTimeout 
	{{ end }}
	// 操作者信息
	_, _, e := middleware.GetUserInfo(ctx, 0, "")
	if e != nil {
		cErr := mozitools.CustomeErr{ErrCode: commonPb.ErrorReason_Unauthenticated}
		cErr.Err = e
		return {{ if eq .Reply $s1 }}{{ else }}nil, {{ end }}cErr
		// return nil, hiKratos.ResponseErr(ctx, commonPb.ErrorUnauthenticated)
	}
	{{ if eq .Reply $s1 }}{{ else }}rep := &{{ .GrpcPbName }}.{{ .Reply }}{}{{ end }}
	
	return {{ if eq .Reply $s1 }}{{ else }}rep, {{ end }}nil
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
