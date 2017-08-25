package a320

import (
	"log"
	"time"

	"github.com/apoloval/simavionics"
	"github.com/apoloval/simavionics/a320/apu"
)

const (
	apuPowerOff apuStatus = "apu/status/power_off"
	apuPowerOn  apuStatus = "apu/status/power_on"

	apuActionMasterSwOn simavionics.EventName = "apu/action/master_switch_on"

	apuStateMasterSwOn simavionics.EventName = "apu/state/master_switch_on"

	apuFlapOpenTime = 6 * time.Second
)

type apuStatus string

type apuState struct {
	// Engine parameters
	speed float64
	egt   float64

	// Bleed air pressure
	bleed float64

	// ECB signals
	masterSwitch  bool
	startBtnOn    bool
	startBtnAvail bool
	flapOpen      bool
	fuelLowPre    bool
	lowOilLevel   bool

	// AC generator
	gen GenState
}

type APU struct {
	simavionics.RealTimeSystem

	state  apuState
	status apuStatus

	bus    simavionics.EventBus
	flap   *apu.Flap
	engine *apu.Engine

	apuMasterSwActionChan <-chan simavionics.EventValue
}

func NewAPU(ctx simavionics.Context) *APU {
	apu := &APU{
		RealTimeSystem: simavionics.NewRealTimeSytem(ctx.RealTimeDilation),
		status:         apuPowerOff,
		bus:            ctx.Bus,
		flap:           apu.NewFlap(ctx),
		engine:         apu.NewEngine(ctx),

		apuMasterSwActionChan: ctx.Bus.Subscribe(apuActionMasterSwOn),
	}
	go apu.run()
	return apu
}

func (apu *APU) Start() {
	simavionics.PublishEvent(apu.bus, apuActionMasterSwOn, true)
}

func (apu *APU) run() {
	log.Printf("[apu] Starting a new APU module")
	for {
		select {
		case event := <-apu.apuMasterSwActionChan:
			apu.handleMasterSw(event.Bool())
		case action := <-apu.DeferredActionChan:
			action()
		}
	}
}

func (apu *APU) handleMasterSw(on bool) {
	if on && apu.status == apuPowerOff {
		log.Printf("[apu] Received a master switch action: on -> %v", on)
		apu.status = apuPowerOn
		apu.updateMasterSw(true)
		apu.flap.Open()
		apu.engine.Start()
	}
}

func (apu *APU) updateMasterSw(on bool) {
	apu.updateBool(apuStateMasterSwOn, &apu.state.masterSwitch, on)
}

func (apu *APU) updateBool(en simavionics.EventName, value *bool, update bool) {
	if *value != update {
		*value = update
		simavionics.PublishEvent(apu.bus, en, update)
	}
}
