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
	for i := range satellites {
		satellite := satellites[i]
		// satellite.GetPosition(timestamp)
		satellite.Position = satellite.GetPosition(timestamp)
	}
	endTime := time.Now()
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

func (e *EmulationInstance) EasyCalculateLinks(timestamp time.Time) error {
	// log.Println("instance: EasyCalculateLinks")
	satellites := e.Satellites
	stations := e.Stations
	updateSatellitePositions(satellites, timestamp)
	updateStationPositions(stations, timestamp)
	links := MakeLinks(stations, satellites)
	startTime := time.Now()
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
		_ = sat.Position
		_ = dst.position

		// Ar := CalculateSatelliteLink(link, srcPos, dstPos)
		link.Ar = 0
	}
	log.Println("links count: ", len(links))
	log.Println("LinkCal took: ", time.Since(startTime))
	// log.Println("Satellite size:", unsafe.Sizeof(Satellite{}))
	// log.Println("Station size:", unsafe.Sizeof(Station{}))
	// log.Println("LinkCache size:", unsafe.Sizeof(LinkCache{}))

	return nil

}
