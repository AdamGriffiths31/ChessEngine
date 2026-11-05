package main

import (
	"encoding/gob"
	"fmt"
	"os"
)

type checkpoint struct {
	Epoch     int
	Step      int
	Positions int
	Hidden    int
	Params    [][]float32
	AdamM     [][]float32
	AdamV     [][]float32
}

func (c *checkpoint) save(path string) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := gob.NewEncoder(f).Encode(c); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func loadCheckpoint(path string) (*checkpoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var c checkpoint
	if err := gob.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("reading checkpoint %s: %w", path, err)
	}
	return &c, nil
}

func (c *checkpoint) restore(positions int, m, adamM, adamV *model) error {
	if c.Positions != positions || c.Hidden != hidden {
		return fmt.Errorf("checkpoint was made from %d positions with %d hidden units; this run has %d and %d",
			c.Positions, c.Hidden, positions, hidden)
	}
	for _, pair := range []struct {
		dst *model
		src [][]float32
	}{{m, c.Params}, {adamM, c.AdamM}, {adamV, c.AdamV}} {
		dst := pair.dst.params()
		if len(pair.src) != len(dst) {
			return fmt.Errorf("checkpoint has %d parameter groups, want %d", len(pair.src), len(dst))
		}
		for k := range dst {
			if len(pair.src[k]) != len(dst[k]) {
				return fmt.Errorf("checkpoint parameter group %d has %d values, want %d", k, len(pair.src[k]), len(dst[k]))
			}
			copy(dst[k], pair.src[k])
		}
	}
	return nil
}
