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
func incrementMethodData(fileDir, fileName string, sourceFile *SourceFile, s *Service) bool {

	// todo: infopb 需要参数告知
	methodCodeBlock := fmt.Sprintf(`func (d *Data) %[1]sGrpc() *infopb.%[1]sClient {
				return d.%[1]sCli
			}`, s.Service)
	methodKey := fmt.Sprintf("*Data-%sGrpc", s.Service)

	if _, ok := sourceFile.Methods[methodKey]; ok {
		return false
	}
	filePath := path.Join(fileDir, fileName)
	bs, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	// type Data struct {
	if v, ok := sourceFile.Structs["Data"]; ok {
		dataField := fmt.Sprintf(`	%sCli     infopb.%sClient`, s.ServiceLower, s.Service)
		pre := bs[:v.LastFieldEndPos]
		next := append([]byte{}, bs[v.LastFieldEndPos:]...)
		bs = append(pre, []byte(getEndOfLine()+dataField)...)
		bs = append(bs, next...)
	}
	bs = append(bs, []byte(getEndOfLine()+methodCodeBlock)...)
	// test
	filePath = path.Join(fileDir, "data2.go")
	if err := os.WriteFile(filePath, bs, 0o644); err != nil {
		log.Fatal(err)
	}

	return true
}

// 判断是否已经存在相同的服务名称
func parseFile(fileDir string, fileName string) *SourceFile {
	filePath := path.Join(fileDir, fileName)
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	sourceFile := &SourceFile{}
	sourceFile.Methods = make(map[string]MethodInfo, 5)
	sourceFile.Structs = make(map[string]StructInfo, 5)

	ast.Inspect(f, func(n ast.Node) bool {
		// Find Return Statements
		// ret, ok := n.(*ast.ReturnStmt)
		switch ret := n.(type) {
		case *ast.FuncDecl:
			fmt.Printf("return statement found on line %v:\n", fset.Position(ret.Pos()))
			ret.Body = nil
			ret.Doc = nil
			b := bytes.Buffer{}
			printer.Fprint(&b, fset, ret)

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
			return true
		case *ast.StructType:
			// type Data struct {
			// 但是 拿不到 type Data
			{
				fmt.Println(ret)
			}
		case *ast.DeclStmt:
			{
				// 不知
			}
		case *ast.TypeSpec:
			// type Data struct {
			{
				fmt.Println(ret)
				si := StructInfo{}
				si.StartPos = int(ret.Pos())
				// si.LastFieldPos = int(ret.Type)
				if retStructType, ok := ret.Type.(*ast.StructType); ok {

					len := len(retStructType.Fields.List)
					if len > 0 {
						si.LastFieldEndPos = int(retStructType.Fields.List[len-1].End())
					}
				}
				si.EndPos = int(ret.End())
				sourceFile.Structs[ret.Name.Name] = si
			}
		}

		return true
	})

	fmt.Println(sourceFile.Methods)
	return sourceFile
}
