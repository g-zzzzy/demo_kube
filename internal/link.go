package demokubenet

import (
	"bufio"
	"demokubenet/itur"
	"demokubenet/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joshuaferrara/go-satellite"
)

type HourlyData struct {
	Temperature2m   []float64 `json:"temperature_2m"`
	Precipitation   []float64 `json:"precipitation"`
	SurfacePressure []float64 `json:"surface_pressure"`
}

type StationWeather struct {
	Lat    float64
	Lon    float64
	Hourly HourlyData `json:"hourly"`
}

type Link struct {
	Uid int64 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Src int32 `protobuf:"varint,2,opt,name=src,proto3" json:"src,omitempty"`
	Dst int32 `protobuf:"varint,3,opt,name=dst,proto3" json:"dst,omitempty"`
	// Properties     *LinkProperties `protobuf:"bytes,4,opt,name=properties,proto3" json:"properties,omitempty"`
	UniDirectional bool `protobuf:"varint,5,opt,name=uni_directional,json=uniDirectional,proto3" json:"uni_directional,omitempty"` // Not support update this field! Please delete and re-add
	// newly added
	SrcNs string `protobuf:"bytes,6,opt,name=src_ns,json=srcNs,proto3" json:"src_ns,omitempty"`
	DstNs string `protobuf:"bytes,7,opt,name=dst_ns,json=dstNs,proto3" json:"dst_ns,omitempty"`
	// SectionId string      `protobuf:"bytes,8,opt,name=section_id,json=sectionId,proto3" json:"section_id,omitempty"`
	// Status    *LinkStatus `protobuf:"bytes,9,opt,name=status,proto3" json:"status,omitempty"`
}

type LinkCache struct {
	// Raw *Link

	SrcNode Node
	DstNode Node

	EnvIndex EnvironmentIndex

	Ar float64
}

type EnvironmentIndex struct {
	Temperature2m float64
	Precipitation float64
	Pressure      float64 // hPa
}

func readStationsCount(filename string, count int) ([]*Station, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var stations []*Station
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(stations) >= count {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue // 忽略格式不对的行
		}
		long, err1 := strconv.ParseFloat(fields[0], 64)
		lat, err2 := strconv.ParseFloat(fields[1], 64)
		if err1 != nil || err2 != nil {
			continue // 忽略无法解析的行
		}
		alt := 0.0
		if len(fields) >= 3 {
			alt, _ = strconv.ParseFloat(fields[2], 64)
		}
		stations = append(stations, &Station{
			position: Position{
				Latitude:  lat,
				Longitude: long,
				Altitude:  alt,
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(stations) < count {
		return nil, fmt.Errorf("终端数量不足: 期望 %d，实际 %d", count, len(stations))
	}
	return stations, nil
}

func readStations(filename string) ([]*Station, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var stations []*Station
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("station_data.txt 每行至少要包含经度和纬度: %s", line)
		}
		long, err1 := strconv.ParseFloat(fields[0], 64)
		lat, err2 := strconv.ParseFloat(fields[1], 64)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("station_data.txt 格式错误: %s", line)
		}
		alt := 0.0
		if len(fields) >= 3 {
			alt, _ = strconv.ParseFloat(fields[2], 64)
		}
		stations = append(stations, &Station{
			position: Position{
				Latitude:  lat,
				Longitude: long,
				Altitude:  alt,
			},
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return stations, nil
}

func readSatellitesCount(filename string, count int) ([]*Satellite, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sats []*Satellite
	scanner := bufio.NewScanner(file)
	for len(sats) < count {
		if !scanner.Scan() {
			break
		}
		line1 := strings.TrimSpace(scanner.Text())
		if line1 == "" || !strings.HasPrefix(line1, "1 ") {
			continue
		}
		if !scanner.Scan() {
			return nil, fmt.Errorf("TLE文件缺少第二行")
		}
		line2 := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line2, "2 ") {
			return nil, fmt.Errorf("TLE第二行格式错误: %s", line2)
		}
		sat := &Satellite{
			TleLine1:      line1,
			TleLine2:      line2,
			SGP4Satellite: satellite.TLEToSat(line1, line2, satellite.GravityWGS72),
		}
		sats = append(sats, sat)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(sats) < count {
		return nil, fmt.Errorf("TLE数据不足: 期望 %d，实际 %d", count, len(sats))
	}
	return sats, nil
}

func readSatellites(filename string) ([]*Satellite, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sats []*Satellite
	scanner := bufio.NewScanner(file)
	for {
		// TLE是两行
		if !scanner.Scan() {
			break
		}
		line1 := strings.TrimSpace(scanner.Text())
		if line1 == "" || !strings.HasPrefix(line1, "1 ") {
			continue
		}
		if !scanner.Scan() {
			return nil, fmt.Errorf("TLE文件缺少第二行")
		}
		line2 := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line2, "2 ") {
			return nil, fmt.Errorf("TLE第二行格式错误: %s", line2)
		}
		sats = append(sats, &Satellite{
			TleLine1:      line1,
			TleLine2:      line2,
			SGP4Satellite: satellite.TLEToSat(line1, line2, satellite.GravityWGS72),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sats, nil
}

func MakeSatelliteLinks() []LinkCache {
	stations, err := readStations("data/station_data500.txt")
	if err != nil {
		log.Fatalf("读取站点失败: %v", err)
	}

	// 读取卫星
	satellites, err := readSatellites("data/satellite_tle_data2000.txt")
	if err != nil {
		log.Fatalf("读取卫星失败: %v", err)
	}

	// station和satellite两两组合
	var links []LinkCache
	for _, station := range stations {
		for _, sat := range satellites {
			link := LinkCache{
				SrcNode: sat,
				DstNode: station,
				// 这里可以做初始的PenetrationFloors计算
			}
			links = append(links, link)
		}
	}
	log.Printf("links: %d", len(links))
	return links

}

type ForecastResponse struct {
	Hourly struct {
		Temperature2m   []float64 `json:"temperature_2m"`
		Precipitation   []float64 `json:"precipitation"`
		SurfacePressure []float64 `json:"surface_pressure"`
	} `json:"hourly"`
}

type WeatherIndex struct {
	T             float64
	P             float64 // surface pressure	(hPa)
	V_t           float64 // total column water vapour (kg/m2)
	rho           float64 // surface water vapour density (g/m3)
	precipitation float64 // rain (mm/s)
	hr            float64
}

func getWeather(stationPos Position) WeatherIndex {
	//temperature_2m: C
	//precipitation: mm/h
	//surface_pressure: hPa
	log.Printf("getWeather...")
	lat := float32(stationPos.Latitude)
	lon := float32(stationPos.Longitude)
	if lon > 180 {
		lon = lon - 360
	}
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.2f&longitude=%.2f&hourly=temperature_2m,precipitation,surface_pressure", lat, lon)
	// 发送 GET 请求
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	// fmt.Printf("Raw response: %s\n", string(body))
	// 解析 JSON 数据
	var forecast ForecastResponse
	err = json.Unmarshal(body, &forecast)
	// fmt.Printf("parsed forecast: %+v\n", forecast)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}
	currentTime := time.Now()
	hours, _, _ := currentTime.Clock()
	// log.Printf("hours: %d", hours)

	weatherIndex := &WeatherIndex{
		T:             float64(forecast.Hourly.Temperature2m[hours]),
		precipitation: float64(forecast.Hourly.Precipitation[hours]),
		P:             float64(forecast.Hourly.SurfacePressure[hours]),
	}
	return *weatherIndex
}

func loadWeatherData(filePath string) ([]StationWeather, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var data []StationWeather
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func getWeatherFromFile(stationPos Position) WeatherIndex {
	weatherData, err := loadWeatherData("data/weather_data.json")
	if err != nil {
		log.Fatalf("Error loading weather data: %v", err)
	}
	hours := time.Now().Hour()

	for _, wd := range weatherData {
		if wd.Lat == stationPos.Latitude && wd.Lon == stationPos.Longitude {
			return WeatherIndex{
				P:   wd.Hourly.Precipitation[hours],
				T:   wd.Hourly.Temperature2m[hours],
				V_t: wd.Hourly.SurfacePressure[hours],
			}
		}
	}
	return WeatherIndex{}
}

func CalculateSatelliteLink(link *LinkCache, satellitePos Position, stationPos Position, pre float64) float64 {
	//TODO
	//天气数据的获取
	// weatherIndex := getWeather(stationPos)
	// weatherIndex := getWeatherFromFile(stationPos)
	// pre := link.EnvIndex.Precipitation
	latGS, lonGS := stationPos.Latitude, stationPos.Longitude
	el := utils.Elevation_angle(satellitePos.Altitude, satellitePos.Latitude, satellitePos.Longitude, latGS, lonGS)

	f := 22.5 // GHz
	p := 0.1
	hs := 0.1 // km
	R001 := pre
	tau := 45.0
	var Ls float64
	Ar := itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)

	return Ar
}
