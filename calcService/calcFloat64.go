package calcService

import (
	"errors"
	"fmt"
	"math"
)

// Реализация канального сообщения как float64 (считаем реальные гипотенузы)
type CalcFloat64 float64

// Типо зависимая часть поставщика данных
// Читает катеты как float64 и возвращает результат.
// Возвращает [0,0] вхлдной поток иссяк (а это фсё!)
// [1,1] -- ошибка тут НЕ число ввели..
func (t *CalcFloat64) Get() ([2]interface{}, error) {
	var pair [2]float64

	fmt.Printf("\nCalcFloat64::get(): Задайте катеты через пробел:")
	num, err := fmt.Scanf("%f %f\n", &pair[0], &pair[1])
	if err != nil {
		fmt.Printf("\nCalcFloat64::get() Error into read data")
		return [2]interface{}{1, 1}, err
	}
	if num == 0 {
		fmt.Printf("\nCalcFloat64::get() end of data. Need stop all.")
		return [2]interface{}{0, 0}, nil
	}
	return [2]interface{}{pair[0], pair[1]}, nil
}

// Возведение в квадрат числа (катет в квадрате)
func (t *CalcFloat64) Square(x interface{}) (interface{}, error) {
	res, ok := x.(float64)
	if !ok {
		return nil, errors.New("\nFloat64::Square() Param is not float64 value!")
	}

	return res * res, nil
}

// Сумматор квадратов катетов как float64
func (t *CalcFloat64) Add(x1 interface{}, x2 interface{}) (interface{}, error) {
	res1, ok := x1.(float64)
	if !ok {
		return nil, errors.New("\nFloat64::Add() First param is not float64")
	}

	res2, ok := x2.(float64)
	if !ok {
		return nil, errors.New("\nFloat64::Add() Second param is not float64")
	}

	return res1 + res2, nil
}

// Извлекатель корня, как результата вычислений гипотенузы
func (t *CalcFloat64) Sqrt(x interface{}) (interface{}, error) {
	res, ok := x.(float64)
	if !ok {
		return nil, errors.New("\nFloat64::Sqrt() Param is not float64")
	}

	return math.Sqrt(res), nil
}
