package service

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/common"
)

type ResolutionChangeMessage struct {
	Type   string `json:"type"`
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}

type ResolutionMonitor struct {
	mutex           sync.RWMutex
	lastWidth       uint16
	lastHeight      uint16
	isMonitoring    bool
	stopChan        chan bool
	wsClients       map[*websocket.Conn]bool
	wsClientsMutex  sync.RWMutex
}

var (
	resolutionMonitor *ResolutionMonitor
	resolutionOnce    sync.Once
)

func GetResolutionMonitor() *ResolutionMonitor {
	resolutionOnce.Do(func() {
		resolutionMonitor = &ResolutionMonitor{
			wsClients: make(map[*websocket.Conn]bool),
			stopChan:  make(chan bool),
		}
	})
	return resolutionMonitor
}

func (rm *ResolutionMonitor) AddClient(conn *websocket.Conn) {
	rm.wsClientsMutex.Lock()
	defer rm.wsClientsMutex.Unlock()
	
	rm.wsClients[conn] = true
	
	// Start monitoring if this is the first client
	if len(rm.wsClients) == 1 && !rm.isMonitoring {
		go rm.StartMonitoring()
	}
}

func (rm *ResolutionMonitor) RemoveClient(conn *websocket.Conn) {
	rm.wsClientsMutex.Lock()
	defer rm.wsClientsMutex.Unlock()
	
	delete(rm.wsClients, conn)
	
	// Stop monitoring if no clients are connected
	if len(rm.wsClients) == 0 && rm.isMonitoring {
		rm.StopMonitoring()
	}
}

func (rm *ResolutionMonitor) StartMonitoring() {
	rm.mutex.Lock()
	if rm.isMonitoring {
		rm.mutex.Unlock()
		return
	}
	rm.isMonitoring = true
	rm.mutex.Unlock()

	log.Debug("Starting resolution monitoring")

	// Initialize with current resolution
	width, height := rm.getCurrentResolution()
	rm.lastWidth = width
	rm.lastHeight = height

	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-rm.stopChan:
			rm.mutex.Lock()
			rm.isMonitoring = false
			rm.mutex.Unlock()
			log.Debug("Stopped resolution monitoring")
			return
		case <-ticker.C:
			rm.checkResolutionChange()
		}
	}
}

func (rm *ResolutionMonitor) StopMonitoring() {
	rm.mutex.RLock()
	isMonitoring := rm.isMonitoring
	rm.mutex.RUnlock()

	if !isMonitoring {
		return
	}

	select {
	case rm.stopChan <- true:
	default:
	}
}

func (rm *ResolutionMonitor) getCurrentResolution() (uint16, uint16) {
	// Read from the same files that the C++ code writes to
	widthFile := "/kvmapp/kvm/width"
	heightFile := "/kvmapp/kvm/height"

	width := uint16(1920) // Default
	height := uint16(1080) // Default

	if data, err := os.ReadFile(widthFile); err == nil {
		if w, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && w > 0 {
			width = uint16(w)
		}
	}

	if data, err := os.ReadFile(heightFile); err == nil {
		if h, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && h > 0 {
			height = uint16(h)
		}
	}

	return width, height
}

func (rm *ResolutionMonitor) checkResolutionChange() {
	currentWidth, currentHeight := rm.getCurrentResolution()

	rm.mutex.RLock()
	lastWidth := rm.lastWidth
	lastHeight := rm.lastHeight
	rm.mutex.RUnlock()

	// Check if resolution has changed
	if currentWidth != lastWidth || currentHeight != lastHeight {
		log.Debugf("Resolution changed from %dx%d to %dx%d", lastWidth, lastHeight, currentWidth, currentHeight)

		rm.mutex.Lock()
		rm.lastWidth = currentWidth
		rm.lastHeight = currentHeight
		rm.mutex.Unlock()

		// For auto mode (0x0), calculate optimal resolution
		if screen := common.GetScreen(); screen.Width == 0 && screen.Height == 0 {
			optimalWidth, optimalHeight := common.GetOptimalResolution(currentWidth, currentHeight)
			log.Debugf("Auto mode detected, using optimal resolution: %dx%d", optimalWidth, optimalHeight)
			
			// Update screen configuration
			screen.Width = optimalWidth
			screen.Height = optimalHeight
			
			// Notify clients with the optimal resolution instead of the raw resolution
			rm.notifyClients(optimalWidth, optimalHeight)
		} else {
			// Notify clients with the actual detected resolution
			rm.notifyClients(currentWidth, currentHeight)
		}
	}
}

func (rm *ResolutionMonitor) notifyClients(width, height uint16) {
	rm.wsClientsMutex.RLock()
	clients := make([]*websocket.Conn, 0, len(rm.wsClients))
	for client := range rm.wsClients {
		clients = append(clients, client)
	}
	rm.wsClientsMutex.RUnlock()

	if len(clients) == 0 {
		return
	}

	message := ResolutionChangeMessage{
		Type:   "resolution_change",
		Width:  width,
		Height: height,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Errorf("Failed to marshal resolution change message: %v", err)
		return
	}

	for _, conn := range clients {
		if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Debugf("Failed to send resolution change to client: %v", err)
		}
	}
}