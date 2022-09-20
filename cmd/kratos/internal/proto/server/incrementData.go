package server

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path"
	"strings"
)

// https://www.jianshu.com/p/937d649039ec

// 增量处理
/*
	internal/prottocol/register.go
	internal/data/data.go
	cmd/server/wire.go
*/

// 增加方法
//  在 // service grpc 之后添加?
//  1. Data里添加
// 		oplogCli     infopb.OpLogClient
//  2. NewData 里添加
// 		d.infoCli = infopb.NewFixedPoiClient(connInfo)
/*
需要有代码
clientAssign:
if false {
	// 生成代码使用
	goto clientAssign
}
*/
//  3. NewData后面添加
// 		func (d *Data) OpLogGrpc() *infopb.OpLogClient {
// 			return d.oplogCli
// 		}

// 只在http服务里有
func incrementMethodData(fileDir, fileName string, sourceFile *SourceFile, s *Service, pbAlias string) bool {

	filePath := path.Join(fileDir, fileName)
	bs, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	// todo: infopb 需要参数告知
	// 3. 添加获取方法
	methodCodeBlock := fmt.Sprintf(`//Code generated by %s. (字段已存在则不覆盖).
func (d *Data) %sGrpc() %s.%sClient {
	return d.%sCli
}`, s.ToolName, s.Service, pbAlias, s.Service, s.ServiceLower)
	methodKey := fmt.Sprintf("*Data-%sGrpc", s.Service)

	if _, ok := sourceFile.Methods[methodKey]; !ok {
		newDataMethodKey := "-NewData"
		if v, ok := sourceFile.Methods[newDataMethodKey]; ok {
			pre := bs[:v.EndPos]
			next := append([]byte{}, bs[v.EndPos:]...)
			bs = append(pre, []byte(getEndOfLine()+methodCodeBlock)...)
			bs = append(bs, next...)
		} else {
			bs = append(bs, []byte(getEndOfLine()+methodCodeBlock)...)
		}
	} else {
		fmt.Printf("get Grpc client方法 %s 已存在\n", methodKey)
	}
	// 2. func NewData 内部
	clientAssign := fmt.Sprintf("	d.%sCli = %s.New%sClient(connInfo)", s.ServiceLower, s.GrpcPbName, s.Service)
	if !sourceFile.LineExist[clientAssignKey] {
		if v, ok := sourceFile.Labels[clientAssignKey]; ok {
			pre := bs[:v.EndPos]
			next := append([]byte{}, bs[v.EndPos:]...)
			bs = append(pre, []byte(getEndOfLine()+clientAssign)...)
			bs = append(bs, next...)
		} else {
			fmt.Printf("没有找到label %s\n", "clientAssign")
		}
	} else {
		fmt.Printf("设置 Grpc client语句 %s 已存在\n", clientAssign)
	}
	// 1. type Data struct {
	if v, ok := sourceFile.Structs["Data"]; ok {
		fieldKey := fmt.Sprintf("%sCli-%s.%sClient", s.ServiceLower, s.GrpcPbName, s.Service)
		if _, ok := v.Fields[fieldKey]; !ok {
			dataField := fmt.Sprintf(`	//Code generated by %s. (字段已存在则不覆盖).
	%sCli     %s.%sClient`+"\n", s.ToolName, s.ServiceLower, s.GrpcPbName, s.Service)
			pre := bs[:v.LastFieldEndPos]
			next := append([]byte{}, bs[v.LastFieldEndPos:]...)
			bs = append(pre, []byte(dataField)...)
			bs = append(bs, next...)
		} else {
			fmt.Printf("Data field %s 已存在\n", fieldKey)
		}
	}

	// 0. import {{ .GrpcPbName }} "{{ .Package }}" {{- if eq 2 1 }} 注释, pb {{ end }}
	importLine := fmt.Sprintf(`	//Code generated by %s. (字段已存在则不覆盖).
	%s "%s"`, s.ToolName, s.GrpcPbName, s.GrpcPackage) //使用common里的package
	if !sourceFile.LineExist[dataPbImport] {
		pre := bs[:sourceFile.Imports.EndPos]
		next := append([]byte{}, bs[sourceFile.Imports.EndPos:]...)
		bs = append(pre, []byte(importLine)...)
		var newLine byte = '\n' //增加换行
		bs = append(bs, newLine)
		bs = append(bs, next...)

	} else {
		fmt.Printf("data pb import语句 %s 已存在\n", importLine)
	}

	// test
	filePath = path.Join(fileDir, fileName)
	if err := os.WriteFile(filePath, bs, 0o644); err != nil {
		log.Fatal(err)
	}

	return true
}

// 判断是否已经存在相同的服务名称
func parseFile(fileDir string, fileName string, s *Service) *SourceFile {
	filePath := path.Join(fileDir, fileName)
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	sourceFile := &SourceFile{}
	sourceFile.LineExist = make(map[string]bool, 5)
	sourceFile.LineExist[clientAssignKey] = false
	// ...
	sourceFile.Methods = make(map[string]MethodInfo, 5)
	sourceFile.Structs = make(map[string]StructInfo, 5)
	sourceFile.Comments = make(map[string]CommentIfno, 5)
	sourceFile.Labels = make(map[string]LabelIfno, 5)
	ast.Inspect(f, func(n ast.Node) bool {
		// Find Return Statements
		// ret, ok := n.(*ast.ReturnStmt)
		switch ret := n.(type) {
		case *ast.Comment: //, *ast.CommentGroup
			// fmt.Println(ret)
			// printer.Fprint(os.Stdout, fset, ret)
			// fmt.Println()
			commStr := ret.Text
			comment := CommentIfno{}
			comment.StartPos = int(ret.Pos())
			comment.EndPos = int(ret.End())
			comment.Comment = commStr
			commKey := strings.ReplaceAll(commStr, "//", "")
			commKey = strings.TrimSpace(commKey)
			commKeys := strings.SplitN(commKey, " ", 2)
			if len(commKeys) > 0 {
				commKey = commKeys[0]
				sourceFile.Comments[commKey] = comment
			}
		case *ast.LabeledStmt:
			// fmt.Println(ret)
			label := LabelIfno{}
			label.Text = ret.Label.Name
			label.StartPos = int(ret.Pos())
			label.EndPos = int(ret.Colon)
			sourceFile.Labels[ret.Label.Name] = label
		case *ast.AssignStmt:
			// fmt.Println(ret)
			// printer.Fprint(os.Stdout, fset, ret)
			b := bytes.Buffer{}
			printer.Fprint(&b, fset, ret)
			line := strings.TrimSpace(b.String())
			if s.Layer == "http" {
				want := fmt.Sprintf("d.%sCli = %s.New%sClient(connInfo)", s.ServiceLower, s.GrpcPbName, s.Service)
				if line == want {
					sourceFile.LineExist[clientAssignKey] = true
				}

			}
		case *ast.BlockStmt:

			/* fmt.Println(ret)
			printer.Fprint(os.Stdout, fset, ret)
			fmt.Println() */
		}

		return true
	})
	ast.Inspect(f, func(n ast.Node) bool {
		// Find Return Statements
		// ret, ok := n.(*ast.ReturnStmt)
		switch ret := n.(type) {
		case *ast.FuncDecl:
			// fmt.Printf("return statement found on line %v:\n", fset.Position(ret.Pos()))
			if ret.Name.Name == "autoWireApp" {
				if len(ret.Body.List) > 0 {
					panic := ret.Body.List[0].(*ast.ExprStmt)
					params := GetExprPara(panic.X)
					if len(params) > 0 {
						build, ok := params[0].(*ast.CallExpr)
						if ok {
							params := GetExprPara(build)
							length := len(params)
							if length > 0 {
								want := fmt.Sprintf("%s.ProviderSet", s.ServiceLower)
								lastParam := params[length-1]
								sourceFile.AutoWireAppPanicBuildLastParam.StartPos = int(lastParam.Pos()) - len("n")
								sourceFile.AutoWireAppPanicBuildLastParam.EndPos = int(lastParam.End())
								for _, item := range params {
									bParam := bytes.Buffer{}
									printer.Fprint(&bParam, fset, item)
									if want == bParam.String() {
										sourceFile.LineExist[wireAutoWireAppPanicBuildParam] = true
									}
								}
							}
						}
					}

				}
			} else {
				if ret.Name.Name == "RegisterHTTP" {
					if len(ret.Body.List) > 0 {
						pbPkgName := importAlias

						registerService := fmt.Sprintf(`%s.Register%[2]sHTTPServer(srv, s.%[2]s)`,
							pbPkgName, s.Service)

						for _, item := range ret.Body.List {
							registerLine := bytes.Buffer{}
							printer.Fprint(&registerLine, fset, item)
							if registerService == registerLine.String() {
								sourceFile.LineExist[lineAssignRegister] = true
							}
						}
					}
				} else if ret.Name.Name == "RegisterRPC" {
					if len(ret.Body.List) > 0 {
						pbPkgName := importAlias

						registerService := fmt.Sprintf(`%s.Register%[2]sServer(srv, s.%[2]s)`,
							pbPkgName, s.Service)
						for _, item := range ret.Body.List {
							registerLine := bytes.Buffer{}
							printer.Fprint(&registerLine, fset, item)
							if registerService == registerLine.String() {
								sourceFile.LineExist[lineAssignRegisterGrpc] = true
							}
						}
					}
				}
				body := ret.Body
				doc := ret.Doc
				ret.Body = nil
				ret.Doc = nil
				b := bytes.Buffer{}
				printer.Fprint(&b, fset, ret)
				// 使用过以后，还原
				ret.Body = body
				ret.Doc = doc

				var recvType string
				if ret.Recv != nil {
					bType := bytes.Buffer{}
					printer.Fprint(&bType, fset, ret.Recv.List[0].Type)
					// fmt.Println(bType.String())
					recvType = bType.String()
				}
				bName := bytes.Buffer{}
				printer.Fprint(&bName, fset, ret.Name)
				mapKey := fmt.Sprintf("%s-%s", recvType, bName.String())
				mi := MethodInfo{}
				mi.MethodDefind = b.String()
				mi.StartPos = int(ret.Pos())
				mi.EndPos = int(ret.End())
				sourceFile.Methods[mapKey] = mi
				// printer.Fprint(os.Stdout, fset, ret)
				// fmt.Println()
			}
			return true
		case *ast.StructType:
			// type Data struct {
			// 但是 拿不到 type Data
			{
				// fmt.Println(ret)
			}
		case *ast.DeclStmt:
			{
				// 不知
			}
		case *ast.TypeSpec:
			// type Data struct {
			{
				// fmt.Println(ret)
				si := StructInfo{}
				si.StartPos = int(ret.Pos())
				// si.LastFieldPos = int(ret.Type)
				if retStructType, ok := ret.Type.(*ast.StructType); ok {

					len := len(retStructType.Fields.List)
					if len > 0 {
						si.LastFieldEndPos = int(retStructType.Fields.List[len-1].End())
						si.Fields = make(map[string]bool, 5)
						for _, item := range retStructType.Fields.List {
							typeName := bytes.Buffer{}
							printer.Fprint(&typeName, fset, item.Type)
							fieldKey := fmt.Sprintf("%s-%s", item.Names[0].Name, typeName.String())
							si.Fields[fieldKey] = true

						}
					}
				}
				si.EndPos = int(ret.End())
				sourceFile.Structs[ret.Name.Name] = si
			}
		case *ast.ImportSpec:
			// 注意, ret.Path.Value不包括别名
			// 1. Name=别名
			// 2. doc 是包上方的文档注释
			// 3. comment 是包后面, 行内的注释
			{
				// wire import
				// fmt.Println(ret)
				moduleService := changeDir(s.Package, "/api/", fmt.Sprintf("/internal/module/%s", s.ServiceLower))
				want := fmt.Sprintf(`"%s"`, moduleService)
				if ret.Path.Value == want {
					sourceFile.LineExist[wireImport] = true
				}
				sourceFile.Imports.StartPos = int(ret.Pos())
				if ret.Comment != nil {
					if ret.Comment.End().IsValid() {
						sourceFile.Imports.EndPos = int(ret.Comment.End())
					}
				} else {
					sourceFile.Imports.EndPos = int(ret.Path.Pos()) + len(ret.Path.Value)
				}
				sourceFile.Imports.Text = ret.Path.Value
			}
			{
				// register import
				// 使用本服务(http,grpc)
				want := fmt.Sprintf(`"%s"`, s.Package)
				if ret.Path.Value == want {
					sourceFile.LineExist[registerPbImport] = true
				}
				sourceFile.Imports.StartPos = int(ret.Pos())
				if ret.Comment != nil {
					if ret.Comment.End().IsValid() {
						sourceFile.Imports.EndPos = int(ret.Comment.End())
					}
				} else {
					sourceFile.Imports.EndPos = int(ret.Path.Pos()) + len(ret.Path.Value)
				}
				sourceFile.Imports.Text = ret.Path.Value
			}
			{
				// data import(只有http使用, 且是common服务)
				want := fmt.Sprintf(`"%s"`, s.GrpcPackage)
				if ret.Path.Value == want {
					sourceFile.LineExist[dataPbImport] = true
				}
				sourceFile.Imports.StartPos = int(ret.Pos())
				if ret.Comment != nil {
					if ret.Comment.End().IsValid() {
						sourceFile.Imports.EndPos = int(ret.Comment.End())
					}
				} else {
					sourceFile.Imports.EndPos = int(ret.Path.Pos()) + len(ret.Path.Value)
				}
				sourceFile.Imports.Text = ret.Path.Value
			}
		}

		return true
	})

	// fmt.Println(sourceFile.Methods)
	return sourceFile
}
