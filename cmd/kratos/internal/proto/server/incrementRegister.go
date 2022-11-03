package server

import (
	"fmt"
	"log"
	"os"
	"path"
)

// https://www.jianshu.com/p/937d649039ec

// 增量处理
/*
	internal/prottocol/register.go
*/
var (
	lineAssignRegister     = "*PbServer-RegisterHTTP"
	lineAssignRegisterGrpc = "*PbServer-RegisterRPC"
	// 不能变动, 同label
	clientAssignKey = "clientAssign"
	// 当前服务的module, 例 "git.hiscene.net/hiar_mozi/server/mozi-device-service/internal/module/device_org"
	wireImport                     = "module-service"
	wireAutoWireAppPanicBuildParam = "wireAutoWireAppPanicBuildParam"
	registerPbImport               = "pb-import"
	dataPbImport                   = "data-pb-import"
)

// 增加方法
//  在 // service grpc 之后添加?
//  0. import 检查是否已存在pb
//  1. PbServer 里添加
// 		grpc 服务里，不需要http
// 		TaskVideoPic pbInfo.TaskVideoPicHTTPServer
//  2. RegisterHTTP 里添加
// 		pbInfo.RegisterTaskVideoPicHTTPServer(srv, s.TaskVideoPic)
func incrementRegister(pbPkgName, fileDir, fileName string, sourceFile *SourceFile, s *Service) bool {

	filePath := path.Join(fileDir, fileName)
	bs, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	// todo: infopb 需要参数告知

	// 3. RPC注册, 这三则顺序需要代码里确认
	// down todo: RegisterRPC grpc会重复，需要确认
	if s.Layer == "grpc" {
		registerService := fmt.Sprintf("\n"+`	//Code generated by %s. (语句已存在则不覆盖).
	%s.Register%[3]sServer(srv, s.%[3]s)`,
			s.ToolName, pbPkgName, s.Service)
		if !sourceFile.LineExist[lineAssignRegisterGrpc] {
			if v, ok := sourceFile.Methods[lineAssignRegisterGrpc]; ok {
				pre := bs[:v.EndPos-len(getEndOfLine()+"}")]
				next := append([]byte{}, bs[v.EndPos-len(getEndOfLine()+"}"):]...)
				bs = append(pre, []byte(registerService)...)
				bs = append(bs, next...)
			} else {
				fmt.Printf("没有找到func %s\n", lineAssignRegisterGrpc)
			}
		} else {
			fmt.Printf("设置 Register service语句 %s 已存在\n", registerService)
		}
	} else {
		// http时, 使用当前唯一一个服务的包名, 约定为pb
		pbPkgName = "pb"
	}
	// 2. HTTP注册
	registerService := fmt.Sprintf("\n"+`	//Code generated by %s. (语句已存在则不覆盖).
	%s.Register%[3]sHTTPServer(srv, s.%[3]s)`,
		s.ToolName, pbPkgName, s.Service)
	if !sourceFile.LineExist[lineAssignRegister] {
		if v, ok := sourceFile.Methods[lineAssignRegister]; ok {
			pre := bs[:v.EndPos-len(getEndOfLine()+"}")]
			next := append([]byte{}, bs[v.EndPos-len(getEndOfLine()+"}"):]...)
			bs = append(pre, []byte(registerService)...)
			bs = append(bs, next...)
		} else {
			fmt.Printf("没有找到func %s\n", lineAssignRegister)
		}
	} else {
		fmt.Printf("设置 Register service语句 %s 已存在\n", registerService)
	}
	// 1. type PbServer struct {
	if v, ok := sourceFile.Structs["PbServer"]; ok {
		fieldKey := fmt.Sprintf("%s-%s.%sHTTPServer", s.Service, pbPkgName, s.Service)
		if s.Layer == "grpc" {
			fieldKey = fmt.Sprintf("%s-%s.%sServer", s.Service, pbPkgName, s.Service)
		}
		if _, ok := v.Fields[fieldKey]; !ok {
			dataField := fmt.Sprintf(`	//Code generated by %s. (字段已存在则不覆盖).
	%s	%s.%sHTTPServer`, s.ToolName, s.Service, pbPkgName, s.Service)
			if s.Layer == "grpc" {
				dataField = fmt.Sprintf(`	//Code generated by %s. (字段已存在则不覆盖).
	%s	%s.%sServer`, s.ToolName, s.Service, pbPkgName, s.Service)
			}
			var pre []byte
			var next []byte
			if v.LastFieldEndPos != 0 {
				pre = bs[:v.LastFieldEndPos]
				next = append([]byte{}, bs[v.LastFieldEndPos:]...)
			} else {
				pre = bs[:v.EndPos-3]
				next = append([]byte{}, bs[v.EndPos-3:]...)
				dataField = "\n" + dataField
			}
			bs = append(pre, []byte(dataField)...)
			bs = append(bs, next...)
		} else {
			fmt.Printf("PbServer field %s 已存在\n", fieldKey)
		}
	}

	// 0. import {{ .GrpcPbName }} "{{ .Package }}" {{- if eq 2 1 }} 注释, pb {{ end }}
	importLine := fmt.Sprintf(`	//Code generated by %s. (字段已存在则不覆盖).
	%s "%s"`, s.ToolName, s.GrpcPbName, s.Package)
	if !sourceFile.LineExist[registerPbImport] {
		pre := bs[:sourceFile.Imports.EndPos]
		next := append([]byte{}, bs[sourceFile.Imports.EndPos:]...)
		bs = append(pre, []byte(importLine)...)
		var newLine byte = '\n' //增加换行
		bs = append(bs, newLine)
		bs = append(bs, next...)

	} else {
		fmt.Printf("register pb import语句 %s 已存在\n", importLine)
	}

	// test
	filePath = path.Join(fileDir, fileName)
	if err := os.WriteFile(filePath, bs, 0o644); err != nil {
		log.Fatal(err)
	}

	return true
}
