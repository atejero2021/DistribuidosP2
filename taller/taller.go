package main

import (
	"fmt"
	"strings"
)

func asignarVehiculoATaller() {
	fmt.Println("\n--- ASIGNAR VEHÍCULO A TALLER ---")
	
	respChannel := make(chan RespuestaControl)
	
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_estado_taller",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	if !respuesta.Exito {
		fmt.Println("Error: No se pudo obtener el estado del taller.")
		return
	}
	
	estadoTaller := respuesta.Datos.(map[string]interface{})
    
	totalP := estadoTaller["totalPlazas"].(int)
	ocupadas := estadoTaller["plazasOcupadas"].(int)
	
	plazasLibres := totalP - ocupadas
	
	fmt.Printf("Plazas disponibles: %d de %d\n", plazasLibres, totalP)

	if plazasLibres <= 0 {
		fmt.Println("No hay plazas disponibles en el taller")
		return
	}

	fmt.Print("Matrícula del vehículo: ")
	matricula := leerLinea()

	CanalControl <- PeticionControl{
		TipoOperacion: "asignar_vehiculo_a_taller",
		Matricula: matricula,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Vehículo asignado al taller correctamente")
	} else {
		fmt.Println("Error al asignar vehículo:", respuesta.Mensaje)
	}
}

func sacarVehiculoDeTaller() {
	fmt.Println("\n--- SACAR VEHÍCULO DEL TALLER (LIBERAR PLAZA) ---")
	
	fmt.Print("Matrícula del vehículo a sacar: ")
	matricula := leerLinea()

	respChannel := make(chan RespuestaControl)

	CanalControl <- PeticionControl{
		TipoOperacion: "sacar_vehiculo_de_taller",
		Matricula: matricula,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Vehículo sacado del taller. Plaza liberada correctamente.")
	} else {
		fmt.Println("Error al sacar vehículo:", respuesta.Mensaje)
		
		if strings.Contains(respuesta.Mensaje, "Advertencia") {
			fmt.Print("¿Desea continuar y sacar el vehículo de todas formas? (s/n): ")
			confirmacion := leerLinea()
			if confirmacion == "s" {
				CanalControl <- PeticionControl{
					TipoOperacion: "sacar_vehiculo_de_taller",
					Matricula: matricula,
					Data: true, 
					Respuesta: respChannel,
				}
				respuesta = <-respChannel
				if respuesta.Exito {
					fmt.Println("Vehículo sacado del taller (Operación forzada). Plaza liberada.")
				} else {
					fmt.Println("Error al forzar la salida:", respuesta.Mensaje)
				}
			} else {
				fmt.Println("Operación cancelada.")
			}
		}
	}
}


func verEstadoTaller() {
	fmt.Println("\n========================================")
	fmt.Println("       ESTADO DEL TALLER")
	fmt.Println("========================================")
	
	respChannel := make(chan RespuestaControl)
	
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_estado_taller",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Error al obtener el estado del taller:", respuesta.Mensaje)
		return
	}
	estadoTaller := respuesta.Datos.(map[string]interface{})

	totalP, okT := estadoTaller["totalPlazas"].(int)
	ocupadas, okO := estadoTaller["plazasOcupadas"].(int)
	vehiculosEnTaller, okV := estadoTaller["vehiculosEnTaller"].([]Vehiculo)
	
	if !okT || !okO || !okV {
		fmt.Println("Error de formato al recibir datos del Gestor.")
		return
	}
	
	fmt.Printf("Total de plazas: %d\n", totalP)
	fmt.Printf("Plazas ocupadas: %d\n", ocupadas)
	fmt.Printf("Plazas libres: %d\n", totalP-ocupadas)
	fmt.Println("\nVehículos en el taller:")

	if len(vehiculosEnTaller) == 0 {
		fmt.Println("  No hay vehículos en el taller")
		return
	}
	
	for _, v := range vehiculosEnTaller {
		fmt.Printf("  - %s (%s %s) | Cliente ID: %d\n",
			v.Matricula, v.Marca, v.Modelo, v.IDCliente)
	}
}