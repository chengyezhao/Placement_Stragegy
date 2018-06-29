# Placement_Stragegy

## <a name="e7noic"></a>需求描述
在 TiDB 集群中,数据是以多个分片(Region)存储在多个 TiKV 节点的,每个 Region 会同
时存储多个副本(副本数通常是 3 或者 5)来达到高可用的目的。PD(Placement Driver)的
一个重要功能就是从全局视角出发控制每个 Region 的多个副本的存放位置。

Store 是存储节点(tikv-server)的抽象,每个 Store 有唯一 ID 和一些键值对形式的标签,用
于标识节点的位置分布及硬件信息。节点的位置通常是树形的拓扑结构,比如集群中的节点分布在不同的 DC(数据中心),不同的 DC 中包含多个 Rack(机架),每个 Rack 中又有多台 Host(主机)。硬件信息可能包含主机的类别(compute/memory/storage),磁盘的类别(hdd/ssd)等。为了后面描述的方便，定义节点属性的字段名称：


| 名称 | 类型 | 描述 |
| --- | --- | --- |
| dc | 字符串 | Store所在的数据中心 |
| rack | 字符串 | Store所在的机架 |
| host | 字符串 | 机器名 |
| compute | 字符串 | 计算机型号 |
| memory | 字符串 | 内存型号 |
| storage | 字符串 | 存储型号 |
| disk | 字符串 | 硬盘型号(hdd/ssd) |

放置策略核心功能是在满足一定策略约束的情况下，在所有的Store中挑选出Region副本的存放位置。举个例子，一个Region有5个副本，一共有10个Store分布在3个DC上，一种策略是保证这个五个副本在分布在两个DC上，那么一种方案就是前面3个副本在DC1，后面2个副本在DC2。

上面仅仅是一个例子，以下是一些常见的策略：
* 3 副本随机分布在不同的节点
* 节点分布在多个 Rack 上,3 副本中的任意 2 副本都不能在同一个 Rack
* 3 DC,3 副本分别分布在不同的 DC
* 2 DC,每个 DC 内有多个 Rack,3 副本分布在不同的 Rack 且不能 3 副本都在同一个
* DC
* 2 DC,5 副本,其中一个 DC 放置 3 副本,另外一个 DC 放置 2 副本
* 3 副本,要求存储在 ssd 磁盘
* 3 副本至少有一个副本存储在 ssd 磁盘
* 以上策略进行组合

放置策略系统需要对外提供的服务包括三个部分：
1. 针对已有的约束，提供若干初始化的满足约束的放置方案
2. 检测已有的一个Region，是否满足某个策略的约束
3. 如果一个已有的Region不满足约束，例如：
    1. 节点变化导致的失效：系统中Store节点变化，导致当前放置方案中某些节点失效，则需要给出一个新的方案
    2. 策略变化导致的失效：Region中的Store节点都还有效，但是不满足当前放置策略，需要重新寻找放置方案
如果产生以上的情况，重新生成一个与当前放置情况尽可能接近的新方案

#### <a name="xrr1ix"></a>对外接口
```go
//stores 是集群中的所有存储节点,region 标识了当前的副本分布情况
//,strategy 是这个 Region 对应的 placement 策略。
func Check(stores []Store, region Region, strategy Strategy) Region {
    ...
}
```

## <a name="61qyih"></a>实现原理
系统实现的包括几个部分
1. 提供一个比较好的框架用来描述不同类型的放置策略，并且允许不同放置策略之间的组合
2. 检测检测某个Region是否符合当前Store节点和放置策略的约束要求。
3. 基于Region的初始放置方案，Store节点的列表，以及放置策略，计算出可用的放置方案

### <a name="u7w1sc"></a>内部数据结构
实现StoreInfo类来保存从[]store数组解析而来的信息
```go
type StoreInfo struct {
	Stores      []Store
	IDs         []int
	IDSet       map[int]bool
	ID2Rack     map[int]string
	DCs         []string
	Racks       []string
	Rack2Dc     map[string]string
	RackSSD     map[string]bool
	StoreNumber int
	DCNumber    int
	RackNumber  int
}
```

### <a name="bw2lxy"></a>策略框架
首先定义策略类如下，包括副本数量的信息以及多个约束
```go
type Strategy struct {
	RepNumber   int
	Constraints []Constraint
}
```

策略可以添加不同的约束条件Constraint，每个约束有Verify方法可以单独来验证放置方案
```go
type Constraint interface {
	Verify(region Region, storeInfo *StoreInfo) bool
}
```
### <a name="9nzhzw"></a>定义策略
我设计了一种简单的基于规则的的约束描述机制，然后用不同的参数组合就可以实现很多规则
```go
//针对一个Region的所有Rep，使用一些规则来定义约束
//所有约束，取负值表示不约束
type GeneralConstraint struct {
	//DC 数量
	dcNumber int
	//每个使用的DC最少几个Rep
	minRepInOneDC int
	//每个使用的DC最多几个Rep
	maxRepInOneDC int
	//Rack数量
	rackNumber int
	//每个使用的Rack内最少几个Rep
	minRepInOneRack int
	//每个使用的Rack内最多几个Rep
	maxRepInOneRack int
	//节点数量
	storeNumber int
	//每个使用的节点最少的Rep数量
	minRepInOneStore int
	//每个使用的节点最多的Rep数量
	maxRepInOneStore int
	//用到的SSD最少数量
	minSSDRepNumber int
	//用到的SSD最多的数量
	maxSSDRepNumber int
}
```
#### <a name="0q2ekf"></a>3 副本随机分布在不同的节点
```go
func getRandomAllDifferentIDConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, -1, -1, -1, 3, 1, 1, -1, -1}
}
```
#### <a name="56apqs"></a>节点分布在多个 Rack 上,3 副本中的任意 2 副本都不能在同一个 Rack
```go
func getThreeRepWithDifferntRackConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, 3, 1, 1, -1, -1, -1, -1, -1}
}
```
#### <a name="hdt8wy"></a>3 DC,3 副本分别分布在不同的 DC
```go
func getThreeRepWithDifferntDCConstraint() GeneralConstraint {
	return GeneralConstraint{3, 1, 1, -1, -1, -1, -1, -1, -1, -1, -1}
}�
```
#### <a name="qlpdgl"></a><span data-type="color" style="color:#262626">2 DC,每个 DC 内有多个 Rack,3 副本分布在不同的 Rack 且不能 3 副本都在同一个DC</span>
```go
func getThreeRepWithDifferntRackNotInOneDCConstraint() GeneralConstraint {
	return GeneralConstraint{2, 1, 2, 3, 1, 1, -1, -1, -1, -1, -1}
}�
```
#### <a name="ev9alo"></a><span data-type="color" style="color:#262626">2 DC,5 副本,其中一个 DC 放置 3 副本,另外一个 DC 放置 2 副本</span>
```go
func getFiveRepWith2DCSplitConstraint() GeneralConstraint {
	return GeneralConstraint{2, 2, 3, -1, -1, -1, -1, -1, -1, -1, -1}
}�
```
#### <a name="av4lol"></a>3 副本,要求存储在 ssd 磁盘
```go
func getThreeRepAllInSSDConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, -1, -1, -1, -1, -1, -1, 3, 3}
}
```
#### <a name="kozogb"></a><span data-type="color" style="color:#262626">3 副本至少有一个副本存储在 ssd 磁盘</span>
```go
func getThreeRepAtLeastOneInSSDConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, -1, -1, -1, -1, -1, -1, 1, 3}
}�
```
#### <a name="eyvrbz"></a>计算放置方案
如果当前方案不满足约束，则需要我们计算出一个与当前方案最接近的有效方案。为了简化问题，我们假设DC和Rack都是同质的，副本在DC之间的移动代价都一样，且远远大于（超过一个数量级）在同一个DC内的Rack之间的移动代价。所以我们尽量使得副本所在的DC不变，先在同一个DC内搜索放置方案；如果实在找不到方案，再考虑跨DC搜索。

首先，我们进行Rack内搜索，选择一个或者多个副本，在当前DC内选择新的Rack，遍历所有的可能组合，直到找到一个可行的方案。这里选择变动副本的数量随着搜索的进行不断的增加。

如果上面的搜索没有找到方案，则在其他DC里面进行搜索。选择一个或者多个副本，在新的DC里面进行搜索，遍历所有可能的组合方案，直到找到可行的方案。这里选择变动副本的数量随着搜索的进行不断增加。

如果假设副本在跨DC之间移动代价略大于（1～5倍）同一个DC内移动，则在有些情况下上述方案不一定找到最优解。

搜索中可以添加了一些特殊处理来剪枝。例如某个搜索节点已经违反了一些必须满足的约束，则可以不再展开往下搜索。不过目前还没有做这方面优化。



# 运行

运行下面的命令会执行编译和测试
./install
