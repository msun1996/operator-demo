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
```shell script
#自定义CRD
operator-sdk create api --group=paas --version=v1beta1 --kind=Instance
```
```shell script
├─api
│  └─v1beta1  # 定义operator crd 的字段信息
├─config
│  ├─crd  # 自定义operator的yaml
│  │  └─patches
│  ├─rbac # 更新加入 rbac 新的crd 的相关权限
│  ├─samples # 部署对应crd实例yaml
├─controllers  # 定义apply实例后，进入程序，程序代码操作
└─hack
```
# 安装CRD
```shell script
make install
---
/root/go/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/root/go/bin/kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/instances.paas.dome.com created
---
```
#### 执行controller部分代码，相当于k8s schema部分代码
```shell script
make run # 使用本地的.kube/conf 文件，不需要绑定角色权限等
---
/root/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
/root/go/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
go run ./main.go
---
```
##### 自动生成 deepcopy 相关代码
`更新zz_generated.deepcopy.go文件,在修改api里面对象字段后执行`
```
make generate
/root/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
```
##### 生成新的crd
```
make manifests
/root/go/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```
#### 创建instance实例
```shell script
kubectl apply -f config/samples/paas_v1beta1_instance.yaml 
instance.paas.dome.com/instance-sample created
kubectl get instance
NAME              AGE
instance-sample   12s
```

### instance controller节点代码主要分析
#### 触发条件
```go
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&paasv1beta1.Instance{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
//For() 主要监视里面定义资源的增删改
//Owns() 辅助监视里面定义资源的增删改
```
# 触发执行方法
```go
func (r *InstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	instance := &paasv1beta1.Instance{}
	err := r.Get(ctx, req.NamespacedName, instance)
}
// req为触发该方法实例的信息，仅有 NamespacedName信息
```

### 镜像运行
```shell script
# 打包运行，打包方法是修改Makefile部分内容生成镜像包和yaml文件(打包部分内容需要外网环境)
make package IMG="harbor.common.com:9443/library/paas:0.1"
kubectl apply -f deploy/deploy.yaml
```

### webhook
#### 创建webhook
```shell script
operator-sdk create webhook --group=paas --version=v1beta1 --kind=Instance --defaulting --programmatic-validation
```
#### 部署cert-manager(`https://book.kubebuilder.io/cronjob-tutorial/cert-manager.html`)
```shell script
# https://book.kubebuilder.io/cronjob-tutorial/cert-manager.html
```
```shell script
# config/default/kustomization.yaml 文件关于webhook取消注解
```
#### 部署
```shell script
make docker-build docker-push IMG=harbor.common.com:9443/library/paas:0.1
make deploy IMG=harbor.common.com:9443/library/paas:0.1
```

