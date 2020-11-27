package doubleChan

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type IDEntry = interface{} // any type may be in/out channel!

// Интерфейс блокирующегося канала для парного чтения/записи данных
type IDChannel interface {
	ReadPair() ([2]IDEntry, error)
	WritePair( [2]IDEntry ) error
}

// 1. Тип для парных операций в каналах с захватом
type DChannel struct {
	Dch chan IDEntry      `dch:"channel for twice push/pull operations"`
	M   sync.Mutex        `mutex:"(un)lock channel for twice operations"`
	Ctx context.Context   `ctx:"canceled context"`
}

func (t *DChannel) ReadPair() ([2]IDEntry, error) {
	var res[2] IDEntry
	var ok bool

	fmt.Printf("\nReadPair() lock channel")
	t.M.Lock()
	for i := 0; i < 2; i++ {
		fmt.Print("\nReadPair() wait select")
		select {
		case res[i],ok = <-t.Dch:
			if !ok {
				fmt.Printf("\nReadPair() error have closed channel")
				return [2]IDEntry{nil,nil}, errors.New("DChannel ERROR: Input channel is closed into pair read")
			}
		case <- t.Ctx.Done():
			fmt.Printf("\nReadPair() have Done()")
			t.M.Unlock()
			return [2]IDEntry{nil,nil}, t.Ctx.Err()
		}
	}
	t.M.Unlock()
	fmt.Printf("\nReadPair() UN-lock channel")
	return res, nil
}

func (t *DChannel) WritePair( data [2]IDEntry ) error {
	fmt.Printf("\nWritePair() lock channel")
	t.M.Lock()
	for i := 0; i < 2; i++ {
		fmt.Printf("\nWritePair wait select")
		select {
		case t.Dch <- data[i]:
		case <-t.Ctx.Done():
			fmt.Printf("\nWritePair have Done()")
			t.M.Unlock()
			return t.Ctx.Err()
		}
	}
	t.M.Unlock()
	fmt.Printf("\nWritePair() UN-lock channel")
	return nil
}
