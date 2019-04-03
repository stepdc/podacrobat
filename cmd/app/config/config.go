package config

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

type Policy string

const (
	PodsCount string = "podscount"
	NodesLoad string = "loadutil"
)

type Config struct {
	Policy              string
	EvictCountThreshold string
	IdleCountThreshold  string
	// care cpu & memory only
	// percentage
	CpuUtilEvictThreshold string
	CpuUtilIdleThreshold  string

	MemUtilEvictThreshold string
	MemUtilIdleThreshold  string
}

func (cfg *Config) Validate() error {
	log.Printf("config: %#v", *cfg)
	if cfg.Policy == PodsCount {
		_, err := strconv.Atoi(cfg.EvictCountThreshold)
		if err != nil {
			return err
		}
		_, err = strconv.Atoi(cfg.IdleCountThreshold)
		if err != nil {
			return err
		}
	} else if cfg.Policy == NodesLoad {
		if err := validateUtilPercentage(cfg.CpuUtilEvictThreshold); err != nil {
			return err
		}
		if err := validateUtilPercentage(cfg.CpuUtilIdleThreshold); err != nil {
			return err
		}
		if err := validateUtilPercentage(cfg.MemUtilEvictThreshold); err != nil {
			return err
		}
		if err := validateUtilPercentage(cfg.MemUtilIdleThreshold); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported police %q", cfg.Policy)
	}

	return nil
}

var ErrIllegalUtil = errors.New("illegal util percentage")

func validateUtilPercentage(s string) error {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ErrIllegalUtil
	}
	if f < 0 || f > 100 {
		return ErrIllegalUtil
	}

	return nil
}
