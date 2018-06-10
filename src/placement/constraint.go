package placement

import "fmt"

//约束的表达式
type Constraint interface {
	Verify(region Region, storeInfo *StoreInfo) (bool, string)
}

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

func (self GeneralConstraint) Verify(region Region, storeInfo *StoreInfo) (bool, string) {
	dcUsage := make(map[string]int)
	rackUsage := make(map[string]int)
	storeUsage := make(map[int]int)
	ssdNumber := 0

	for _, rep := range region.Replicas {
		rack := storeInfo.ID2Rack[rep]
		dc := storeInfo.Rack2Dc[rack]
		ssd := storeInfo.RackSSD[rack]
		if ssd {
			ssdNumber = ssdNumber + 1
		}
		//统计DC的使用
		usage, ok := dcUsage[dc]
		if !ok {
			usage = 0
		}
		dcUsage[dc] = usage + 1

		//统计Rack的使用
		usage2, ok2 := rackUsage[rack]
		if !ok2 {
			usage2 = 0
		}
		rackUsage[rack] = usage2 + 1

		//统计节点的使用
		usage3, ok3 := storeUsage[rep]
		if !ok3 {
			usage = 0
		}
		storeUsage[rep] = usage3 + 1
	}

	dcNumber := len(dcUsage)
	rackNumber := len(rackUsage)
	storeNumber := len(storeUsage)

	if self.dcNumber >= 0 {
		if dcNumber != self.dcNumber {
			return false, fmt.Sprintf("DC number mismatch, expected %d, get %d", self.dcNumber, dcNumber)
		}
	}

	if self.rackNumber >= 0 {
		if rackNumber != self.rackNumber {
			return false, fmt.Sprintf("Rack number mismatch, expected %d, get %d", self.rackNumber, rackNumber)
		}
	}

	if self.storeNumber >= 0 {
		if storeNumber != self.storeNumber {
			return false, fmt.Sprintf("Store number mismatch, expected %d, get %d", self.storeNumber, storeNumber)
		}
	}

	if self.minSSDRepNumber >= 0 {
		if ssdNumber < self.minSSDRepNumber {
			return false, fmt.Sprintf("Min SSD number mismatch, expected >= %d, get %d", self.minSSDRepNumber, ssdNumber)
		}
	}

	if self.maxSSDRepNumber >= 0 {
		if ssdNumber > self.maxSSDRepNumber {
			return false, fmt.Sprintf("Max SSD number mismatch, expected <= %d, get %d", self.maxSSDRepNumber, ssdNumber)
		}
	}

	for _, v := range dcUsage {
		if self.minRepInOneDC >= 0 {
			if v < self.minRepInOneDC {
				return false, fmt.Sprintf("Min reps number in DC mismatch, expected >= %d, get %d", self.minRepInOneDC, v)
			}
		}

		if self.maxRepInOneDC >= 0 {
			if v > self.maxRepInOneDC {
				return false, fmt.Sprintf("Max reps number in DC mismatch, expected <= %d, get %d", self.maxRepInOneDC, v)
			}
		}
	}

	for _, v := range rackUsage {
		if self.minRepInOneRack >= 0 {
			if v < self.minRepInOneRack {
				return false, fmt.Sprintf("Min reps number in Rack mismatch, expected <= %d, get %d", self.minRepInOneRack, v)
			}
		}
		if self.maxRepInOneRack >= 0 {
			if v > self.maxRepInOneRack {
				return false, fmt.Sprintf("Max reps number in Rack mismatch, expected >= %d, get %d", self.maxRepInOneRack, v)
			}
		}
	}

	for _, v := range storeUsage {
		if self.minRepInOneStore >= 0 {
			if v < self.minRepInOneStore {
				return false, fmt.Sprintf("Min reps number in Store mismatch, expected <= %d, get %d", self.minRepInOneStore, v)
			}
		}

		if self.maxRepInOneStore >= 0 {
			if v > self.maxRepInOneStore {
				return false, fmt.Sprintf("Max reps number in Store mismatch, expected >= %d, get %d", self.maxRepInOneStore, v)
			}
		}
	}

	return true, ""
}

//3 副本随机分布在不同的节点
func getRandomAllDifferentIDConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, -1, -1, -1, 3, 1, 1, -1, -1}
}

//节点分布在多个Rack 上,3 副本中的任意 2 副本都不能在同一个 Rack
func getThreeRepWithDifferntRackConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, 3, 1, 1, -1, -1, -1, -1, -1}
}

//3 DC,3 副本分别分布在不同的 DC
func getThreeRepWithDifferntDCConstraint() GeneralConstraint {
	return GeneralConstraint{3, 1, 1, -1, -1, -1, -1, -1, -1, -1, -1}
}

//2 DC,每个 DC 内有多个 Rack,3 副本分布在不同的 Rack 且不能 3 副本都在同一个DC
func getThreeRepWithDifferntRackNotInOneDCConstraint() GeneralConstraint {
	return GeneralConstraint{2, 1, 2, 3, 1, 1, -1, -1, -1, -1, -1}
}

// 2 DC,5 副本,其中一个 DC 放置 3 副本,另外一个 DC 放置 2 副本
func getFiveRepWith2DCSplitConstraint() GeneralConstraint {
	return GeneralConstraint{2, 2, 3, -1, -1, -1, -1, -1, -1, -1, -1}
}

// 3 副本都在SSD磁盘上
func getThreeRepAllInSSDConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, -1, -1, -1, -1, -1, -1, 3, 3}
}

// 3 副本至少有一个副本存储在 ssd 磁盘
func getThreeRepAtLeastOneInSSDConstraint() GeneralConstraint {
	return GeneralConstraint{-1, -1, -1, -1, -1, -1, -1, -1, -1, 1, 3}
}
