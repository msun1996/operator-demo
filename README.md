# operator-sdk
- 自定义CRD 
- webhook

---

```shell script
# 初始化工程
# dome.com 为crd的域
operator-sdk init --domain=dome.com --license apache2 --owner "dome"
```
```shell script
├─bin # 编译生成文件目录
├─config  # 配置
│  ├─certmanager # 整个目录下的内容用于为 TLS(https)双向认证机制签名生成证书
│  ├─default  # 是对crd，rbac，manager三个目录的整合，及其本身需要的配置
│  ├─manager # 该工程实际的部署实例
│  ├─prometheus  # 该工程部署监控采集
│  ├─rbac  # 默认需要的ClusterRole，ServiceAccount ，绑定关系ClusterRoleBinding，及其具体需要的权限
│  ├─scorecard # 
│  │  ├─bases
│  │  └─patches
│  └─webhook # k8s中需要经过的webhooks拦截器
└─hack
```
---



