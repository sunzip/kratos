切到分支 
1. gti eheckout createServiceByTemplate
2. go install .

会生成 mykratos.exe ( windows 为例 ) 文件, 目录$GOPATH/bin

生成错误码翻译文件, 

使用:
 mykratos proto client api/mozi/common/v1/errors.proto 


> 本仓库调试, 会报:  
> D:\gopath\src\github.com\go-kratos\kratos: warning: directory does not exist.  
> D:\gopath\src\github.com\go-kratos\kratos\third_party: warning: directory does not exist.
> 不要担心