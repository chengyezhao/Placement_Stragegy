package placement

import "fmt"
import "reflect"

//策略接口
type Strategy struct {
	RepNumber   int
	Constraints []Constraint
}

func newStrategy(repNumber int, constraints []Constraint) *Strategy {
	return &Strategy{repNumber, constraints}
}

func (self *Strategy) Check(region Region, storeInfo *StoreInfo) (bool, string) {
	flag := true
	error := ""
	//检查是否有节点变化
	for _, r := range region.Replicas {
		_, ok := storeInfo.IDSet[r]
		if !ok {
			//某个节点没有找到
			error = fmt.Sprintf("failed on unknown store %d", r)
			return false, error
		}
	}

	//查看策略的副本数量和region是否一致
	if self.RepNumber >= 0 {
		flag = (self.RepNumber == len(region.Replicas))
		if !flag {
			error = fmt.Sprintf("Rep number mismatch, expected %d, get %d", self.RepNumber, len(region.Replicas))
		}
	}

	//对所有约束进行检查
	for _, c := range self.Constraints {
		flag, error := c.Verify(region, storeInfo)
		if flag == false {
			error = fmt.Sprintf("failed on %s with reason : \n \t\t %s", reflect.TypeOf(c), error)
			return false, error
		}
	}

	return flag, error
}
