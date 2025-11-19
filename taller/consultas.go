package main

import (
	"fmt"
	"strconv"
)

func menuConsultas() {
	for {
		limpiarPantalla()
		fmt.Println("\n--- CONSULTAS Y LISTADOS ---")
		fmt.Println("1. Listar incidencias de un vehículo")
		fmt.Println("2. Listar vehículos de un cliente")
		fmt.Println("3. Listar mecánicos disponibles")
		fmt.Println("4. Listar incidencias de un mecánico")
		fmt.Println("5. Listar clientes con vehículos en taller")
		fmt.Println("6. Listar todas las incidencias y su estado")
		fmt.Println("7. Volver")
		fmt.Print("Opción: ")

		opcion := leerLinea()

		switch opcion {
		case "1":
			listarIncidenciasVehiculo()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "2":
			listarVehiculosCliente()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "3":
			listarMecanicosDisponibles()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "4":
			listarIncidenciasMecanico()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "5":
			listarClientesEnTaller()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "6":
			listarTodasIncidencias()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "7":
			return
		default:
			fmt.Println("Opción no válida")
			fmt.Print("Presione Enter para continuar...")
			leerLinea()
		}
	}
}

func listarIncidenciasVehiculo() {
	fmt.Print("Matrícula del vehículo: ")
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

	fmt.Printf("\nIncidencia asociada al vehículo %s:\n", matricula)
	
	if vehiculo.IDIncidencia == 0 {
		fmt.Println("  No hay incidencias registradas")
		return
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_incidencia_por_id",
		ID: vehiculo.IDIncidencia,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel

	if respuesta.Exito {
		inc := respuesta.Datos.(Incidencia)
		fmt.Printf("  ID: %d | Tipo: %s | Estado: %s | Prioridad: %s | Tiempo Acumulado: %.2f s\n",
			inc.ID, inc.Tipo, inc.Estado, inc.Prioridad, inc.TiempoAcumulado)
	} else {
		fmt.Println("  Error: Incidencia asociada no encontrada en la base de datos.")
	}
}

func listarVehiculosCliente() {
	fmt.Print("ID del cliente: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_cliente_por_id",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	if !respuesta.Exito {
		fmt.Println("Cliente no encontrado")
		return
	}

	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_vehiculos",
		Respuesta: respChannel,
	}
	respuesta = <-respChannel

	if !respuesta.Exito {
		fmt.Println("Error al obtener vehículos:", respuesta.Mensaje)
		return
	}
	vehiculosList := respuesta.Datos.([]Vehiculo)

	fmt.Printf("\nVehículos del cliente %d:\n", id)
	hayVehiculos := false
	for _, v := range vehiculosList {
		if v.IDCliente == id {
			hayVehiculos = true
			fmt.Printf("  %s - %s %s (En taller: %v)\n",
				v.Matricula, v.Marca, v.Modelo, v.EnTaller)
		}
	}
	if !hayVehiculos {
		fmt.Println("  El cliente no tiene vehículos registrados")
	}
}

func listarMecanicosDisponibles() {
	fmt.Println("\n--- MECÁNICOS DISPONIBLES ---")

	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_mecanicos",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if !respuesta.Exito {
		fmt.Println("Error al obtener mecánicos:", respuesta.Mensaje)
		return
	}
	mecanicosList := respuesta.Datos.([]Mecanico)
	
	hayDisponibles := false
	for _, m := range mecanicosList {
		if m.Activo {
			if ch, ok := CanalesTrabajoMecanicos[m.ID]; ok {
				if len(ch) == 0 {
					hayDisponibles = true
					fmt.Printf("  ID: %d | %s | Especialidad: %s\n",
						m.ID, m.Nombre, m.Especialidad)
				}
			}
		}
	}
	
	if !hayDisponibles {
		fmt.Println("  No hay mecánicos disponibles")
	}
}

func listarIncidenciasMecanico() {
	fmt.Print("ID del mecánico: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_incidencias",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	if !respuesta.Exito {
		fmt.Println("Error al obtener incidencias:", respuesta.Mensaje)
		return
	}
	incidenciasList := respuesta.Datos.([]Incidencia)

	fmt.Printf("\nIncidencias asignadas al mecánico %d:\n", id)
	hayIncidencias := false
	for _, inc := range incidenciasList {
		for _, idMec := range inc.MecanicosID {
			if idMec == id {
				hayIncidencias = true
				fmt.Printf("  ID: %d | Tipo: %s | Estado: %s | Prioridad: %s | T. Acumulado: %.2f s\n",
					inc.ID, inc.Tipo, inc.Estado, inc.Prioridad, inc.TiempoAcumulado)
				break
			}
		}
	}
	if !hayIncidencias {
		fmt.Println("  No hay incidencias asignadas a este mecánico")
	}
}

func listarClientesEnTaller() {
	fmt.Println("\n--- CLIENTES CON VEHÍCULOS EN TALLER ---")

	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_estado_taller",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	if !respuesta.Exito {
		fmt.Println("Error al obtener estado del taller:", respuesta.Mensaje)
		return
	}
	
	estadoTaller := respuesta.Datos.(map[string]interface{})
	vehiculosEnTaller := estadoTaller["vehiculosEnTaller"].([]Vehiculo)
	
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_clientes",
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	if !respuesta.Exito {
		fmt.Println("Error al obtener clientes:", respuesta.Mensaje)
		return
	}
	clientesList := respuesta.Datos.([]Cliente)


	// Crear un mapa con los IDs de clientes que tienen vehículos en taller
	clientesEnTaller := make(map[int]bool)
	for _, v := range vehiculosEnTaller {
		clientesEnTaller[v.IDCliente] = true
	}

	// Mostrar información de esos clientes
	hayClientes := false
	for _, c := range clientesList {
		if clientesEnTaller[c.ID] {
			hayClientes = true
			fmt.Printf("  ID: %d | Nombre: %s | Tel: %s\n",
				c.ID, c.Nombre, c.Telefono)
		}
	}
	if !hayClientes {
		fmt.Println("  No hay clientes con vehículos en el taller")
	}
}

func listarTodasIncidencias() {
	fmt.Println("\n--- TODAS LAS INCIDENCIAS DEL TALLER ---")
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_incidencias",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	if !respuesta.Exito {
		fmt.Println("Error al obtener incidencias:", respuesta.Mensaje)
		return
	}
	incidenciasList := respuesta.Datos.([]Incidencia)

	if len(incidenciasList) == 0 {
		fmt.Println("No hay incidencias registradas")
		return
	}

	for _, inc := range incidenciasList {
		fmt.Printf("\nID: %d\n", inc.ID)
		fmt.Printf("  Tipo: %s\n", inc.Tipo)
		fmt.Printf("  Prioridad: %s\n", inc.Prioridad)
		fmt.Printf("  Estado: %s\n", inc.Estado)
		fmt.Printf("  Tiempo Acumulado: %.2f s\n", inc.TiempoAcumulado)
		fmt.Printf("  Descripción: %s\n", inc.Descripcion)
		fmt.Printf("  Mecánicos asignados: %v\n", inc.MecanicosID)
	}
}