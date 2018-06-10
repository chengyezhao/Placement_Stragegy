package placement

import "fmt"

func ContraintSolve(storeInfo *StoreInfo, constraint []Constraint, region Region, strategy *Strategy) Region {
	flag1, newRegion1 := SearchInDC(storeInfo, constraint, region, strategy)
	if !flag1 {
		flag2, newRegion2 := SearchAcrossDC(storeInfo, constraint, region, strategy, true)
		if flag2 {
			fmt.Println(fmt.Sprintf("SearchAcrossDC success %v", newRegion2))
			return newRegion2
		}
		fmt.Println("No new region found")
		return region
	} else {
		fmt.Println(fmt.Sprintf("SearchInDC success %v", newRegion1))
		return newRegion1
	}
}

func SearchRegions(storeInfo *StoreInfo, reps []int, IDsInDC map[string]IDList, acrossDC bool, constraint []Constraint, region Region, strategy *Strategy) (bool, Region) {
	stack := [][]int{}
	fmt.Println(reps)
	IDs := storeInfo.IDs
	if !acrossDC {
		rep := storeInfo.IDs[reps[0]]
		rack := storeInfo.ID2Rack[rep]
		dc := storeInfo.Rack2Dc[rack]
		IDs = IDsInDC[dc]
	}
	for _, v := range IDs {
		stack = append(stack, []int{v})
		for i := 0; i < len(stack); i++ {
			top := stack[i]
			// fmt.Println(top)
			newReplicas := make([]int, len(region.Replicas))
			copy(newReplicas, region.Replicas)
			for j := 0; j < len(top); j++ {
				newReplicas[reps[j]] = top[j]
			}
			flag, _ := strategy.Check(Region{newReplicas}, storeInfo)
			if flag {
				return true, Region{newReplicas}
			}
		}
	}
	for i := 1; i < len(reps); i++ {
		IDs := storeInfo.IDs
		if !acrossDC {
			rep := storeInfo.IDs[reps[i]]
			rack := storeInfo.ID2Rack[rep]
			dc := storeInfo.Rack2Dc[rack]
			IDs = IDsInDC[dc]
		}
		stackLen := len(stack)
		newStack := [][]int{}
		for _, v := range IDs {
			for j := 0; j < stackLen; j++ {
				comb := stack[j]
				newComb := make([]int, len(comb))
				copy(newComb, comb)
				newComb = append(newComb, v)
				newStack = append(newStack, newComb)
			}
		}
		stack = newStack
		for i := 0; i < len(stack); i++ {
			top := stack[i]
			// fmt.Println(top)
			newReplicas := make([]int, len(region.Replicas))
			copy(newReplicas, region.Replicas)
			for j := 0; j < len(top); j++ {
				newReplicas[reps[j]] = top[j]
			}
			flag, _ := strategy.Check(Region{newReplicas}, storeInfo)
			if flag {
				return true, Region{newReplicas}
			}
		}
	}

	return false, region
}

func SearchAcrossDC(storeInfo *StoreInfo, constraints []Constraint, region Region, strategy *Strategy, acrossDC bool) (bool, Region) {
	//计算DC内的Rep
	IDsInDC := make(map[string]IDList)
	if !acrossDC {
		for id, rack := range storeInfo.ID2Rack {
			dc := storeInfo.Rack2Dc[rack]
			_, ok := IDsInDC[dc]
			if !ok {
				IDsInDC[dc] = []int{}
			}
			IDsInDC[dc] = append(IDsInDC[dc], id)
		}
	}

	storeNumber := len(region.Replicas)
	//遍历哪些rep需要改动
	stack := [][]int{}
	for i := 0; i < storeNumber; i++ {
		stack = append(stack, []int{i})
	}
	fmt.Println(fmt.Sprintf("stack: %v, pointer : %d", stack, 0))

	stackLen := storeNumber
	for searchPoint := 0; searchPoint < stackLen; searchPoint++ {
		top := stack[searchPoint]
		//处理栈顶
		flag, newRegion := SearchRegions(storeInfo, top, IDsInDC, acrossDC, constraints, region, strategy)
		if flag {
			return true, newRegion
		}
		//插入新元素
		topLen := len(top)
		for i := top[topLen-1] + 1; i < storeNumber; i++ {
			newTop := make([]int, topLen)
			copy(newTop, top)
			newTop = append(newTop, i)
			stack = append(stack, newTop)
			stackLen = stackLen + 1
			fmt.Println(fmt.Sprintf("add %v to stack, len %d", newTop, stackLen))
		}
		fmt.Println(fmt.Sprintf("stack pointer : %d", searchPoint))
	}
	return false, region
}

func SearchInDC(storeInfo *StoreInfo, constraint []Constraint, region Region, strategy *Strategy) (bool, Region) {
	return SearchAcrossDC(storeInfo, constraint, region, strategy, false)
}

func Check(stores []Store, region Region, strategy *Strategy) Region {
	//获取存储的信息
	storeInfo := newStoreInfo(stores)
	fmt.Println(fmt.Sprintf("==================checking region %v===================", region.Replicas))
	needReCompute := false

	if !needReCompute {
		//检查策略是否通过
		flag, error := strategy.Check(region, storeInfo)
		if !flag {
			fmt.Println(error)
			needReCompute = true
		}
	}

	//如果需要重新计算
	if needReCompute {
		fmt.Println(fmt.Sprintf("recalulating region"))
		region = ContraintSolve(storeInfo, strategy.Constraints, region, strategy)
		return region
	} else {
		fmt.Println(fmt.Sprintf("region is ok"))
		return region
	}
}
