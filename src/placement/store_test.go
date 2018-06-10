package placement

import (
	"fmt"
	"testing"
)

func Test_parse_stores(t *testing.T) {
	stores := []Store{}
	i := 10001
	//4个DC，每个DC有3个Rack，每个Rack有2个机器
	fmt.Println("----Stores------")
	for d := 5; d >= 1; d-- {
		for r := 1; r <= 3; r++ {
			for h := 1; h <= 2; h++ {
				disk := "hdd"
				//偶数的DC，偶数Rack是SSD
				if r%2 == 0 && d%2 == 0 {
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

	storeInfo := newStoreInfo(stores)
	fmt.Println(storeInfo.DCNumber)
	fmt.Println(storeInfo.RackNumber)
	fmt.Println(storeInfo.StoreNumber)
}
