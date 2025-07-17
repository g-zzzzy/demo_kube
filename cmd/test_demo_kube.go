package main

import (
	internal "demokubenet/internal"
	"log"
	"sync"
	"time"
)

func main() {
	for i := 0; i < 20; i++ {

		startTime := time.Now()
		// 创建EmulationInstance
		inst, err := internal.NewEmulationInstance()
		if err != nil {
			log.Printf("failed to create EmulationInstance: %v", err)
		}

		// 启动scheduler
		inst.Start()

		var wg sync.WaitGroup
		wg.Add(1)

		inst.Scheduler.PublishWithWait(internal.Event{Type: internal.EasyEvent}, &wg)

		wg.Wait()

		// 验证所有LinkCache中的Ar被更新
		// links := inst.GetSatelliteLinks()

		// for i, link := range links {
		// 	fmt.Println("i = ", i, "link.Ar = ", link.Ar)
		// }

		endTime := time.Now()
		log.Println("interval = ", endTime.Sub(startTime))
	}

}
