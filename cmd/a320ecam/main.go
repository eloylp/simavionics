package main

import (
	"github.com/apoloval/simavionics"
	"github.com/apoloval/simavionics/event/remote"
	"github.com/apoloval/simavionics/ui"
	"github.com/op/go-logging"
	"github.com/veandco/go-sdl2/sdl"
)

type page interface {
	processEvents()
	render(display *ui.Display)
}

var log = logging.MustGetLogger("ecam")

func main() {
	simavionics.EnableLogging()

	var err error
	cfg, err := loadConfig()
	if err != nil {
		log.Panic(err)
	}

	log.Info("Initializing display")
	display, err := ui.NewDisplay("SimAvionics A320 Lower ECAM", cfg.width, cfg.height)
	if err != nil {
		log.Panic(err)
	}

	log.Info("Initializing SimAvionics remote bus")
	bus, err := remote.NewSlaveEventBus("tcp://localhost:7001")
	if err != nil {
		log.Panic(err)
	}

	apuPage, err := newAPUPage(bus, display)
	if err != nil {
		log.Panic(nil)
	}
	disconnPage := newDisconnectionPage(bus)

	for {
		disconnPage.processEvents()
		apuPage.processEvents()

		if disconnPage.isDisconnected {
			disconnPage.render(display)
		} else {
			apuPage.render(display)
		}

		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}
		}
	}
}
