package placement

const (
	CONST_LABEL_DC      = "dc"
	CONST_LABEL_RACK    = "rack"
	CONST_LABEL_HOST    = "host"
	CONST_LABEL_COMPUTE = "compute"
	CONST_LABEL_MEMORY  = "memory"
	CONST_LABEL_STORAGE = "storage"
	CONST_LABEL_DISK    = "disk" //ssd, hdd
)

type Store struct {
	// store唯一id
	ID int
	// store的各种属性
	Labels map[string]string
}

type IDList []int

//处理后的store出具，方便其他类使用
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

func newStoreInfo(stores []Store) *StoreInfo {
	DCs := []string{}
	Racks := []string{}
	IDs := make([]int, len(stores))
	IDSet := make(map[int]bool)

	DCSet := make(map[string]bool)
	RackSet := make(map[string]bool)
	//Rack是否有SSD
	RackSSDMap := make(map[string]bool)
	//Rack所属于的DC
	Racks2DCMap := make(map[string]string)
	//ID属于Rack
	ID2RackMap := make(map[int]string)

	//分析store所有的标签
	for i, store := range stores {
		IDs[i] = store.ID
		IDSet[IDs[i]] = true
		var dc string
		var rack string
		ssd := false
		for k, v := range store.Labels {
			switch k {
			case CONST_LABEL_DC:
				_, ok := DCSet[v]
				if !ok {
					DCs = append(DCs, v)
				}
				dc = v
				DCSet[v] = true
			case CONST_LABEL_RACK:
				_, ok := RackSet[v]
				if !ok {
					Racks = append(Racks, v)
				}
				RackSet[v] = true
				rack = v
			case CONST_LABEL_DISK:
				if v == "ssd" {
					ssd = true
				}
			default:
			}
		}
		ID2RackMap[IDs[i]] = rack
		RackSSDMap[rack] = ssd
		Racks2DCMap[rack] = dc
	}
	// fmt.Println("----Racks2DCMap------")
	// fmt.Println(Racks2DCMap)
	// fmt.Println("----RackSSDMap------")
	// fmt.Println(RackSSDMap)
	// fmt.Println("----IDs------")
	// fmt.Println(IDs)

	return &StoreInfo{stores, IDs, IDSet, ID2RackMap, DCs, Racks, Racks2DCMap, RackSSDMap,
		len(stores), len(DCs), len(Racks)}
}
