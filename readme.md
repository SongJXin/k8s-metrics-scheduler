# 根据资源实际使用情况调度pod

## 背景

1. 默认k8s调度pod时,是根据资源请求(resources的requests)来调度,在调度的时候,会根据每个node上的所有的pod的requests总和对node进行打分,以确定该pod最终部署到哪台机器上
2. 但是,有的用户在部署pod的时候,不写资源限制,或者不按照项目实际运行所需的资源写配置,就会导致某些node的资源requests总和不高，但是实际已经使用了很多了.调度上去的pod无法正常运行。
3. 所以,需要一个调度器,根据资源实际使用情况调度pod,以保证每个node上的资源使用率。

## 实现

计划: 修改打分策略,从metrics-server处获取node的资源使用情况,然后根据资源使用情况打分。  
理论： 调库框架向现有的调度器中添加了一组插件化的 API，该 API 在保持调度程序“核心”简单且易于维护的同时，使得大部分的调度功能以插件的形式存在
参考文档:
1. https://blog.csdn.net/fly910905/article/details/124000222
2. https://developer.aliyun.com/article/766998
3. https://zhuanlan.zhihu.com/p/113620537

## 编码

### 首先解决的是k8s.io/kubernetes的以来问题

1. 使用git 下载kubernetes源码,签出对应版本的分支或tag。
2. go mod中添加replace，(replace内容可以从 kubernetes源码的go.mod中复制出来，只复制替换路径为本地)
3. 开发
4. 打包
5. 部署

## 部署注意事项

1. 新部署的插件，不能加入kube-scheduler的选举中。
2. 插件要有足够的权限获取metrics-server的数据。

## 仍然存在的问题

1. 只能根据CPU和内存调度
2. 所有关于 尽量(Prefer)有关的调度策略，都无法使用。比如，尽量避免调度到该节点(taint:PreferNoschedule),pod尽量避免调度到同一台机器等。