package main

import (
	internal "demokubenet/internal"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <station_num> <satellite_nums> <round>", os.Args[0])
	}
	// 解析参数
	stationCount, err1 := strconv.Atoi(os.Args[1])
	satelliteCount, err2 := strconv.Atoi(os.Args[2])
	round, err3 := strconv.Atoi(os.Args[3])
	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatalf("Invalid arguments: %v, %v, %v", err1, err2, err3)
	}
	fmt.Printf("KubeDemo running with %d stations and %d satellites for %d rounds\n", stationCount, satelliteCount, round)

	// for i := 0; i < 20; i++ {

	startTime := time.Now()
	// 创建EmulationInstance
	inst, err := internal.NewEmulationInstanceScale(stationCount, satelliteCount)
	// inst, err := internal.NewEmulationInstance()
	if err != nil {
		log.Printf("failed to create EmulationInstance: %v", err)
	}

	// 启动scheduler
	inst.Start()

	for i := 0; i < round; i++ {
		var wg sync.WaitGroup
		wg.Add(1)

		inst.Scheduler.PublishWithWait(internal.Event{Type: internal.EasyEvent}, &wg)

		wg.Wait()
		log.Printf("Run %d complete\n", i+1)
	}

	// 验证所有LinkCache中的Ar被更新
	// links := inst.GetSatelliteLinks()

	// for i, link := range links {
	// 	fmt.Println("i = ", i, "link.Ar = ", link.Ar)
	// }

	endTime := time.Now()
	log.Println("interval = ", endTime.Sub(startTime))
	// }

}
