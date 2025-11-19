package main

import (
	"fmt"
	"math/rand"
	"time"
)

// obtenerTiempoMedio devuelve el tiempo de atención requerido en segundos.
func obtenerTiempoMedio(tipo string) float64 {
	switch tipo {
	case "mecanica":
		return 5.0
	case "electrica":
		return 7.0
	case "carroceria":
		return 11.0
	default:
		return 5.0 
	}
}

// mecanicoTrabajador es la goroutine que simula el trabajo de un mecánico.
func mecanicoTrabajador(mecanicoID int, especialidad string, chTrabajo chan PeticionTrabajo) {
	
	for peticion := range chTrabajo {
		
		if mecanicoID >= 1000 {
			fmt.Printf("Mecanico Emergencia %d (%s) atiende incidencia %d de inmediato.\\n", mecanicoID, especialidad, peticion.IDIncidencia)
		}

		// OBTENER la incidencia del Gestor Central 
		respChannel := make(chan RespuestaControl)
		CanalControl <- PeticionControl{
			TipoOperacion: "obtener_incidencia_por_id",
			ID:            peticion.IDIncidencia,
			Respuesta:     respChannel,
		}
		respuesta := <-respChannel
		
		if !respuesta.Exito {
			continue
		}
		
		incidencia := respuesta.Datos.(Incidencia) 
		
		// Determinar el tiempo de atención (simulación con variación)
		tiempoMedio := obtenerTiempoMedio(incidencia.Tipo)
		variacion := (rand.Float64() - 0.5) * 0.4 * tiempoMedio 
		tiempoAtencion := tiempoMedio + variacion
		
		// Simular el trabajo (bloqueo)
		time.Sleep(time.Duration(tiempoAtencion * 1000) * time.Millisecond)

		// Notificar al Gestor Central que el trabajo ha terminado
		notificacion := NotificacionMecanico{
			IDIncidencia: peticion.IDIncidencia,
			TiempoUsado:  tiempoAtencion,
			IDMecanico:   mecanicoID,
			Especialidad: especialidad,
		}
		NotificacionesMecanico <- notificacion
	}
}