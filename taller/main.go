package main

import (
	"fmt"
	"time" 
)

func crearMecanicoSimulado(id int, nombre, especialidad string, exp int) {
	respChannel := make(chan RespuestaControl)
	
	mecanico := Mecanico{id, nombre, especialidad, exp, true}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "crear_mecanico",
		ID: id,
		Data: mecanico,
		Respuesta: respChannel,
	}
	// Esperar confirmación
	select {
	case <-respChannel:
	case <-time.After(100 * time.Millisecond):
	}
}

func mostrarMenuPrincipal() {
	for {
		limpiarPantalla()
		fmt.Println("\n========================================")
		fmt.Println("    SISTEMA DE GESTIÓN DE TALLER")
		fmt.Println("========================================")
		fmt.Println("1. Gestión de Clientes")
		fmt.Println("2. Gestión de Vehículos")
		fmt.Println("3. Gestión de Incidencias")
		fmt.Println("4. Gestión de Mecánicos")
		fmt.Println("5. Asignar vehículo a plaza del taller")
		fmt.Println("6. Sacar vehículo del taller (Liberar plaza)") 
		fmt.Println("7. Ver estado del taller")
		fmt.Println("8. Consultas y Listados")
		fmt.Println("9. Salir")
		fmt.Print("\nSeleccione una opción: ")

		opcion := leerLinea()

		switch opcion {
		case "1":
			menuClientes() 
		case "2":
			menuVehiculos() 
		case "3":
			menuIncidencias() 
		case "4":
			menuMecanicos() 
		case "5":
			asignarVehiculoATaller() 
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "6":
			sacarVehiculoDeTaller() 
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "7":
			verEstadoTaller() 
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "8":
			menuConsultas() 
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "9":
			limpiarPantalla()
			fmt.Println("Saliendo del sistema...")
			return
		default:
			fmt.Println("Opción no válida")
			fmt.Print("Presione Enter para continuar...")
			leerLinea()
		}
	}
}

func main() {
	inicializarDatos()
	
	// LANZAR LA GOROUTINE CENTRAL
	go gestorDeTaller()
	
	time.Sleep(200 * time.Millisecond)
	
	mostrarMenuPrincipal()
}