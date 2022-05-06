@REM http 部分

MODULE_NAME=taskReivew
MODULE_NAME_LOW=taskreivew
rm -dr "D:/workspace/drone/drone-appservice/internal/module/$(MODULE_NAME)"
rm "D:/workspace/drone/drone-appservice/internal/domain/$(MODULE_NAME_LOW)_domain.go"
@REM 变量不行


rm -dr "c:/workspace/drone/drone-appservice/internal/module/taskReivew"
rm "c:/workspace/drone/drone-appservice/internal/domain/taskreivew_domain.go"

:: # 使用bash执行

@REM http 部分

@REM MODULE_NAME=opLog
@REM MODULE_NAME_LOW=oplog
set MODULE_NAME=taskReivew
set MODULE_NAME_LOW=taskreivew
rm -dr "c:/workspace/drone/drone-appservice/internal/module/%MODULE_NAME%"
rm "c:/workspace/drone/drone-appservice/internal/domain/%MODULE_NAME_LOW%_domain.go"

:: # 使用bash执行