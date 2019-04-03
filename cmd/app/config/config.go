package config

import (
	"errors"
	"fmt"
	"log"
)

type Policy string

const (
	PodsCount string = "podscount"
	NodesLoad string = "nodeutil"
)

type Config struct {
	Policy string

	EvictCountThreshold int
	IdleCountThreshold  int

	// care cpu & memory only
	// unit is percentage
	CpuUtilEvictThreshold float64
	CpuUtilIdleThreshold  float64
	MemUtilEvictThreshold float64
	MemUtilIdleThreshold  float64
}

func (cfg *Config) Validate() error {
	log.Printf("config: %#v", *cfg)
	if cfg.Policy == PodsCount {
		if err := lessThan(float64(cfg.IdleCountThreshold), float64(cfg.EvictCountThreshold)); err != nil {
			return err
		}
	} else if cfg.Policy == NodesLoad {
		if err := validateUtilPercentage(cfg.CpuUtilEvictThreshold); err != nil {
			return err
		}
		if err := validateUtilPercentage(cfg.CpuUtilIdleThreshold); err != nil {
			return err
		}
		if err := lessThan(cfg.CpuUtilIdleThreshold, cfg.CpuUtilEvictThreshold); err != nil {
			return err
		}
		if err := validateUtilPercentage(cfg.MemUtilEvictThreshold); err != nil {
			return err
		}
		if err := validateUtilPercentage(cfg.MemUtilIdleThreshold); err != nil {
			return err
		}
		if err := lessThan(cfg.MemUtilIdleThreshold, cfg.MemUtilEvictThreshold); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported police %q", cfg.Policy)
	}

	return nil
}

var ErrIllegalUtil = errors.New("illegal util percentage")

func validateUtilPercentage(f float64) error {
	if f < 0 || f > 100 {
		return ErrIllegalUtil
	}

	return nil
}

func lessThan(f1, f2 float64) error {
	if f1 > f2 {
		return fmt.Errorf("parameter not matched")
	}
	return nil
}
