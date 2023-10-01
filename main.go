package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Размер буфера Стадии 3.
const bufferSize int = 10

// Интервал времени опустошения буфера на Стадии 3.
const timeInterval time.Duration = 5 * time.Second

// Кольцевой буфер целых чисел.
type RingIntBuffer struct {
	array []int
	pos   int
	size  int
	m     sync.Mutex
}

// Создание нового буфера целых чисел.
func NewRingIntBuffer(size int) *RingIntBuffer {
	return &RingIntBuffer{make([]int, size), 0, size, sync.Mutex{}}
}

// Добавление нового элемента в конец буфера.
// При попытке добавления нового элемента в заполненный буфер самое старое значение затирается.
func (r *RingIntBuffer) Push(el int) {
	pushLog := log.New(os.Stdout, "Step3 (Push): ", log.Ldate|log.Ltime|log.Lshortfile)
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos == r.size {
		r.pos = 0
		r.array[r.pos] = el
		pushLog.Println("Запись в буфер значения: ", el)
	} else {
		r.array[r.pos] = el
		pushLog.Println("Запись в буфер значения: ", el)
		r.pos++

	}
}

// Получение всех элементов буфера и его последующая очистка.
func (r *RingIntBuffer) Get() []int {
	getLog := log.New(os.Stdout, "Step3 (Get): ", log.Ldate|log.Ltime|log.Lshortfile)
	r.m.Lock()
	defer r.m.Unlock()
	var output []int = r.array[:r.pos]
	r.pos = 0
	getLog.Println("Получение значений из буфера: ", output)
	return output
}

// Источник данных. Функция чтения из консоли.
func reader(toStep1 chan<- int, done chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	checkLog := log.New(os.Stdout, "Check: ", log.Ldate|log.Ltime|log.Lshortfile)
	var data string
	fmt.Println("Введите целые положительные или отрицательные числа или 'exit' для выхода из программы")
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			fmt.Println("Программа завершила работу.")
			checkLog.Println("Получен 'exit'. Выход из программы.")
			close(done)
			os.Exit(0)
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("Программа обрабатывает только целые числа!")
			checkLog.Println("Введено не число. Введён символ: ", data)
			continue
		}
		toStep1 <- i
	}
}

// Стадия фильтрации отрицательных чисел.
func step1(fromReader <-chan int, toStep2 chan<- int, done <-chan bool) {
	step1Log := log.New(os.Stdout, "Step1 (Filter): ", log.Ldate|log.Ltime|log.Lshortfile)
	for {
		select {
		case data := <-fromReader:
			if data > 0 {
				step1Log.Println("Значение не отрицательное: ", data, " Обрабатываем дальше.")
				toStep2 <- data
			}
		case <-done:
			return
		}
	}
}

// Стадия фильтрации 0 и чисел не кратных 3.
func step2(fromStep1 <-chan int, toStep3 chan<- int, done <-chan bool) {
	step2Log := log.New(os.Stdout, "Step2 (Filter): ", log.Ldate|log.Ltime|log.Lshortfile)
	for {
		select {
		case data := <-fromStep1:
			if data != 0 && data%3 == 0 {
				step2Log.Println("Значение не '0' и кратно 3: ", data, " Обрабатываем дальше.")
				toStep3 <- data
			}
		case <-done:
			return
		}
	}
}

// Стадия буферизации данных в кольцевом буфере.
func step3(fromStep2 <-chan int, final chan<- int, done <-chan bool, size int) {
	buffer := NewRingIntBuffer(size)
	for {
		select {
		case data := <-fromStep2:
			buffer.Push(data)
		case <-time.After(timeInterval):
			bufferData := buffer.Get()
			if bufferData != nil {
				for _, data := range bufferData {
					final <- data
				}
			}
		case <-done:
			return
		}
	}
}

func main() {

	input := make(chan int)
	done := make(chan bool)
	go reader(input, done)

	channelStep1 := make(chan int)
	go step1(input, channelStep1, done)

	channelStep2 := make(chan int)
	go step2(channelStep1, channelStep2, done)

	channelStep3 := make(chan int)
	go step3(channelStep2, channelStep3, done, bufferSize)

	for {
		select {
		case data := <-channelStep3:
			fmt.Printf("Получены данные: %d\n", data)
		case <-done:
			return
		}
	}
}
