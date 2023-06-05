# nacosserver

## 如何使用

#### 构建

执行构建脚本 `build.sh` 即可

```bash
# ${nacos_server_plugin_version}: nacosserver 插件版本，默认为 main 的最新 commit
# ${polaris_server_tag}: 北极星服务端版本信息，默认为 main 分支
bash build.sh ${nacos_server_plugin_version} ${polaris_server_tag}
```

#### 配置文件调整

修改 **conf/polaris-server.yaml** 文件，参考下列配置补充

```yaml
bootstrap:
  # Global log
  logger:
    nacos-apiserver:
      rotateOutputPath: log/runtime/nacos-apiserver.log
      errorRotateOutputPath: log/runtime/nacos-apiserver-error.log
      rotationMaxSize: 100
      rotationMaxBackups: 10
      rotationMaxAge: 7
      outputLevel: info
      # outputPaths:
      # - stdout
      # errorOutputPaths:
      # - stderr
apiservers:
  - name: service-nacos
    option:
      listenIP: "0.0.0.0"
      listenPort: 8848
      connLimit:
        openConnLimit: false
        maxConnPerHost: 128
        maxConnLimit: 10240
```


## 其他

- NACOS 中的 struct 数据结构定义大部份引用自 [nacos-sdk-go](https://github.com/nacos-group/nacos-sdk-go)