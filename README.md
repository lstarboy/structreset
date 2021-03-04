# structreset

### 下载安装

```
git clone git@git.kgidc.cn:beckjiang/structreset.git

cd structreset

./install.sh
```

### 检测代码 

```
cd $project
structresetx path/to/package
```
比如：
```
cd ~/go/src/kugou.net/kmr_entity_api

structresetx ./postprocessor/v2album ./postprocessor/v2audio ./postprocessor/v2author ./postprocessor/v2video ./postprocessor/v2work
// or
structresetx -d ./postprocessor
```

#### 其他说明
```
structresetx help
```