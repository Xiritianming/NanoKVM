package common

import "sync"

type Screen struct {
	Width   uint16
	Height  uint16
	FPS     int
	Quality uint16
	BitRate uint16
	GOP     uint8
}

var (
	screen     *Screen
	screenOnce sync.Once
)

// ResolutionMap height to width
var ResolutionMap = map[uint16]uint16{
	1080: 1920, // 16:9
	960:  1280, // 4:3
	900:  1600, // 16:9
	864:  1152, // 4:3
	800:  1280, // 16:10
	720:  1280, // 16:9
	768:  1024, // 4:3
	600:  800,  // 4:3
	480:  640,  // 4:3
	0:    0,    // Auto
}

var QualityMap = map[uint16]bool{
	100: true,
	80:  true,
	60:  true,
	50:  true,
}

var BitRateMap = map[uint16]bool{
	5000: true,
	3000: true,
	2000: true,
	1000: true,
}

func GetScreen() *Screen {
	screenOnce.Do(func() {
		screen = &Screen{
			Width:   0,
			Height:  0,
			Quality: 80,
			FPS:     30,
			BitRate: 3000,
			GOP:     30,
		}
	})

	return screen
}

func SetScreen(key string, value int) {
	switch key {
	case "resolution":
		height := uint16(value)
		if width, ok := ResolutionMap[height]; ok {
			screen.Width = width
			screen.Height = height
		}

	case "quality":
		if value > 100 {
			screen.BitRate = uint16(value)
		} else {
			screen.Quality = uint16(value)
		}

	case "fps":
		screen.FPS = validateFPS(value)

	case "gop":
		screen.GOP = uint8(value)
	}
}

func CheckScreen() {
	if _, ok := ResolutionMap[screen.Height]; !ok {
		screen.Width = 1920
		screen.Height = 1080
	}

	if _, ok := QualityMap[screen.Quality]; !ok {
		screen.Quality = 80
	}

	if _, ok := BitRateMap[screen.BitRate]; !ok {
		screen.BitRate = 3000
	}
}

// GetOptimalResolution returns the best matching resolution for a given aspect ratio
func GetOptimalResolution(inputWidth, inputHeight uint16) (uint16, uint16) {
	if inputWidth == 0 || inputHeight == 0 {
		return 1920, 1080 // Default 16:9
	}
	
	inputRatio := float64(inputWidth) / float64(inputHeight)
	
	// Define common aspect ratios with their priorities
	aspectRatios := []struct {
		ratio float64
		width uint16
		height uint16
		priority int
	}{
		{16.0/9.0, 1920, 1080, 1},   // 16:9 - highest priority
		{16.0/9.0, 1600, 900, 2},    // 16:9
		{16.0/9.0, 1280, 720, 3},    // 16:9
		{4.0/3.0, 1280, 960, 1},     // 4:3 - high priority
		{4.0/3.0, 1024, 768, 2},     // 4:3
		{4.0/3.0, 800, 600, 3},      // 4:3
		{4.0/3.0, 640, 480, 4},      // 4:3
		{16.0/10.0, 1280, 800, 2},   // 16:10
		{3.0/2.0, 1152, 864, 3},     // 3:2
	}
	
	bestMatch := aspectRatios[0]
	minDiff := 999.0
	
	for _, ar := range aspectRatios {
		diff := abs(inputRatio - ar.ratio)
		if diff < minDiff || (diff == minDiff && ar.priority < bestMatch.priority) {
			minDiff = diff
			bestMatch = ar
		}
	}
	
	return bestMatch.width, bestMatch.height
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func validateFPS(fps int) int {
	if fps > 60 {
		return 60
	}
	if fps < 10 {
		return 10
	}

	return fps
}
