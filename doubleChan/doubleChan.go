package doubleChan

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

//
// Тип для канальных сообщений попарного чтения/записи.
// Первичное назначение - вычисление гипотенуз треугольников тип в канале float64
// @see calcService/calcFloat64 - реализация интерфейса
//
type IDEntry = interface {
	Get() ([2]interface{}, error)
	Square(interface{}) (interface{}, error)
	Add(interface{}, interface{}) (interface{}, error)
	Sqrt(interface{}) (interface{}, error)
}

// Интерфейс блокирующегося канала для парного чтения/записи данных
type IDChannel interface {
	ReadPair(ctx context.Context) (interface{}, interface{}, error)
	WritePair(ctx context.Context, val1 interface{}, val2 interface{}) error
}

// Реализация блокирующегося канала: Тип для парных операций в каналах с захватом
type DChannel struct {
	Dch chan interface{} // "channel for twice push/pull operations"`
	M   sync.Mutex       // "(un)lock channel for twice operations"`
}

func (t *DChannel) ReadPair(ctx context.Context) (interface{}, interface{}, error) {
	var res [2]interface{}
	var ok bool

	fmt.Printf("\nReadPair() lock channel")
	t.M.Lock()
	defer t.M.Unlock()
	for i := 0; i < 2; i++ {
		fmt.Print("\nReadPair() wait select")
		select {
		case res[i], ok = <-t.Dch:
			if !ok {
				fmt.Printf("\nReadPair() error have closed channel")
				return nil, nil, errors.New("DChannel ERROR: Input channel is closed into pair read")
			}
		case <-ctx.Done():
			fmt.Printf("\nReadPair() have Done()")
			return nil, nil, ctx.Err()
		}
	}
	fmt.Printf("\nReadPair() UN-lock channel")
	return res[0], res[1], nil
}

func (t *DChannel) WritePair(ctx context.Context, val1 interface{}, val2 interface{}) error {
	var data = [2]interface{}{val1, val2}

	fmt.Printf("\nWritePair() lock channel")
	t.M.Lock()
	defer t.M.Unlock()
	for i := 0; i < 2; i++ {
		fmt.Printf("\nWritePair wait select")
		select {
		case t.Dch <- data[i]:
		case <-ctx.Done():
			fmt.Printf("\nWritePair have Done()")
			return ctx.Err()
		}
	}
	fmt.Printf("\nWritePair() UN-lock channel")
	return nil
}
