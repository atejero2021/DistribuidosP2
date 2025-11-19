package main

import (
	"fmt"
	"strconv"
)

func menuMecanicos() {
	for {
		limpiarPantalla()
		fmt.Println("\n--- GESTIÓN DE MECÁNICOS ---")
		fmt.Println("1. Crear mecánico")
		fmt.Println("2. Visualizar mecánicos")
		fmt.Println("3. Modificar mecánico")
		fmt.Println("4. Eliminar mecánico")
		fmt.Println("5. Dar de alta/baja mecánico")
		fmt.Println("6. Volver")
		fmt.Print("Opción: ")

		opcion := leerLinea()

		switch opcion {
		case "1":
			crearMecanico()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "2":
			visualizarMecanicos()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "3":
			modificarMecanico()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "4":
			eliminarMecanico()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "5":
			cambiarEstadoMecanico()
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

func crearMecanico() {
	fmt.Println("\n--- CREAR MECÁNICO ---")
	fmt.Print("ID: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}

	// 1. Validar unicidad de ID a través del Gestor Central
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_mecanico_por_id",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Error: Ya existe un mecánico con ese ID")
		return
	}

	fmt.Print("Nombre: ")
	nombre := leerLinea()
	fmt.Print("Especialidad (mecanica/electrica/carroceria): ")
	especialidad := leerLinea()
	fmt.Print("Años de experiencia: ")
	expStr := leerLinea()
	experiencia, err := strconv.Atoi(expStr)
	if err != nil {
		fmt.Println("Error: Años de experiencia debe ser un número válido")
		return
	}

	mecanico := Mecanico{id, nombre, especialidad, experiencia, true}

	CanalControl <- PeticionControl{
		TipoOperacion: "crear_mecanico",
		ID: id,
		Data: mecanico,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel

	if respuesta.Exito {
		// La goroutine se lanza dentro del Gestor al recibir el mensaje de crear_mecanico.
		fmt.Println("Mecánico creado correctamente y puesto en servicio.")
	} else {
		fmt.Println("Error al registrar el mecánico:", respuesta.Mensaje)
	}
}

func visualizarMecanicos() {
	fmt.Println("\n--- LISTA DE MECÁNICOS ---")
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_mecanicos",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Error al obtener la lista de mecánicos:", respuesta.Mensaje)
		return
	}

	mecanicosList := respuesta.Datos.([]Mecanico)

	if len(mecanicosList) == 0 {
		fmt.Println("No hay mecánicos registrados")
		return
	}
	
	for _, m := range mecanicosList {
		estado := "Activo"
		if !m.Activo {
			estado = "Baja"
		}
		
		ocupado := "Estado centralizado" 
		
		fmt.Printf("ID: %d | Nombre: %s | Especialidad: %s | Exp: %d años | Estado: %s | Ocupación: %s\n",
			m.ID, m.Nombre, m.Especialidad, m.Experiencia, estado, ocupado)
	}
}

func modificarMecanico() {
	fmt.Print("ID del mecánico a modificar: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_mecanico_por_id",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Mecánico no encontrado.")
		return
	}
	
	m := respuesta.Datos.(Mecanico)
	
	fmt.Printf("Nombre actual: %s\n", m.Nombre)
	fmt.Print("Nuevo nombre (enter para mantener): ")
	nombre := leerLinea()
	if nombre == "" {
		nombre = m.Nombre
	}
	
	fmt.Printf("Especialidad actual: %s\n", m.Especialidad)
	fmt.Print("Nueva especialidad (enter para mantener): ")
	especialidad := leerLinea()
	if especialidad == "" {
		especialidad = m.Especialidad
	}

	m.Nombre = nombre
	m.Especialidad = especialidad
	
	CanalControl <- PeticionControl{
		TipoOperacion: "modificar_mecanico",
		ID: id,
		Data: m, 
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Mecánico modificado correctamente")
	} else {
		fmt.Println("Error al modificar el mecánico:", respuesta.Mensaje)
	}
}

func eliminarMecanico() {
	fmt.Print("ID del mecánico a eliminar: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "eliminar_mecanico",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if respuesta.Exito {
		fmt.Println("Mecánico eliminado correctamente y su servicio detenido.")
	} else {
		fmt.Println("Error al eliminar el mecánico:", respuesta.Mensaje)
	}
}

func cambiarEstadoMecanico() {
	fmt.Print("ID del mecánico: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_mecanico_por_id",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel
	
	if !respuesta.Exito {
		fmt.Println("Mecánico no encontrado.")
		return
	}
	
	m := respuesta.Datos.(Mecanico)
	
	fmt.Printf("Estado actual: %s\n", map[bool]string{true: "Activo", false: "Baja"}[m.Activo])
	fmt.Println("1. Dar de alta")
	fmt.Println("2. Dar de baja")
	fmt.Print("Opción: ")
	opcion := leerLinea()

	nuevoEstado := m.Activo
	if opcion == "1" {
		nuevoEstado = true
	} else if opcion == "2" {
		nuevoEstado = false
	} else {
		fmt.Println("Opción no válida")
		return
	}

	CanalControl <- PeticionControl{
		TipoOperacion: "cambiar_estado_mecanico",
		ID: id,
		Data: nuevoEstado, 
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if respuesta.Exito {
		if nuevoEstado {
			fmt.Println("Mecánico dado de alta.")
		} else {
			fmt.Println("Mecánico dado de baja.")
		}
	} else {
		fmt.Println("Error al cambiar el estado:", respuesta.Mensaje)
	}
}