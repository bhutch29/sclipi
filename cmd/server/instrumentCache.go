package main

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/bhutch29/sclipi/internal/utils"
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

func (ic *instrumentCache) get(address string, port int, timeout time.Duration, progressFn func(int)) (utils.Instrument, error) {
    fullAddress := address + ":" + strconv.Itoa(port)

    ic.mu.RLock()
    inst, exists := ic.cache[fullAddress]
    ic.mu.RUnlock()

    if exists {
        return inst, nil
    }

    ic.mu.Lock()
    defer ic.mu.Unlock()

    inst, err := connectInstrument(address, port, timeout, progressFn)
    if err != nil {
        return nil, err
    }

    ic.cache[fullAddress] = inst
    return inst, nil
}

func (ic *instrumentCache) invalidate(address string, port int) {
    fullAddress := address + ":" + strconv.Itoa(port)

    ic.mu.Lock()
    defer ic.mu.Unlock()

    if inst, exists := ic.cache[fullAddress]; exists {
        inst.Close()
        delete(ic.cache, fullAddress)
        log.Printf("Invalidated cached connection to %s", fullAddress)
    }
}

func connectInstrument(address string, port int, timeout time.Duration, progressFn func(int)) (utils.Instrument, error) {
    log.Printf("Connecting to instrument at address '%s'\n", address)
    var inst utils.Instrument
    if address == "simulated" {
	inst = utils.NewSimInstrument(timeout, false)
    } else {
	inst = utils.NewScpiInstrument(timeout, false)
    }

    if err := inst.Connect(address+":"+strconv.Itoa(port), progressFn); err != nil {
	return inst, err
    }

    return inst, nil
}

