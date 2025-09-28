# modclean

一行命令，把 Go 模块缓存里 **用不到的旧版本** 全部干掉，磁盘瞬间回春。

modclean – 一键删无用 Go 模块缓存  
## install:  
`go install github.com/wusphinx/modclean@latest`  

## 示例
`go.mod` 如下，示例项目并不依赖 `github.com/gin-gonic/gin`，因此该依赖应删除
```
module demo

go 1.25.0

replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.8.1
```

dry-run
```
modclean
DRY-RUN:  go mod edit -droprequire github.com/gin-gonic/gin  &&  go mod edit -dropreplace github.com/gin-gonic/gin
```

clean
```
modclean -dry-run=false
```
