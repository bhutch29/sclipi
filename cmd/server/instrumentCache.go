package main

import (
    "github.com/bhutch29/sclipi/internal/utils"
    "sync"
    "time"
    "log"
)

type instrumentCache struct {
    mu    sync.RWMutex
    cache map[string]utils.Instrument
}

func newInstrumentCache() *instrumentCache {
    return &instrumentCache{
        cache: make(map[string]utils.Instrument),
    }
}

func (ic *instrumentCache) get(address string, port string, timeout time.Duration, progressFn func(int)) (utils.Instrument, error) {
    fullAddress := address + ":" + port

    ic.mu.RLock()
    inst, exists := ic.cache[fullAddress]
    ic.mu.RUnlock()

    if exists {
        return inst, nil
    }

    ic.mu.Lock()
    defer ic.mu.Unlock()

    if inst, exists := ic.cache[fullAddress]; exists {
        return inst, nil
    }

    inst, err := connectInstrument(address, port, timeout, progressFn)
    if err != nil {
        return nil, err
    }

    ic.cache[fullAddress] = inst
    return inst, nil
}

func connectInstrument(address string, port string, timeout time.Duration, progressFn func(int)) (utils.Instrument, error) {
    log.Printf("Connecting to instrument at address '%s'\n", address)
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

