package demokubenet

import (
	"log"
	"time"
)

const EasyEvent EventType = "EasyEvent"

type EmulationInstance struct {
	Scheduler  *EventBus
	Satellites []*Satellite
	Stations   []*Station
	Links      []LinkCache
}

func NewEmulationInstanceScale(stationCount, satelliteCount int) (*EmulationInstance, error) {

	var errList error
	stations, err := readStationsCount("data/terminal.txt", stationCount)
	if err != nil {
		log.Fatalf("读取站点失败: %v", err)
	}

	// 读取卫星
	satellites, err := readSatellitesCount("data/satellite_4000.txt", satelliteCount)
	if err != nil {
		log.Fatalf("读取卫星失败: %v", err)
	}
	// satelliteLinks := MakeSatelliteLinks()
	sched := NewEventBus(10)
	instance := &EmulationInstance{
		Scheduler:  sched,
		Satellites: satellites,
		Stations:   stations,
		// SatelliteLinks: satelliteLinks,
	}
	sched.Subscribe(EasyEvent, func(eb *EventBus, event Event) error {
		return instance.EasyCalculateLinks(time.Now())
	})
	return instance, errList
}

func NewEmulationInstance() (*EmulationInstance, error) {
	var errList error
	stations, err := readStations("data/station_data500.txt")
	if err != nil {
		log.Fatalf("读取站点失败: %v", err)
	}

	// 读取卫星
	satellites, err := readSatellites("data/satellite_tle_data2000.txt")
	if err != nil {
		log.Fatalf("读取卫星失败: %v", err)
	}
	// satelliteLinks := MakeSatelliteLinks()
	sched := NewEventBus(10)
	instance := &EmulationInstance{
		Scheduler:  sched,
		Satellites: satellites,
		Stations:   stations,
		// SatelliteLinks: satelliteLinks,
	}
	sched.Subscribe(EasyEvent, func(eb *EventBus, event Event) error {
		return instance.EasyCalculateLinks(time.Now())
	})
	return instance, errList
}

func (e *EmulationInstance) Start() {
	log.Println("emulation_instance.Start")
	// Start the scheduler
	e.Scheduler.Start()
}

// func (e *EmulationInstance) GetSatelliteLinks() []LinkCache {
// 	fmt.Println("Emulation instance: GetSatelliteLinks")
// 	return e.SatelliteLinks
// }

func MakeLinks(stations []*Station, satellites []*Satellite) []LinkCache {
	log.Println("MakeLinks...")
	startTime := time.Now()
	var links []LinkCache
	for _, station := range stations {
		for _, sat := range satellites {
			link := LinkCache{
				SrcNode: sat,
				DstNode: station,
			}
			links = append(links, link)
		}
	}
	log.Printf("links: %d", len(links))
	endTime := time.Now()
	log.Printf("MakeLinks took %v", endTime.Sub(startTime))
	return links
}

func updateSatellitePositions(satellites []*Satellite, timestamp time.Time) {
	log.Println("updateSatellitePositions...")
	startTime := time.Now()
	cnt := 0
	for i := range satellites {
		satellite := satellites[i]
		// satellite.GetPosition(timestamp)
		satellite.Position = satellite.GetPosition(timestamp)
		cnt++
	}
	endTime := time.Now()
	log.Printf("satellite count: %d", cnt)
	log.Printf("updateSatellitePositions took %v", endTime.Sub(startTime))
}

func updateStationPositions(stations []*Station, timestamp time.Time) {
	log.Printf("updateStationPositions...")
	startTime := time.Now()
	for i := range stations {
		station := stations[i]
		station.position = station.GetPosition(timestamp)
	}
	endTime := time.Now()
	log.Printf("updateStationPositions took %v", endTime.Sub(startTime))
}

func getWeatherBasedOnTerminal(terminal *Station, timestamp time.Time) EnvironmentIndex {

	return EnvironmentIndex{Temperature2m: 10.0, Precipitation: 0.0, Pressure: 1010.0} // 模拟返回一些天气数据
}

func updateEnvironmentIndex(stations []*Station, timestamp time.Time) {
	log.Println("updateEnvironmentIndex...")
	startTime := time.Now()
	count := 0
	for i := range stations {
		station := stations[i]
		station.WeatherIdx = getWeatherBasedOnTerminal(station, timestamp)
		count++
	}

	// for i := range links {
	// 	link := &links[i]
	// 	dst, ok := link.DstNode.(*Station)
	// 	if !ok {
	// 		log.Printf("Error: DstNode is not a Station")
	// 		continue
	// 	}
	// 	count++
	// 	EnvironmentIndex := getWeatherBasedOnTerminal(dst, timestamp)
	// 	link.EnvIndex = EnvironmentIndex
	// }
	log.Printf("updateEnvironmentIndex count: %d", count)
	log.Printf("updateEnvironmentIndex took %v", time.Since(startTime))
}

func updateLinkProperties(links []LinkCache) {
	log.Println("updateLinkProperties...")
	startTime := time.Now()
	count := 0
	for i := range links {
		link := &links[i]
		sat, ok := link.SrcNode.(*Satellite)
		if !ok {
			log.Printf("Error: SrcNode is not a Satellite")
			continue
		}
		dst, ok := link.DstNode.(*Station)
		if !ok {
			log.Printf("Error: DstNode is not a Station")
			continue
		}
		srcPos := sat.Position
		dstPos := dst.position
		count++
		link.Ar = CalculateSatelliteLink(link, srcPos, dstPos, dst.WeatherIdx.Precipitation)

	}
	log.Printf("updateLinkProperties count: %d", count)
	log.Printf("updateLinkProperties took %v", time.Since(startTime))
}

func (e *EmulationInstance) EasyCalculateLinks(timestamp time.Time) error {
	// log.Println("instance: EasyCalculateLinks")

	updateSatellitePositions(e.Satellites, timestamp)
	// updateStationPositions(e.Stations, timestamp)
	e.Links = MakeLinks(e.Stations, e.Satellites)
	// updateEnvironmentIndex(e.Links, timestamp)
	updateEnvironmentIndex(e.Stations, timestamp)
	updateLinkProperties(e.Links)

	log.Println("links count: ", len(e.Links))
	// log.Println("Satellite size:", unsafe.Sizeof(Satellite{}))
	// log.Println("Station size:", unsafe.Sizeof(Station{}))
	// log.Println("LinkCache size:", unsafe.Sizeof(LinkCache{}))

	return nil

}
