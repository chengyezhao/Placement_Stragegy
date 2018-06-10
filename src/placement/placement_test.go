package placement

import (
	"fmt"
	"testing"
)

func Test_placement_3_reps_RandomAllDifferentIDConstraint(t *testing.T) {

	stores := []Store{}
	i := 10001
	//10个DC，每个DC有3个Rack，每个Rack有2个机器
	fmt.Println("----Stores------")
	for d := 1; d <= 5; d++ {
		for r := 1; r <= 3; r++ {
			for h := 1; h <= 2; h++ {
				disk := "hdd"
				//偶数的DC，偶数Rack是SSD
				if r%2 == 0 && d%2 == 1 {
					disk = "ssd"
				}

				labels := map[string]string{"dc": fmt.Sprintf("DC%d", d),
					"rack": fmt.Sprintf("DC%d-Rack%d", d, r),
					"host": fmt.Sprintf("DC%d-Rack%d-Host%d", d, r, h),
					"disk": disk}
				store := Store{i, labels}
				stores = append(stores, store)

				fmt.Print(store.ID)
				fmt.Print("===>")
				fmt.Println(store.Labels)
				i = i + 1
			}
		}
	}

	cons1 := getRandomAllDifferentIDConstraint()
	cons2 := getThreeRepWithDifferntRackConstraint()
	cons3 := getThreeRepWithDifferntDCConstraint()
	cons4 := getThreeRepWithDifferntRackNotInOneDCConstraint()
	cons5 := getFiveRepWith2DCSplitConstraint()
	cons6 := getThreeRepAllInSSDConstraint()
	cons7 := getThreeRepAtLeastOneInSSDConstraint()

	//run strategy1
	strategy1 := newStrategy(3, []Constraint{cons1})
	region1 := Region{[]int{10001, 10002, 10100}}
	Check(stores, region1, strategy1)

	//run strategy1
	strategy2 := newStrategy(3, []Constraint{cons2})
	region2 := Region{[]int{10001, 10003, 10001}}
	Check(stores, region2, strategy2)

	// //run strategy3
	strategy3 := newStrategy(3, []Constraint{cons2, cons6})
	region3 := Region{[]int{10001, 10003, 10005}}
	Check(stores, region3, strategy3)

	// //run strategy4
	strategy4 := newStrategy(3, []Constraint{cons2, cons3})
	region4 := Region{[]int{10001, 10007, 10012}}
	Check(stores, region4, strategy4)

	// //run strategy5
	strategy5 := newStrategy(3, []Constraint{cons2, cons4})
	region5 := Region{[]int{10001, 10007, 10014}}
	Check(stores, region5, strategy5)

	// //run strategy6
	strategy6 := newStrategy(5, []Constraint{cons5})
	region6 := Region{[]int{10001, 10006, 10007, 10009, 10015}}
	Check(stores, region6, strategy6)

	// //run strategy7
	strategy7 := newStrategy(3, []Constraint{cons7})
	region7 := Region{[]int{10001, 10006, 10007}}
	Check(stores, region7, strategy7)

	fmt.Println("end")
}
