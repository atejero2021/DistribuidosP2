package main

import (
	"fmt"
) 

// colaEspera es la lista interna de incidencias esperando a un mecánico.
var colaEspera []PeticionTrabajo 

// gestorDeTaller es la goroutine central que maneja todo el sistema
func gestorDeTaller() {
	
	colaEspera = make([]PeticionTrabajo, 0)
	
	for {
		select {
		
		case peticion := <-PeticionesAsignacion:
			manejarNuevaPeticion(peticion)

		case notificacion := <-NotificacionesMecanico:
			manejarNotificacionMecanico(notificacion)
			
		case control := <-CanalControl:
			manejarPeticionControl(control)

		case <-NuevoMecanicoListo:
			revisarCola() 
		}
	}
}

func manejarNuevaPeticion(peticion PeticionTrabajo) {
	
	incidencia, _, found := buscarIncidenciaPorID(peticion.IDIncidencia)
	if !found {
		return
	}
	
	if incidencia.Estado == "cerrada" {
		return
	}
	
	// Intenta asignar el trabajo. Si falla, va a la cola de espera.
	if asignarTrabajoPorEspecialidad(peticion) {
	} else {
		colaEspera = append(colaEspera, peticion)
	}
}

func manejarNotificacionMecanico(notificacion NotificacionMecanico) {
	incidencia, incIndex, found := buscarIncidenciaPorID(notificacion.IDIncidencia)
	if !found {
		revisarCola()
		return
	}
	
	if incidencia.Estado == "cerrada" {
		revisarCola()
		return
	}
	
	// Actualizar tiempo acumulado y estado
	incidencias[incIndex].TiempoAcumulado += notificacion.TiempoUsado
	incidencias[incIndex].Estado = "en proceso"
	
	// Añadir mecánico a la lista si no está
	yaAsignado := false
	for _, id := range incidencias[incIndex].MecanicosID {
		if id == notificacion.IDMecanico {
			yaAsignado = true
			break
		}
	}
	if !yaAsignado {
		incidencias[incIndex].MecanicosID = append(incidencias[incIndex].MecanicosID, notificacion.IDMecanico)
	}

	tiempoActual := incidencias[incIndex].TiempoAcumulado
	tiempoRequerido := obtenerTiempoMedio(incidencia.Tipo)
	
	// Lógica de escalamiento si supera 15s 
	if tiempoActual > 15.0 && incidencias[incIndex].Prioridad != "escalada" {
		incidencias[incIndex].Prioridad = "escalada"
		
		peticionEscalada := PeticionTrabajo{IDIncidencia: incidencia.ID}
		if reasignarOCotratar(incidencia.Tipo, peticionEscalada) {
			return
		} else {
			// Si no se pudo reasignar NI contratar, la petición va al frente de la cola
			colaEspera = append([]PeticionTrabajo{peticionEscalada}, colaEspera...)
			return
		}
	}
	
	// Verificar si necesita más tiempo
	if tiempoActual < tiempoRequerido {
		tiempoRestante := tiempoRequerido - tiempoActual
		if tiempoRestante > 0.5 {
			peticionContinuar := PeticionTrabajo{IDIncidencia: incidencia.ID}
			if !asignarTrabajoPorEspecialidad(peticionContinuar) {
				colaEspera = append(colaEspera, peticionContinuar)
			}
		} else {
			cerrarIncidencia(incIndex, incidencia.ID)
		}
	} else {
		cerrarIncidencia(incIndex, incidencia.ID)
	}
	
	revisarCola()
}

func cerrarIncidencia(incIndex int, incidenciaID int) {
	incidencias[incIndex].Estado = "cerrada"
	
	liberarPlazaPorIncidencia(incidenciaID)
	
	if canalFinTest != nil {
		select {
		case canalFinTest <- true:
		default:
		}
	}
}

func asignarTrabajoPorEspecialidad(peticion PeticionTrabajo) bool {
	incidencia, _, _ := buscarIncidenciaPorID(peticion.IDIncidencia)
	especialidadRequerida := incidencia.Tipo 
	
	// Buscar por especialidad
	for id, ch := range CanalesTrabajoMecanicos {
		mecanico, _, found := buscarMecanicoPorID(id)
		
		if found && mecanico.Activo && mecanico.Especialidad == especialidadRequerida {
			select {
			case ch <- peticion:
				return true
			default:
				// Mecánico ocupado, intentar con el siguiente	
				continue 
			}
		}
	}
	return false
}

func reasignarOCotratar(especialidadRequerida string, peticion PeticionTrabajo) bool {
	
	// 1. Intentar asignar a CUALQUIER mecánico libre (requisito de prioridad)
	for id, ch := range CanalesTrabajoMecanicos {
		mecanico, _, found := buscarMecanicoPorID(id)
		
		if found && mecanico.Activo {
			select {
			case ch <- peticion:
				return true
			default:
				continue
			}
		}
	}
	
	// 2. Contratar mecánico de emergencia si no hay más mecánicos	
	idNuevoMec := 1000 + len(mecanicos) // ID alto para diferenciar los creados

	mecanico := Mecanico{
		ID: idNuevoMec, 
		Nombre: fmt.Sprintf("Emergencia-%d", idNuevoMec), 
		Especialidad: especialidadRequerida, 
		Experiencia: 0, 
		Activo: true,
	}
	mecanicos = append(mecanicos, mecanico)
	recalcularPlazas()

	// El canal de un mecánico de emergencia debe tener buffer 1 para que acepte la petición
	chTrabajo := make(chan PeticionTrabajo, 1)
	CanalesTrabajoMecanicos[idNuevoMec] = chTrabajo
	go mecanicoTrabajador(idNuevoMec, especialidadRequerida, chTrabajo) 
	
	select {
	case chTrabajo <- peticion:
		return true
	default:
		return false
	}
}

func revisarCola() {
	if len(colaEspera) == 0 {
		return
	}

	peticion := colaEspera[0] 

	// Intenta asignar el primer trabajo de la cola (FIFO)
	if asignarTrabajoPorEspecialidad(peticion) {
		colaEspera = colaEspera[1:]
	}
}

func liberarPlazaPorIncidencia(idIncidencia int) {
	vIndex := -1
	for i := range vehiculos {
		if vehiculos[i].IDIncidencia == idIncidencia {
			vIndex = i
			break
		}
	}

	if vIndex != -1 {
		if vehiculos[vIndex].EnTaller {
			vehiculos[vIndex].EnTaller = false
			plazasOcupadas--
		}
		vehiculos[vIndex].IDIncidencia = 0
	}
}

// FUNCIONES DE BÚSQUEDA (Necesarias para la lógica del Gestor y CanalControl)

func buscarIncidenciaPorID(id int) (Incidencia, int, bool) {
	for i, inc := range incidencias {
		if inc.ID == id {
			return inc, i, true
		}
	}
	return Incidencia{}, -1, false
}

func buscarMecanicoPorID(id int) (Mecanico, int, bool) {
	for i, m := range mecanicos {
		if m.ID == id {
			return m, i, true
		}
	}
	return Mecanico{}, -1, false
}

func buscarVehiculoPorMatricula(matricula string) (Vehiculo, int, bool) {
	for i, v := range vehiculos {
		if v.Matricula == matricula {
			return v, i, true
		}
	}
	return Vehiculo{}, -1, false
}

func buscarClientePorID(id int) (Cliente, int, bool) {
	for i, c := range clientes {
		if c.ID == id {
			return c, i, true
		}
	}
	return Cliente{}, -1, false
}

func buscarVehiculoPorIDCliente(idCliente int) (Vehiculo, int, bool) {
	for i, v := range vehiculos {
		if v.IDCliente == idCliente {
			return v, i, true
		}
	}
	return Vehiculo{}, -1, false
}

func buscarVehiculoPorIDIncidencia(idIncidencia int) (Vehiculo, int, bool) {
	for i, v := range vehiculos {
		if v.IDIncidencia == idIncidencia {
			return v, i, true
		}
	}
	return Vehiculo{}, -1, false
}


// MANEJO DE PETICIONES DE CONTROL (Lectura/Escritura de datos globales)
func manejarPeticionControl(control PeticionControl) {
	resp := RespuestaControl{Exito: false}

	switch control.TipoOperacion {
	
	case "obtener_estado_taller":
		vehiculosEnTaller := []Vehiculo{}
		for _, v := range vehiculos {
			if v.EnTaller {
				vehiculosEnTaller = append(vehiculosEnTaller, v)
			}
		}
		resp.Datos = map[string]interface{}{
			"totalPlazas": totalPlazas,
			"plazasOcupadas": plazasOcupadas,
			"vehiculosEnTaller": vehiculosEnTaller,
		}
		resp.Exito = true
	
	case "obtener_clientes":
		resp.Datos = clientes 
		resp.Exito = true
	case "obtener_vehiculos":
		resp.Datos = vehiculos
		resp.Exito = true
	case "obtener_incidencias":
		resp.Datos = incidencias
		resp.Exito = true
	case "obtener_mecanicos":
		resp.Datos = mecanicos
		resp.Exito = true

	case "obtener_mecanico_por_id":
		m, _, found := buscarMecanicoPorID(control.ID)
		if found {
			resp.Datos = m
			resp.Exito = true
		} else {
			resp.Mensaje = "Mecánico no encontrado"
		}
	case "obtener_incidencia_por_id":
		inc, _, found := buscarIncidenciaPorID(control.ID)
		if found {
			resp.Datos = inc
			resp.Exito = true
		} else {
			resp.Mensaje = "Incidencia no encontrada"
		}
	case "obtener_cliente_por_id":
		c, _, found := buscarClientePorID(control.ID)
		if found {
			resp.Datos = c
			resp.Exito = true
		} else {
			resp.Mensaje = "Cliente no encontrado"
		}
	case "obtener_vehiculo_por_matricula":
		v, _, found := buscarVehiculoPorMatricula(control.Matricula)
		if found {
			resp.Datos = v
			resp.Exito = true
		} else {
			resp.Mensaje = "Vehículo no encontrado"
		}
	case "obtener_vehiculo_por_incidencia_id":
		v, _, found := buscarVehiculoPorIDIncidencia(control.ID)
		if found {
			resp.Datos = v
			resp.Exito = true
		} else {
			resp.Mensaje = "Vehículo no asociado a esta incidencia"
		}

	case "crear_mecanico":
		nuevoMec := control.Data.(Mecanico)
		mecanicos = append(mecanicos, nuevoMec)
		
		// Crear canal de trabajo e iniciar goroutine
		chTrabajo := make(chan PeticionTrabajo, 1) 
		CanalesTrabajoMecanicos[control.ID] = chTrabajo
		go mecanicoTrabajador(control.ID, nuevoMec.Especialidad, chTrabajo)
		
		recalcularPlazas() 
		resp.Exito = true
		
	case "modificar_mecanico":
		nuevoMec := control.Data.(Mecanico)
		_, i, found := buscarMecanicoPorID(control.ID)
		if found {
			mecanicos[i] = nuevoMec
			recalcularPlazas()
			resp.Exito = true
		} else {
			resp.Mensaje = "Mecánico no encontrado"
		}
		
	case "eliminar_mecanico":
		_, i, found := buscarMecanicoPorID(control.ID)
		if found {
			if ch, ok := CanalesTrabajoMecanicos[control.ID]; ok {
				close(ch)
				delete(CanalesTrabajoMecanicos, control.ID)
			}
			mecanicos = append(mecanicos[:i], mecanicos[i+1:]...)
			recalcularPlazas()
			resp.Exito = true
		} else {
			resp.Mensaje = "Mecánico no encontrado"
		}
		
	case "cambiar_estado_mecanico":
		nuevoEstado := control.Data.(bool)
		_, i, found := buscarMecanicoPorID(control.ID)
		if found {
			mecanicos[i].Activo = nuevoEstado
			recalcularPlazas()
			resp.Exito = true
		} else {
			resp.Mensaje = "Mecánico no encontrado"
		}
	
	case "crear_cliente":
		nuevoCliente := control.Data.(Cliente)
		clientes = append(clientes, nuevoCliente)
		resp.Exito = true
		
	case "modificar_cliente":
		clienteModificado := control.Data.(Cliente)
		_, i, found := buscarClientePorID(control.ID)
		if found {
			clientes[i] = clienteModificado
			resp.Exito = true
		} else {
			resp.Mensaje = "Cliente no encontrado"
		}
		
	case "eliminar_cliente":
		c, i, found := buscarClientePorID(control.ID)
		if found {
			_, _, vFound := buscarVehiculoPorIDCliente(c.ID)
			if vFound {
				resp.Mensaje = "Cliente tiene vehículos registrados"
				goto end_control
			}
			clientes = append(clientes[:i], clientes[i+1:]...)
			resp.Exito = true
		} else {
			resp.Mensaje = "Cliente no encontrado"
		}
		
	case "crear_vehiculo":
		nuevoVehiculo := control.Data.(Vehiculo)
		vehiculos = append(vehiculos, nuevoVehiculo)
		resp.Exito = true
		
	case "modificar_vehiculo":
		vehiculoModificado := control.Data.(Vehiculo)
		_, i, found := buscarVehiculoPorMatricula(control.Matricula)
		if found {
			vehiculos[i] = vehiculoModificado
			resp.Exito = true
		} else {
			resp.Mensaje = "Vehículo no encontrado"
		}

	case "eliminar_vehiculo":
		v, i, found := buscarVehiculoPorMatricula(control.Matricula)
		if found {
			if v.EnTaller {
				resp.Mensaje = "Vehículo en taller"
				goto end_control
			}
			if v.IDIncidencia != 0 {
				inc, _, incFound := buscarIncidenciaPorID(v.IDIncidencia)
				if incFound && (inc.Estado == "abierta" || inc.Estado == "en proceso") {
					resp.Mensaje = "Vehículo con incidencia activa"
					goto end_control
				}
			}
			vehiculos = append(vehiculos[:i], vehiculos[i+1:]...)
			resp.Exito = true
		} else {
			resp.Mensaje = "Vehículo no encontrado"
		}
		
	case "crear_incidencia":
		nuevaIncidencia := control.Data.(Incidencia)
		incidencias = append(incidencias, nuevaIncidencia)
		
		_, iVeh, found := buscarVehiculoPorMatricula(control.Matricula)
		if found {
			vehiculos[iVeh].IDIncidencia = nuevaIncidencia.ID
			resp.Exito = true
		} else {
			resp.Mensaje = "Vehículo no encontrado"
			resp.Exito = false 
		}

	case "modificar_incidencia":
		incidenciaModificada := control.Data.(Incidencia)
		_, i, found := buscarIncidenciaPorID(control.ID)
		if found {
			incidencias[i] = incidenciaModificada
			resp.Exito = true
		} else {
			resp.Mensaje = "Incidencia no encontrada"
		}
	
	case "eliminar_incidencia":
		inc, i, found := buscarIncidenciaPorID(control.ID)
		if found {
			v, iVeh, vFound := buscarVehiculoPorIDIncidencia(inc.ID)
			
			if vFound {
				if v.EnTaller || inc.Estado == "abierta" || inc.Estado == "en proceso" {
					resp.Mensaje = "Incidencia activa o vehículo en taller"
					goto end_control
				}
				vehiculos[iVeh].IDIncidencia = 0
			}
			
			incidencias = append(incidencias[:i], incidencias[i+1:]...)
			resp.Exito = true
		} else {
			resp.Mensaje = "Incidencia no encontrada"
		}
	
	case "cambiar_estado_incidencia":
		_, i, found := buscarIncidenciaPorID(control.ID)
		nuevoEstado := control.Data.(string)
		
		if found {
			estadoAnterior := incidencias[i].Estado
			incidencias[i].Estado = nuevoEstado
			resp.Exito = true
			
			if nuevoEstado == "cerrada" && estadoAnterior != "cerrada" {
				liberarPlazaPorIncidencia(control.ID)
				
				if canalFinTest != nil {
					select {
					case canalFinTest <- true:
					default:
					}
				}
			}
			
		} else {
			resp.Mensaje = "Incidencia no encontrada"
		}

	case "asignar_vehiculo_a_taller":
		v, i, found := buscarVehiculoPorMatricula(control.Matricula)
		if !found {
			resp.Mensaje = "Vehículo no encontrado"
			goto end_control
		}
		if v.EnTaller {
			resp.Mensaje = "Vehículo ya en taller"
			goto end_control
		}
		if plazasOcupadas >= totalPlazas {
			resp.Mensaje = "No hay plazas disponibles"
			goto end_control
		}
		
		vehiculos[i].EnTaller = true
		plazasOcupadas++
		resp.Exito = true

	case "sacar_vehiculo_de_taller":
		v, i, found := buscarVehiculoPorMatricula(control.Matricula)
		if !found {
			resp.Mensaje = "Vehículo no encontrado"
			goto end_control
		}
		if !v.EnTaller {
			resp.Mensaje = "Vehículo no está en taller"
			goto end_control
		}
		
		// Lógica de Advertencia/Forzar salida
		if v.IDIncidencia != 0 {
			inc, _, incFound := buscarIncidenciaPorID(v.IDIncidencia)
			if incFound && (inc.Estado == "abierta" || inc.Estado == "en proceso") {
				if forced, ok := control.Data.(bool); !ok || !forced {
					resp.Mensaje = "Advertencia: Vehículo con incidencia activa. Debe forzar la salida para eliminarlo."
					goto end_control 
				} else {
					// Si es forzado, se fuerza el cierre de la incidencia también
					incidencias[i].Estado = "cerrada"
				}
			}
		}

		vehiculos[i].EnTaller = false
		plazasOcupadas--
		resp.Exito = true

	default:
		resp.Mensaje = fmt.Sprintf("Operación '%s' no reconocida", control.TipoOperacion)
	}

	end_control:
	// La respuesta debe enviarse siempre, incluso en caso de error o éxito.
	control.Respuesta <- resp
}