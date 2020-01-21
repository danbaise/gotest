/*
@Author : zj
@Time : 2020/1/21
*/
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"
)

func init() {
	gob.Register(&task{})
}

type snapshot struct {
	exitChan chan struct{}
}

func NewSnapshot() *snapshot {
	return &snapshot{
		exitChan: make(chan struct{}),
	}
}

func (s *snapshot) Run(e interface{}) *snapshot {
	go func() {
		interval := time.Second
		ticker := time.NewTicker(interval)

		f, err := os.OpenFile("snapshot", os.O_CREATE|os.O_TRUNC, 0644)
		defer f.Close()
		if err != nil {
			log.Println(err)
		}
		var buf bytes.Buffer
		for {
			select {
			case <-ticker.C:
				func() {
					buf.Reset()
					enc := gob.NewEncoder(&buf)
					err := enc.Encode(e)
					if err != nil {
						log.Println(err)
					}
					f.Truncate(0)
					f.Seek(0, 0)
					f.Write(buf.Bytes())
					f.Sync()

					fmt.Println("once save")
				}()
			case <-s.exitChan:
				return
			}
		}
	}()
	return s
}
