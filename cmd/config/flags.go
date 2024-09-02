package config

import (
	"flag"
)

// переменная FlagRunAddr содержит адрес и порт для запуска сервера
var (
	FlagRunAddr    string
	FlagSuffixAddr string
)

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagSuffixAddr, "b", "http://localhost:8080/", "address before short url")
	flag.Parse()
}
