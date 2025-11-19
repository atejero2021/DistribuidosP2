package main

import (
	"fmt"
	"strconv"
)

func menuClientes() {
	for {
		limpiarPantalla()
		fmt.Println("\n--- GESTIÓN DE CLIENTES ---")
		fmt.Println("1. Crear cliente")
		fmt.Println("2. Visualizar clientes")
		fmt.Println("3. Modificar cliente")
		fmt.Println("4. Eliminar cliente")
		fmt.Println("5. Volver")
		fmt.Print("Opción: ")

		opcion := leerLinea()

		switch opcion {
		case "1":
			crearCliente()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "2":
			visualizarClientes()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "3":
			modificarCliente()
			fmt.Print("\nPresione Enter para continuar...")
			leerLinea()
		case "4":
			eliminarCliente()
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

func crearCliente() {
	fmt.Println("\n--- CREAR CLIENTE ---")
	fmt.Print("ID: ")
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
	
	if respuesta.Exito {
		fmt.Println("Error: Ya existe un cliente con ese ID")
		return
	}

	fmt.Print("Nombre: ")
	nombre := leerLinea()
	fmt.Print("Teléfono: ")
	telefono := leerLinea()
	fmt.Print("Email: ")
	email := leerLinea()

	cliente := Cliente{
		ID:       id,
		Nombre:   nombre,
		Telefono: telefono,
		Email:    email,
	}
	
	CanalControl <- PeticionControl{
		TipoOperacion: "crear_cliente",
		ID: id,
		Data: cliente,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel
	
	if respuesta.Exito {
		fmt.Println("Cliente creado correctamente")
	} else {
		fmt.Println("Error al crear el cliente:", respuesta.Mensaje)
	}
}

func visualizarClientes() {
	fmt.Println("\n--- LISTA DE CLIENTES ---")
	
	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_clientes",
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if !respuesta.Exito {
		fmt.Println("Error al obtener clientes:", respuesta.Mensaje)
		return
	}
	
	clientesList := respuesta.Datos.([]Cliente)

	if len(clientesList) == 0 {
		fmt.Println("No hay clientes registrados")
		return
	}
	for _, c := range clientesList {
		fmt.Printf("ID: %d | Nombre: %s | Tel: %s | Email: %s\n",
			c.ID, c.Nombre, c.Telefono, c.Email)
	}
}

func modificarCliente() {
	fmt.Print("ID del cliente a modificar: ")
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
	
	cliente := respuesta.Datos.(Cliente)

	fmt.Printf("Nombre actual: %s\n", cliente.Nombre)
	fmt.Print("Nuevo nombre (enter para mantener): ")
	nombre := leerLinea()
	if nombre != "" {
		cliente.Nombre = nombre
	}
	
	fmt.Printf("Teléfono actual: %s\n", cliente.Telefono)
	fmt.Print("Nuevo teléfono (enter para mantener): ")
	telefono := leerLinea()
	if telefono != "" {
		cliente.Telefono = telefono
	}
	
	fmt.Printf("Email actual: %s\n", cliente.Email)
	fmt.Print("Nuevo email (enter para mantener): ")
	email := leerLinea()
	if email != "" {
		cliente.Email = email
	}

	CanalControl <- PeticionControl{
		TipoOperacion: "modificar_cliente",
		ID: id,
		Data: cliente,
		Respuesta: respChannel,
	}
	respuesta = <-respChannel

	if respuesta.Exito {
		fmt.Println("Cliente modificado correctamente")
	} else {
		fmt.Println("Error al modificar el cliente:", respuesta.Mensaje)
	}
}

func eliminarCliente() {
	fmt.Print("ID del cliente a eliminar: ")
	idStr := leerLinea()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: ID debe ser un número válido")
		return
	}

	respChannel := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "eliminar_cliente",
		ID: id,
		Respuesta: respChannel,
	}
	respuesta := <-respChannel

	if respuesta.Exito {
		fmt.Println("Cliente eliminado correctamente")
	} else {
		fmt.Println("Error al eliminar el cliente:", respuesta.Mensaje)
	}
}