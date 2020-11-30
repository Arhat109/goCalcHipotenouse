package calcService

import (
	"context"
	"errors"
	"fmt"
	dch "github.com/calcHypotenuse/doubleChan"
	"math"
	"sync"
)

// 10. interface for double channel data of any calculation types int, float, etc.
type IDChanCalculator interface {
	// Any, have calculation goroutines:
	CalcSquare(num int, chIn, chOut *dch.DChannel) error
	CalcAdd(num int, chIn *dch.DChannel, chOut *chan interface{}) error
	CalcSqrt(num int, ctx context.Context, chIn, chOut *chan interface{}) error
}

// Реализация интерфейса калькулятора пока так
type CalcService struct{}

// 2: функция для горутины, возвращающая квадрат числа:
func (t CalcService) CalcSquare(num int, chIn, chOut *dch.DChannel) error {
	fmt.Printf(" +SQUARE_%d start", num)
	var err error // nil by default!
LOOP:
	for {
		fmt.Printf("\nSQUARE_%d wait read pair", num)
		pair, err := chIn.ReadPair()
		if err != nil {
			break LOOP
		}

		for i := 0; i < 2; i++ {
			pair[i], err = getSquareByType(pair[i])
			if err != nil {
				break LOOP
			}
		}

		err = chOut.WritePair(pair)
		if err != nil {
			break LOOP
		}
		fmt.Printf("\nSQUARE_%d send pair", num)
	}
	fmt.Printf("\nSQUARE_%d has error: %s", num, err)
	return err
}

// 2: горутина, возвращающая сумму пары канальных значений:
func (t CalcService) CalcAdd(num int, chIn *dch.DChannel, chOut *chan interface{}) error {
	fmt.Printf(" +ADD_%d start", num)
	var err error
LOOP:
	for {
		fmt.Printf("\nADD_%d wait read pair", num)
		pair, err := chIn.ReadPair()
		if err != nil {
			break LOOP
		}

		res, err := getSumByType(pair)
		if err != nil {
			break LOOP
		}

		*chOut <- res
		fmt.Printf("\nADD_%d send sum", num)
	}
	fmt.Printf("\nADD_%d read error is:%s", num, err)
	return err
}

// 2: горутина, возвращающая корень канального значения:
func (t CalcService) CalcSqrt(num int, ctx context.Context, chIn *chan interface{}) error {
	fmt.Printf(" +SQRT_%d start", num)
	var op dch.IDEntry
	var ok bool
LOOP:
	for {
		fmt.Printf("\nSQRT_%d: wait select", num)
		select {
		case op, ok = <-*chIn:
			if !ok {
				break LOOP
			}
		case <-ctx.Done():
			err := ctx.Err()
			fmt.Printf("\nSQRT_%d get Done with: %s", num, err)
			return err
		}
		op, err := getSqrtByType(op)
		if err != nil {
			return err
		}

		fmt.Printf("\n%d: Hypotenuse is %f\n", num, op)
	}
	fmt.Printf("\nSQRT_%d break by closed", num)
	return nil
}

// 6: Поставщик данных из входного канала (stdin) в парные каналы по очереди. Получает список каналов
// Ошибки: 1 - пришло завершение контекста, 2 - ошибка канала выдачи
func CalcInputProvider(chOuts []*dch.DChannel, len int) (int, error) {
	var pair [2]dch.IDEntry

	for j := 0; ; j++ {
		fmt.Printf("\nКатеты через пробел:")
		num, err := fmt.Scanf("%f %f\n", pair[0], pair[1])
		if err != nil {
			fmt.Printf("CalcProvider() Error into read data\n")
			return 1, err
		}
		if num == 0 {
			fmt.Printf("CalcProvider() end of data. Need stop all.\n")
			return 0, nil
		}

		if j >= len {
			j = 0
		} // закольцовываем номер исходящего канала

		fmt.Printf("CalcProvider() send to %d channel", j)
		err = chOuts[j].WritePair(pair)
		if err != nil {
			fmt.Printf("CalcProvider() Error from %d channel is:%s\n", j, err)
			return 2, err
		}
	}
}

// Можно так когда в канал лезет чё попало, т.к. тип op изменяется внутри лесенки:
// Можно реализовать интерфейс dch.IDEntry воткнув в него 3 функции, что тут внутри лесенки для каждого типа реализации.
// пока так..
func getSquareByType(data dch.IDEntry) (dch.IDEntry, error) {
	if op, ok := data.(float64); !ok {
		if op, ok := data.(float32); !ok {
			if op, ok := data.(int32); ok {
				if op, ok := data.(uint32); ok {
					if op, ok := data.(int16); ok {
						if op, ok := data.(uint16); ok {
							return nil, errors.New("fetSquareByType() ERROR: can't convert channel data for calculation types")
						} else {
							data = op * op
						}
					} else {
						data = op * op
					}
				} else {
					data = op * op
				}
			} else {
				data = op * op
			}
		} else {
			data = op * op
		}
	} else {
		data = op * op
	}
	return data, nil
}

func getSumByType(data dch.IDEntry) (dch.IDEntry, error) {
	if op, ok := data.(float64); !ok {
		if op, ok := data.(float32); !ok {
			if op, ok := data.(int32); ok {
				if op, ok := data.(uint32); ok {
					if op, ok := data.(int16); ok {
						if op, ok := data.(uint16); ok {
							return nil, errors.New("fetSquareByType() ERROR: can't convert channel data for calculation types")
						} else {
							data = op + op
						}
					} else {
						data = op + op
					}
				} else {
					data = op + op
				}
			} else {
				data = op + op
			}
		} else {
			data = op + op
		}
	} else {
		data = op + op
	}
	return data, nil
}

func getSqrtByType(data dch.IDEntry) (dch.IDEntry, error) {
	if op, ok := data.(float64); !ok {
		if op, ok := data.(float32); !ok {
			if op, ok := data.(int32); ok {
				if op, ok := data.(uint32); ok {
					if op, ok := data.(int16); ok {
						if op, ok := data.(uint16); ok {
							return nil, errors.New("fetSquareByType() ERROR: can't convert channel data for calculation types")
						} else {
							data = uint16(math.Sqrt(float64(op))) // return to own type from float64
						}
					} else {
						data = int16(math.Sqrt(float64(op)))
					}
				} else {
					data = uint32(math.Sqrt(float64(op)))
				}
			} else {
				data = int32(math.Sqrt(float64(op)))
			}
		} else {
			data = float32(math.Sqrt(float64(op)))
		}
	} else {
		data = math.Sqrt(op)
	}
	return data, nil
}

// 7: Сервис, запускающий горутины в нужном количестве:
// Возвращает функцию снятия, канал куда слать пары и набор ошибок
func CreateCalc(ctx context.Context, stopFunc func(), num int) *dch.DChannel {
	fmt.Printf("\nCalcService %d starting .. ", num)
	var (
		chToSquare = dch.DChannel{
			Dch: make(chan dch.IDEntry),
			M:   sync.Mutex{},
			Ctx: ctx,
		}
		chToAdd = dch.DChannel{
			Dch: make(chan dch.IDEntry),
			M:   sync.Mutex{},
			Ctx: ctx,
		}
		chToSum = make(chan dch.IDEntry)
		tmp     CalcService
	)
	go func() {
		err := tmp.CalcSquare(num, &chToSquare, &chToAdd)
		if err != nil {
			stopFunc() // по идее можно, т.к. отвал может быть из-за кривого поиска формата числа..
			// тоже можно логировать или совать куда-то в глобальный контекст.. для наглядности:
			fmt.Printf("\nCalcService(): Error from CalcSquare() is:\n%s", err)
		}
	}()
	go func() {
		err := tmp.CalcAdd(num, &chToAdd, &chToSum)
		if err != nil {
			stopFunc() // по идее можно, т.к. отвал может быть из-за кривого поиска формата числа..
			// тоже можно логировать или совать куда-то в глобальный контекст.. для наглядности:
			fmt.Printf("\nCalcService(): Error from CalcAdd() is:\n%s", err)
		}
	}()
	go func() {
		err := tmp.CalcSqrt(num, ctx, &chToSum)
		if err != nil {
			stopFunc() // по идее можно, т.к. отвал может быть из-за кривого поиска формата числа..
			// тоже можно логировать или совать куда-то в глобальный контекст.. для наглядности:
			fmt.Printf("\nCalcService(): Error from CalcSqrt() is:\n%s", err)
		}
	}()
	return &chToSquare
}

// 8 Фабрика создает провайдера данных и заданное количество сервисов с горутинами
func CalcFabric(countServices int) error {
	fmt.Printf("\nCalcFabric start with %d services", countServices)

	ctx, stopFunc := context.WithCancel(context.Background())
	squareChannels := make([]*dch.DChannel, countServices)

	for i := 0; i < countServices; i++ {
		squareChannels[i] = CreateCalc(ctx, stopFunc, i)
		fmt.Printf(".. CalcFabric(): Service %d started", i)
	}
	go func() {
		num, err := CalcInputProvider(squareChannels, countServices)
		stopFunc() // in any case stop all goroutines!

		// Можно писать в лог файл, можно в какую-то глобальную структуру для main(), тут будет так:
		fmt.Printf("\nCalcProvider stopped. Exit code is %d, err=%s\n", num, err)
	}()
	return nil
}
