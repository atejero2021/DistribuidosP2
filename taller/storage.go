package main

// Variables globales para almacenamiento de datos (ACCESO SOLO PERMITIDO AL Gestor Central)
var (
	clientes       []Cliente
	vehiculos      []Vehiculo
	incidencias    []Incidencia
	mecanicos      []Mecanico
	plazasOcupadas int
	totalPlazas    int
	
	// channels 
	CanalesTrabajoMecanicos map[int]chan PeticionTrabajo	
	PeticionesAsignacion chan PeticionTrabajo
	NotificacionesMecanico chan NotificacionMecanico
	CanalControl chan PeticionControl
	NuevoMecanicoListo chan int 
	canalFinTest chan bool 
)

// inicializarDatos inicializa las variables globales y channels
func inicializarDatos() {
	totalPlazas = 0
	plazasOcupadas = 0

	// Inicialización del mapa de canales de trabajo
	CanalesTrabajoMecanicos = make(map[int]chan PeticionTrabajo)
	
	// Inicialización de channels (buffers grandes para no bloquear el productor)
	PeticionesAsignacion = make(chan PeticionTrabajo, 100)
	NotificacionesMecanico = make(chan NotificacionMecanico, 100)
	CanalControl = make(chan PeticionControl, 10) 
	NuevoMecanicoListo = make(chan int, 5) 
}


func recalcularPlazas() {
	count := 0
	for _, m := range mecanicos {
		if m.Activo {
			count += 2
		}
	}
	totalPlazas = count
}
