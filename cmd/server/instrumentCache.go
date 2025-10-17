package main

import (
    "github.com/bhutch29/sclipi/internal/utils"
    "time"
)

func buildAndConnectInstrument(address string, port string, timeout time.Duration, progressFn func(int)) (utils.Instrument, error) {
    var inst utils.Instrument
    if address == "simulated" {
	inst = utils.NewSimInstrument(timeout)
    } else {
	inst = utils.NewScpiInstrument(timeout)
    }

    if err := inst.Connect(address+":"+port, progressFn); err != nil {
	return inst, err
    }

    return inst, nil
}

