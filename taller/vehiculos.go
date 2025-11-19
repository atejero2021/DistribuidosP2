package main

import (
	"fmt"
	"strconv"
	"time"
)

func menuVehiculos() {
	for {
		limpiarPantalla()
		fmt.Println("\n--- GESTIÓN DE VEHÍCULOS ---")
		fmt.Println("1. Crear vehículo")
		fmt.Println("2. Visualizar vehículos")
		fmt.Println("3. Modificar vehículo")
		fmt.Println("4. Eliminar vehículo")
		fmt.Println("5. Volver")
		fmt.Print("Opción: ")

		opcion := leerLinea()

		switch opcion {
		case "1":
			crearVehiculo()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "2":
			visualizarVehiculos()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "3":
			modificarVehiculo()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "4":
			eliminarVehiculo()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "5":
			return
		default:
			fmt.Println("Opción no válida")
			fmt.Print("Presione Enter para continuar...")
			leerLinea()
		}
	}
}

func crearVehiculo() {
	fmt.Println("\n--- CREAR VEHÍCULO ---")
	
	respChannel := make(chan RespuestaControl)

	fmt.Print("Matrícula: ")
	matricula := leerLinea()

	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_vehiculo_por_matricula",
		Matricula: matricula,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Error: Ya existe un vehículo con esa matrícula.")
		return
	}
	
	fmt.Print("ID del Cliente: ")
	idClienteStr := leerLinea()
	idCliente, err := strconv.Atoi(idClienteStr)
	if err != nil {
		fmt.Println("Error: ID del cliente debe ser un número válido.")
		return
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_cliente_por_id",
		ID: idCliente,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Error: Cliente no encontrado.")
		return
	}

	fmt.Print("Marca: ")
	marca := leerLinea()
	fmt.Print("Modelo: ")
	modelo := leerLinea()

	vehiculo := Vehiculo{
		Matricula:    matricula,
		Marca:        marca,
		Modelo:       modelo,
		FechaEntrada: time.Now().Format("02-01-2006 15:04:05"),
		FechaSalida:  "",
		IDCliente:    idCliente,
		IDIncidencia: 0,
		EnTaller:     false,
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "crear_vehiculo",
		Data: vehiculo,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Vehículo creado correctamente con matrícula:", matricula)
	} else {
		fmt.Println("Error al registrar el vehículo:", respuesta.Mensaje)
	}
}

func visualizarVehiculos() {
	fmt.Println("\n--- LISTA DE VEHÍCULOS ---")
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_vehiculos",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Error al obtener vehículos:", respuesta.Mensaje)
		return
	}
	
	vehiculosList := respuesta.Datos.([]Vehiculo)
	
	if len(vehiculosList) == 0 {
		fmt.Println("No hay vehículos registrados")
		return
	}
	
	for _, v := range vehiculosList {
		// Peticiones de datos auxiliares para la visualización
		
		// 1. Obtener nombre del cliente
		CanalControl <- PeticionControl{
			TipoOperacion: "obtener_cliente_por_id",
			ID: v.IDCliente,
			Respuesta: respChannel,
		}
		resCliente := <-respChannel
		var nombreCliente string
		if resCliente.Exito {
			nombreCliente = resCliente.Datos.(Cliente).Nombre
		} else {
			nombreCliente = "[Cliente Desconocido]"
		}
		
		// 2. Obtener estado de la incidencia
		var estadoIncidencia string
		if v.IDIncidencia != 0 {
			CanalControl <- PeticionControl{
				TipoOperacion: "obtener_incidencia_por_id",
				ID: v.IDIncidencia,
				Respuesta: respChannel,
			}
			resIncidencia := <-respChannel
			if resIncidencia.Exito {
				estadoIncidencia = resIncidencia.Datos.(Incidencia).Estado
			} else {
				estadoIncidencia = "Error al buscar (ID: " + strconv.Itoa(v.IDIncidencia) + ")"
			}
		} else {
			estadoIncidencia = "Ninguna"
		}

		estadoTaller := "NO"
		if v.EnTaller {
			estadoTaller = "SÍ"
		}
		
		fmt.Printf("Matrícula: %s | Cliente: %s | Modelo: %s | En Taller: %s | Incidencia: %s (ID %d)\n",
			v.Matricula, nombreCliente, v.Modelo, estadoTaller, estadoIncidencia, v.IDIncidencia)
	}
}

func modificarVehiculo() {
	fmt.Print("Matrícula del vehículo a modificar: ")
	matricula := leerLinea()

	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_vehiculo_por_matricula",
		Matricula: matricula,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if !respuesta.Exito {
		fmt.Println("Vehículo no encontrado")
		return
	}
	
	vehiculo := respuesta.Datos.(Vehiculo)
	
	fmt.Printf("Marca actual: %s\n", vehiculo.Marca)
	fmt.Print("Nueva marca (enter para mantener): ")
	marca := leerLinea()
	if marca != "" {
		vehiculo.Marca = marca
	}
	
	fmt.Printf("Modelo actual: %s\n", vehiculo.Modelo)
	fmt.Print("Nuevo modelo (enter para mantener): ")
	modelo := leerLinea()
	if modelo != "" {
		vehiculo.Modelo = modelo
	}
	
	fmt.Printf("ID Cliente actual: %d\n", vehiculo.IDCliente)
	fmt.Print("Nuevo ID Cliente (enter para mantener): ")
	idClienteStr := leerLinea()
	
	if idClienteStr != "" {
		idCliente, err := strconv.Atoi(idClienteStr)
		if err != nil {
			fmt.Println("Error: ID del cliente debe ser un número válido. Se mantiene el ID anterior.")
		} else {
			CanalControl <- PeticionControl{
				TipoOperacion: "obtener_cliente_por_id",
				ID: idCliente,
				Respuesta: respChannel,
			}
			resCliente := <-respChannel
			
			if resCliente.Exito {
				vehiculo.IDCliente = idCliente
			} else {
				fmt.Println("Error: El nuevo ID de cliente no existe. Se mantiene el ID anterior.")
			}
		}
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "modificar_vehiculo",
		Matricula: matricula,
		Data: vehiculo,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel

	if respuesta.Exito {
		fmt.Println("Vehículo modificado correctamente")
	} else {
		fmt.Println("Error al modificar el vehículo:", respuesta.Mensaje)
	}
}

func eliminarVehiculo() {
	fmt.Print("Matrícula del vehículo a eliminar: ")
	matricula := leerLinea()

	respChannel := make(chan RespuestaControl)
	
	CanalControl <- PeticionControl{
		TipoOperacion: "eliminar_vehiculo",
		Matricula: matricula,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if respuesta.Exito {
		fmt.Println("Vehículo eliminado correctamente")
	} else {
		fmt.Println("Error al eliminar el vehículo:", respuesta.Mensaje)
	}
}