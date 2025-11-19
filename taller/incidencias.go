package main

import (
	"fmt"
	"strconv"
	"strings"
)

func menuIncidencias() {
	for {
		limpiarPantalla()
		fmt.Println("\n--- GESTIÓN DE INCIDENCIAS ---")
		fmt.Println("1. Crear incidencia")
		fmt.Println("2. Visualizar incidencias")
		fmt.Println("3. Modificar incidencia")
		fmt.Println("4. Eliminar incidencia")
		fmt.Println("5. Cambiar estado de incidencia")
		fmt.Println("6. Volver")
		fmt.Print("Opción: ")

		opcion := leerLinea()

		switch opcion {
		case "1":
			crearIncidencia()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "2":
			visualizarIncidencias()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "3":
			modificarIncidencia()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "4":
			eliminarIncidencia()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "5":
			cambiarEstadoIncidencia()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "6":
			return
		default:
			fmt.Println("Opción no válida")
			fmt.Print("Presione Enter para continuar...")
			leerLinea()
		}
	}
}

func crearIncidencia() {
	fmt.Println("\n--- CREAR INCIDENCIA ---")
	
	fmt.Print("ID de la incidencia: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido.")
		return
	}

	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_incidencia_por_id",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Error: Ya existe una incidencia con ese ID. Por favor, ingrese un ID diferente.")
		return
	}
	
	fmt.Print("Matrícula del vehículo: ")
	matricula := leerLinea()

	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_vehiculo_por_matricula",
		Matricula: matricula,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Error: Vehículo no encontrado.")
		return
	}
	
	vehiculo := respuesta.Datos.(Vehiculo)

	// Validación: Un vehículo no puede tener dos incidencias abiertas/en proceso
	if vehiculo.IDIncidencia != 0 {
		CanalControl <- PeticionControl{
			TipoOperacion: "obtener_incidencia_por_id",
			ID: vehiculo.IDIncidencia,
			Respuesta: respChannel,
		}
		respuesta = <-respChannel
		if respuesta.Exito {
			incAnterior := respuesta.Datos.(Incidencia)
			if incAnterior.Estado == "abierta" || incAnterior.Estado == "en proceso" {
				fmt.Println("Error: El vehículo ya tiene una incidencia abierta o en proceso (ID:", vehiculo.IDIncidencia, ").")
				return
			}
		}
	}
	
	fmt.Print("Tipo de incidencia (mecanica, electrica, carroceria): ")
	tipo := strings.ToLower(leerLinea())
	if tipo != "mecanica" && tipo != "electrica" && tipo != "carroceria" {
		fmt.Printf("Error: Tipo de incidencia '%s' no válido.\n", tipo)
		return
	}
	
	fmt.Print("Prioridad (baja, media, alta): ")
	prioridad := strings.ToLower(leerLinea())
	if prioridad != "baja" && prioridad != "media" && prioridad != "alta" {
		fmt.Println("Error: Prioridad no válida.")
		return
	}

	fmt.Print("Descripción: ")
	descripcion := leerLinea()

	incidencia := Incidencia{
		ID:          id, 
		MecanicosID: []int{},
		Tipo:        tipo,
		Prioridad:   prioridad,
		Descripcion: descripcion,
		Estado:      "abierta", 
		TiempoAcumulado: 0.0,
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "crear_incidencia",
		ID: id,
		Matricula: matricula, 
		Data: incidencia,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Incidencia creada correctamente con ID:", id)
		
		PeticionesAsignacion <- PeticionTrabajo{IDIncidencia: id}
		fmt.Println("Petición de asignación enviada al Gestor Central.")
		
	} else {
		fmt.Println("Error al registrar la incidencia:", respuesta.Mensaje)
	}
}

func visualizarIncidencias() {
	fmt.Println("\n--- LISTA DE INCIDENCIAS ---")
	
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
		CanalControl <- PeticionControl{
			TipoOperacion: "obtener_vehiculo_por_incidencia_id",
			ID: inc.ID,
			Respuesta: respChannel,
		}
		respuesta = <-respChannel
		
		var matricula string
		if respuesta.Exito {
			matricula = respuesta.Datos.(Vehiculo).Matricula
		} else {
			matricula = "[NO ASIGNADO]"
		}
		
		fmt.Printf("ID: %d | Matrícula: %s | Tipo: %s | Prioridad: %s | Estado: %s | Tiempo Acumulado: %.2f s\n",
			inc.ID, matricula, inc.Tipo, inc.Prioridad, inc.Estado, inc.TiempoAcumulado)
	}
}

func modificarIncidencia() {
	fmt.Print("ID de la incidencia a modificar: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}

	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_incidencia_por_id",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Incidencia no encontrada")
		return
	}
	
	inc := respuesta.Datos.(Incidencia)

	fmt.Print("Nueva descripción (enter para mantener): ")
	desc := leerLinea()
	if desc != "" {
		inc.Descripcion = desc
	}
	
	fmt.Printf("Prioridad actual: %s\n", inc.Prioridad)
	fmt.Print("Nueva prioridad (baja, media, alta - enter para mantener): ")
	prioridad := strings.ToLower(leerLinea())
	
	if prioridad != "" {
		if prioridad != "baja" && prioridad != "media" && prioridad != "alta" {
			fmt.Println("Error: Prioridad no válida. Se mantiene la prioridad anterior.")
		} else {
			inc.Prioridad = prioridad
		}
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "modificar_incidencia",
		ID: id,
		Data: inc,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel

	if respuesta.Exito {
		fmt.Println("Incidencia modificada correctamente")
	} else {
		fmt.Println("Error al modificar la incidencia:", respuesta.Mensaje)
	}
}

func eliminarIncidencia() {
	fmt.Print("ID de la incidencia a eliminar: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}

	respChannel := make(chan RespuestaControl)
	
	CanalControl <- PeticionControl{
		TipoOperacion: "eliminar_incidencia",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if respuesta.Exito {
		fmt.Println("Incidencia eliminada correctamente. Vehículo desasociado.")
	} else {
		fmt.Println("Error al eliminar la incidencia:", respuesta.Mensaje)
	}
}

func cambiarEstadoIncidencia() {
	fmt.Print("ID de la incidencia: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}

	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_incidencia_por_id",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if !respuesta.Exito {
		fmt.Println("Incidencia no encontrada")
		return
	}
	
	inc := respuesta.Datos.(Incidencia)

	fmt.Printf("Estado actual: %s\n", inc.Estado)
	fmt.Print("Nuevo estado (abierta, en proceso, cerrada): ")
	nuevoEstado := strings.ToLower(leerLinea())

	if nuevoEstado != "abierta" && nuevoEstado != "en proceso" && nuevoEstado != "cerrada" {
		fmt.Println("Error: Estado no válido.")
		return
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "cambiar_estado_incidencia",
		ID: id,
		Data: nuevoEstado,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Estado de la incidencia modificado a:", nuevoEstado)
	} else {
		fmt.Println("Error al cambiar el estado:", respuesta.Mensaje)
	}
}