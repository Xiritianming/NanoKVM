package vm

import (
	"NanoKVM-Server/common"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

var screenFileMap = map[string]string{
	"type":       "/kvmapp/kvm/type",
	"fps":        "/kvmapp/kvm/fps",
	"quality":    "/kvmapp/kvm/qlty",
	"resolution": "/kvmapp/kvm/res",
}

type ScreenInfo struct {
	ConfiguredWidth  uint16 `json:"configuredWidth"`
	ConfiguredHeight uint16 `json:"configuredHeight"`
	ActualWidth      uint16 `json:"actualWidth"`
	ActualHeight     uint16 `json:"actualHeight"`
	IsAutoMode       bool   `json:"isAutoMode"`
}

func (s *Service) GetScreen(c *gin.Context) {
	var rsp proto.Response

	screen := common.GetScreen()
	
	// Get actual resolution from files
	actualWidth, actualHeight := getCurrentActualResolution()
	
	screenInfo := ScreenInfo{
		ConfiguredWidth:  screen.Width,
		ConfiguredHeight: screen.Height,
		ActualWidth:      actualWidth,
		ActualHeight:     actualHeight,
		IsAutoMode:       screen.Width == 0 && screen.Height == 0,
	}

	rsp.OkRspWithData(c, screenInfo)
}

func getCurrentActualResolution() (uint16, uint16) {
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

func (s *Service) SetScreen(c *gin.Context) {
	var req proto.SetScreenReq
	var rsp proto.Response

	err := proto.ParseFormRequest(c, &req)
	if err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	switch req.Type {
	case "type":
		data := "h264"
		if req.Value == 0 {
			data = "mjpeg"
		}
		err = writeScreen("type", data)

	case "gop":
		gop := 30
		if req.Value >= 1 && req.Value <= 100 {
			gop = req.Value
		}
		common.GetKvmVision().SetGop(uint8(gop))

	default:
		data := strconv.Itoa(req.Value)
		err = writeScreen(req.Type, data)
	}

	if err != nil {
		rsp.ErrRsp(c, -2, "update screen failed")
		return
	}

	common.SetScreen(req.Type, req.Value)

	log.Debugf("update screen: %+v", req)
	rsp.OkRsp(c)
}

func writeScreen(key string, value string) error {
	file, ok := screenFileMap[key]
	if !ok {
		return fmt.Errorf("invalid argument %s", key)
	}

	err := os.WriteFile(file, []byte(value), 0o666)
	if err != nil {
		log.Errorf("write kvm %s failed: %s", file, err)
		return err
	}

	return nil
}
