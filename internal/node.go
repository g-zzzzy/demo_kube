package demokubenet

import (
	"math"
	"time"

	"github.com/joshuaferrara/go-satellite"
)

const DELTA_TIME = 100 * time.Millisecond

type Node interface {
	GetPosition(timestamp time.Time) Position
	GetVelocity(timestamp time.Time) Velocity
}

type Satellite struct {
	Node
	TleLine1      string
	TleLine2      string
	SGP4Satellite satellite.Satellite
	Position      Position
}

type Station struct {
	Node
	position   Position
	WeatherIdx EnvironmentIndex
}

type Position struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}

type Velocity struct {
	X float64
	Y float64
	Z float64
}

func (s *Station) GetPosition(timestamp time.Time) Position {
	// log.Println("Station GetPosition")
	return s.position
}

func (s *Satellite) GetVelocity(timestamp time.Time) Velocity {
	t1 := time.Now().UTC()
	t2 := t1.Add(DELTA_TIME)

	lla1 := s.GetPosition(t1)
	lla2 := s.GetPosition(t2)

	ecef1 := LLAtoECEF(lla1)
	ecef2 := LLAtoECEF(lla2)

	return Velocity{
		X: (ecef2[0] - ecef1[0]) / DELTA_TIME.Seconds(),
		Y: (ecef2[1] - ecef1[1]) / DELTA_TIME.Seconds(),
		Z: (ecef2[2] - ecef1[2]) / DELTA_TIME.Seconds(),
	}
}

func (s *Satellite) GetPosition(timestamp time.Time) Position {
	// sat := satellite.ParseTLE(s.TleLine1, s.TleLine2, satellite.GravityWGS72)
	// currentTime := time.Now().UTC()
	// year, month, day := currentTime.Date()
	// hour, min, sec := currentTime.Clock()
	position, _ := satellite.Propagate(s.SGP4Satellite, timestamp.Year(), int(timestamp.Month()), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second())
	// 转换到地理坐标
	gmst := satellite.GSTimeFromDate(timestamp.Year(), int(timestamp.Month()), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second())
	alt, _, lla := satellite.ECIToLLA(position, gmst)

	// 单位转换
	latitudeDeg := lla.Latitude * 180 / math.Pi
	longitudeDeg := lla.Longitude * 180 / math.Pi

	longitudeDeg = math.Mod(longitudeDeg+180+360, 360)

	altitudeMeters := alt * 1000

	return Position{
		Latitude:  latitudeDeg,
		Longitude: longitudeDeg,
		Altitude:  altitudeMeters,
	}
}

const (
	A  = 6378137.0        // WGS84 长半轴
	E2 = 6.69437999014e-3 // 第一偏心率的平方
)

func LLAtoECEF(pos Position) [3]float64 {
	latRad := pos.Latitude * math.Pi / 180
	lonRad := pos.Longitude * math.Pi / 180

	N := A / math.Sqrt(1-E2*math.Sin(latRad)*math.Sin(latRad))

	x := (N + pos.Altitude) * math.Cos(latRad) * math.Cos(lonRad)
	y := (N + pos.Altitude) * math.Cos(latRad) * math.Sin(lonRad)
	z := (N*(1-E2) + pos.Altitude) * math.Sin(latRad)

	return [3]float64{x, y, z}
}
