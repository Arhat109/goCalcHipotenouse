package calcService

import (
	"crypto/md5"
	"errors"
	"fmt"
)

// Реализация канального сообщения как Words (считаем реальные гипотенузы)
type CalcWords string

// Типо зависимая часть поставщика данных
// Читает катеты как Words и возвращает результат.
// Возвращает [0,0] вхлдной поток иссяк (а это фсё!)
// [1,1] -- ошибка тут НЕ число ввели..
func (t *CalcWords) Get() ([2]interface{}, error) {
	var pair [2]string

	fmt.Printf("\nCalcWords::get(): Задайте словосочетание через пробел:")
	num, err := fmt.Scanf("%s %s\n", &pair[0], &pair[1])
	if err != nil {
		fmt.Printf("\nCalcWords::get() Error into read data")
		return [2]interface{}{1, 1}, err
	}
	if num == 0 {
		fmt.Printf("\nCalcWords::get() end of data. Need stop all.")
		return [2]interface{}{0, 0}, nil
	}
	return [2]interface{}{pair[0], pair[1]}, nil
}

// Этот этап не обрабатывает слово никак, просто проверка на совпадение типов сообщений
func (t *CalcWords) Square(x interface{}) (interface{}, error) {
	res, ok := x.(string)
	if !ok {
		return nil, errors.New("\nWords::Square() Param is not Words value!")
	}

	return res, nil
}

// Склеивает словосочетание в одну строку
func (t *CalcWords) Add(x1 interface{}, x2 interface{}) (interface{}, error) {
	res1, ok := x1.(string)
	if !ok {
		return nil, errors.New("\nWords::Add() First param is not Words")
	}

	res2, ok := x2.(string)
	if !ok {
		return nil, errors.New("\nWords::Add() Second param is not Words")
	}

	return res1 + res2, nil
}

// Вычисляет хеш от строки - MD5
func (t *CalcWords) Sqrt(x interface{}) (interface{}, error) {
	res, ok := x.(string)
	if !ok {
		return nil, errors.New("\nWords::Sqrt() Param is not Words")
	}

	return md5.Sum([]byte(res)), nil
}
