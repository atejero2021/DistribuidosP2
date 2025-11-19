package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var scanner = bufio.NewScanner(os.Stdin)

func limpiarPantalla() {
	fmt.Print("\033[H\033[2J")
}

func leerLinea() string {
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}