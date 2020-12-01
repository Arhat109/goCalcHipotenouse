package calcService

import (
	"context"
	"fmt"
	dch "github.com/Arhat109/goCalcHypotenuse/doubleChan"
	"sync"
)

// Интерфейс горутин для вычисления гипотенуз треугольников и всего подряд за три операции с каналами:
type IDChanCalculator interface {
	// поставка данных в канал вычисления квадратов катетов из поставщика данных канального типа сообщений
	CalcGet(ctx context.Context, num int, chOut *dch.DChannel) error
	CalcSquare(ctx context.Context, num int, chIn, chOut *dch.DChannel) error
	CalcAdd(ctx context.Context, num int, chIn *dch.DChannel, chOut *chan interface{}) error
	CalcSqrt(ctx context.Context, num int, chIn *chan interface{}) error
}

// Реализация интерфейса калькулятора пока так
type CalcService struct {
	entry dch.IDEntry
}

// 6: Поставщик данных из входного канала (stdin) в парные каналы по очереди. Получает список каналов
func (t *CalcService) CalcGet(ctx context.Context, num int, chOut *dch.DChannel) error {
	fmt.Printf("\nCalcService::Get_%d starting", num)

	for {
		fmt.Printf("\nCalcService::CalcGet(%d): get data from stdin", num)
		pair, err := t.entry.Get()
		if err != nil {
			return err
		} // ошибка чтения входного потока: пришла какая-то фигня.

		if d, ok := pair[0].(int32); ok {
			if int32(d) == 0 {
				return nil
			} // входной поток кончился, а нет больше ничего..
		}
		// тут имеем пару рабочих значений, посылаем их в канал:
		fmt.Printf("\nCalcService::CalcGet(%d): send pair %v to channel", num, pair)
		err = chOut.WritePair(ctx, pair[0], pair[1])
		if err != nil {
			fmt.Printf("\nCalcService::CalcGet(%d): Error from channel is:%s\n", num, err)
			return err
		}
	}
}

// 2: функция для горутины, возвращающая квадрат числа парами:
func (t CalcService) CalcSquare(ctx context.Context, num int, chIn, chOut *dch.DChannel) error {
	fmt.Printf("\nCalcService: +SQUARE_%d start", num)
	var err error // nil by default!
LOOP:
	for {
		fmt.Printf("\nSQUARE_%d wait read pair", num)
		// читаем пару сообщений из канала с блокировкой
		val1, val2, err := chIn.ReadPair(ctx)
		if err != nil {
			break LOOP
		}

		// считаем квадраты принятого как там канальный тип умеет:
		val1, err = t.entry.Square(val1)
		if err != nil {
			break LOOP
		}
		val2, err = t.entry.Square(val2)
		if err != nil {
			break LOOP
		}

		// и пишем в канал результат парой с блокировкой
		fmt.Printf("\nSQUARE_%d wait pair writing", num)
		err = chOut.WritePair(ctx, val1, val2)
		if err != nil {
			break LOOP
		}
		fmt.Printf("\nSQUARE_%d send pair", num)
	}
	fmt.Printf("\nSQUARE_%d has error: %s", num, err)
	return err
}

// 2: функция для горутины, возвращающая в канал сумму пары канальных значений:
func (t CalcService) CalcAdd(ctx context.Context, num int, chIn *dch.DChannel, chOut *chan interface{}) error {
	fmt.Printf("\nCalcService:: +ADD_%d start", num)
	var err error
LOOP:
	for {
		fmt.Printf("\nADD_%d wait read pair", num)
		// читаем пару из канала с блокировкой:
		val1, val2, err := chIn.ReadPair(ctx)
		if err != nil {
			break LOOP
		}

		// считаем сумму как там канальный тип умеет:
		res, err := t.entry.Add(val1, val2)
		if err != nil {
			break LOOP
		}

		// отправляем результат в обычный канал
		fmt.Printf("\nADD_%d into select for sending", num)
		select {
		case *chOut <- res:
		case <-ctx.Done():
			return ctx.Err()
		}
		fmt.Printf("\nADD_%d send sum", num)
	}
	fmt.Printf("\nADD_%d has error is:%s", num, err)
	return err
}

// 2: горутина, возвращающая корень канального значения:
func (t CalcService) CalcSqrt(ctx context.Context, num int, chIn *chan interface{}) error {
	fmt.Printf("\nCalcService: +SQRT_%d start", num)
	var (
		op  interface{}
		ok  bool
		err error
	)
LOOP:
	for {
		fmt.Printf("\nSQRT_%d: wait select", num)
		select {
		case op, ok = <-*chIn:
			if !ok {
				break LOOP
			}
		case <-ctx.Done():
			err = ctx.Err()
			fmt.Printf("\nSQRT_%d get Done with: %s", num, err)
			break LOOP
		}
		op, err = t.entry.Sqrt(op)
		if err != nil {
			break LOOP
		}

		fmt.Printf("\nRESULT from %d!!! %f\n", num, op)
	}
	fmt.Printf("\nSQRT_%d has error: %s", num, err)
	return err
}

// 7: Сервис, запускающий горутины в нужном количестве:
// Возвращает функцию снятия, канал куда слать пары и набор ошибок
// @param num -- количество запускаемых пакетов горутин
func CreateCalc(ctx context.Context, stopFunc func(), strtype string, num int) {
	fmt.Printf("\nCalcServices %d starting .. ", num)
	var (
		chToSquare = dch.DChannel{
			Dch: make(chan interface{}),
			M:   sync.Mutex{},
		}
		chToAdd = dch.DChannel{
			Dch: make(chan interface{}),
			M:   sync.Mutex{},
		}
		chToSum = make(chan interface{})
		tmp     CalcService
	)
	// определяем тип канального сообщения и его методы обработки:
	switch strtype {
	case "float64":
		tmp.entry = new(CalcFloat64)
	case "words":
		tmp.entry = new(CalcWords)
	}
	// запуск поставщика данных из stdin, один для всего пакета горутин
	go func() {
		err := tmp.CalcGet(ctx, -1, &chToSquare)
		stopFunc() // in any case stop all goroutines!

		// Можно писать в лог файл, можно в какую-то глобальную структуру для main(), тут будет так:
		fmt.Printf("\nCalcProvider stopped. Exit code is, err=%s\n", err)
	}()
	// запускаем нужный пакет горутин
	for i := 0; i < num; i++ {
		go func() {
			err := tmp.CalcSquare(ctx, num, &chToSquare, &chToAdd)
			if err != nil {
				stopFunc() // по идее можно, т.к. отвал может быть из-за кривого поиска формата числа..
				// тоже можно логировать или совать куда-то в глобальный контекст.. для наглядности:
				fmt.Printf("\nCalcService(): Error from CalcSquare() is:\n%s", err)
			}
		}()
		go func() {
			err := tmp.CalcAdd(ctx, num, &chToAdd, &chToSum)
			if err != nil {
				stopFunc() // по идее можно, т.к. отвал может быть из-за кривого поиска формата числа..
				// тоже можно логировать или совать куда-то в глобальный контекст.. для наглядности:
				fmt.Printf("\nCalcService(): Error from CalcAdd() is:\n%s", err)
			}
		}()
		go func() {
			err := tmp.CalcSqrt(ctx, num, &chToSum)
			if err != nil {
				stopFunc() // по идее можно, т.к. отвал может быть из-за кривого поиска формата числа..
				// тоже можно логировать или совать куда-то в глобальный контекст.. для наглядности:
				fmt.Printf("\nCalcService(): Error from CalcSqrt() is:\n%s", err)
			}
		}()
	}
	return
}

// 8 Фабрика создает провайдера данных и заданное количество сервисов с горутинами
func CalcFabric(countServices int, strtype string) error {
	fmt.Printf("\nCalcFabric start with %d services for %s type", countServices, strtype)

	ctx, stopFunc := context.WithCancel(context.Background())

	CreateCalc(ctx, stopFunc, strtype, countServices)
	fmt.Printf("\n.. CalcFabric(): Services %d started", countServices)

	return nil
}
