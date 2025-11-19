package main

// Cliente representa un cliente del taller
type Cliente struct {
	ID       int
	Nombre   string
	Telefono string
	Email    string
}

// Vehiculo representa un vehículo en el taller
type Vehiculo struct {
	Matricula    string
	Marca        string
	Modelo       string
	FechaEntrada string
	FechaSalida  string
	IDCliente    int
	IDIncidencia int
	EnTaller     bool
}

type Incidencia struct {
	ID          int
	MecanicosID []int
	Tipo        string 		// "mecanica", "electrica", "carroceria"
	Prioridad   string 		// "baja", "media", "alta", "escalada"
	Descripcion string
	Estado      string 		// "abierta", "en proceso", "cerrada"
	TiempoAcumulado float64 
}

type Mecanico struct {
	ID           int
	Nombre       string
	Especialidad string
	Experiencia  int
	Activo       bool
}

// Mensajes para entender la ejecuccion

type PeticionTrabajo struct {
	IDIncidencia int
}

type NotificacionMecanico struct {
	IDIncidencia int
	TiempoUsado  float64 // Tiempo real que se ha simulado de atención
	IDMecanico   int
	Especialidad string // Especialidad del mecánico que se libera
}

type PeticionControl struct {
	TipoOperacion string // Ej: "obtener_mecanico", "modificar_incidencia", "eliminar_vehiculo"
	ID            int
	Matricula     string // Usado para pasar matrícula o especialidad de emergencia
	Data          interface{} // Datos adicionales para la operación (ej: el nuevo objeto a guardar o un bool)
	Respuesta     chan RespuestaControl // Canal para que el Gestor devuelva el resultado
}

type RespuestaControl struct {
	Datos interface{} // Los datos solicitados (ej: Mecanico, Incidencia, slice de Vehiculos)
	Exito bool
	Mensaje string
}