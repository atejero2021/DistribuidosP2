package main

import (
	"fmt"
	"testing"
	"time"
)

// Forma de ejecutar los test: go test -v -bench=. -benchtime=3s

const (
	TiempoMecanica   = 5.0
	TiempoElectrica  = 7.0
	TiempoCarroceria = 11.0
)

// Representa la cantidad de mecánicos por especialidad
type ConfiguracionMecanico struct {
	Especialidad string
	Cantidad     int
}

// FUNCIONES AUXILIARES

func simularInicio(t testing.TB, configuracionMecanicos []ConfiguracionMecanico) {
	// 1. Limpiar estado global
	clientes = nil
	vehiculos = nil
	incidencias = nil
	mecanicos = nil
	plazasOcupadas = 0
	totalPlazas = 0
	CanalesTrabajoMecanicos = make(map[int]chan PeticionTrabajo)
	colaEspera = nil

	// 2. Reinicializar channels
	PeticionesAsignacion = make(chan PeticionTrabajo, 100)
	NotificacionesMecanico = make(chan NotificacionMecanico, 100)
	CanalControl = make(chan PeticionControl, 100) // Buffer grande
	NuevoMecanicoListo = make(chan int, 10)

	// 3. Inicializar canal de finalización para tests
	canalFinTest = make(chan bool, 100) // Buffer grande

	// 4. Lanzar el Gestor Central
	go gestorDeTaller()

	// 5. Crear mecánicos según configuración
	idMec := 1
	for _, config := range configuracionMecanicos {
		for i := 0; i < config.Cantidad; i++ {
			m := Mecanico{
				ID:           idMec,
				Nombre:       fmt.Sprintf("Mec-%s-%d", config.Especialidad[:1], idMec),
				Especialidad: config.Especialidad,
				Activo:       true,
			}

			chResp := make(chan RespuestaControl)
			CanalControl <- PeticionControl{
				TipoOperacion: "crear_mecanico",
				ID:            idMec,
				Data:          m,
				Respuesta:     chResp,
			}

			// Esperar con timeout
			select {
			case <-chResp:
			case <-time.After(200 * time.Millisecond):
				t.Logf("Timeout creando mecánico %d", idMec)
			}

			idMec++
		}
	}

	// Dar tiempo a las goroutines para iniciar
	time.Sleep(200 * time.Millisecond)
	t.Logf("Taller iniciado con %d mecánicos (total plazas: %d)", idMec-1, (idMec-1)*2)
}

func simularCarga(t testing.TB, idCounter *int, numCoches int, tipoIncidencia string) {
	for i := 1; i <= numCoches; i++ {
		*idCounter++
		
		clienteID := *idCounter
		matricula := fmt.Sprintf("MAT-%d", *idCounter)
		incidenciaID := *idCounter

		// Crear cliente
		chResp := make(chan RespuestaControl)
		CanalControl <- PeticionControl{
			TipoOperacion: "crear_cliente",
			ID:            clienteID,
			Data:          Cliente{ID: clienteID, Nombre: fmt.Sprintf("Cliente-%d", clienteID)},
			Respuesta:     chResp,
		}
		select {
		case <-chResp:
		case <-time.After(100 * time.Millisecond):
		}

		// Crear vehículo
		chResp = make(chan RespuestaControl)
		CanalControl <- PeticionControl{
			TipoOperacion: "crear_vehiculo",
			Matricula:     matricula,
			Data: Vehiculo{
				Matricula: matricula,
				IDCliente: clienteID,
				EnTaller:  true,
				Marca:     "Test",
				Modelo:    "Model-" + tipoIncidencia,
			},
			Respuesta: chResp,
		}
		select {
		case <-chResp:
		case <-time.After(100 * time.Millisecond):
		}

		// Asignar a taller
		chResp = make(chan RespuestaControl)
		CanalControl <- PeticionControl{
			TipoOperacion: "asignar_vehiculo_a_taller",
			Matricula:     matricula,
			Respuesta:     chResp,
		}
		select {
		case <-chResp:
		case <-time.After(100 * time.Millisecond):
		}

		// Crear incidencia
		chResp = make(chan RespuestaControl)
		CanalControl <- PeticionControl{
			TipoOperacion: "crear_incidencia",
			ID:            incidenciaID,
			Matricula:     matricula,
			Data: Incidencia{
				ID:        incidenciaID,
				Tipo:      tipoIncidencia,
				Estado:    "abierta",
				Prioridad: "media",
			},
			Respuesta: chResp,
		}
		select {
		case <-chResp:
		case <-time.After(100 * time.Millisecond):
		}

		// Enviar petición de asignación
		PeticionesAsignacion <- PeticionTrabajo{IDIncidencia: incidenciaID}
	}

	t.Logf("Carga de %d coches (%s) lanzada", numCoches, tipoIncidencia)
}

func esperarFinalizacion(t testing.TB, numTrabajos int) {
	timeout := time.After(180 * time.Second) // 3 minutos de timeout
	completados := 0

	for completados < numTrabajos {
		select {
		case <-canalFinTest:
			completados++
			if completados%5 == 0 { // Log cada 5 trabajos
				t.Logf("Completados: %d/%d", completados, numTrabajos)
			}
		case <-timeout:
			t.Fatalf("TIMEOUT: Solo %d/%d trabajos completados", completados, numTrabajos)
			return
		}
	}

	t.Logf("Todos los %d trabajos completados", numTrabajos)
}

// Test benchmarks

const numCochesBase = 5   // Número base de coches para las pruebas (5)
const numCochesMixtos = 8 // Usamos 8 para que la división 50/25/25% sea limpia (4/2/2)


// COMPARATIVA 1: DUPLICAR CARGA

func BenchmarkCarga_Coches_Incidencias(b *testing.B) {
	plantilla := []ConfiguracionMecanico{
		{"mecanica", 1},
		{"electrica", 1},
		{"carroceria", 1},
	}
	// Usamos un contador local para generar IDs únicos en esta ejecución
	idCounter := 0 

	for i := 0; i < b.N; i++ {
		simularInicio(b, plantilla)
		b.ResetTimer()
		idCounter = 0 
		
		simularCarga(b, &idCounter, numCochesBase, "mecanica") // 5 coches
		esperarFinalizacion(b, numCochesBase)

		b.StopTimer()
	}
}

func BenchmarkCarga_DobleCoches_Incidencias(b *testing.B) {
	plantilla := []ConfiguracionMecanico{
		{"mecanica", 1},
		{"electrica", 1},
		{"carroceria", 1},
	}
	idCounter := 0

	for i := 0; i < b.N; i++ {
		simularInicio(b, plantilla)
		b.ResetTimer()
		idCounter = 0
		
		simularCarga(b, &idCounter, numCochesBase*2, "mecanica") // 10 coches
		esperarFinalizacion(b, numCochesBase*2) // Espera 10

		b.StopTimer()
	}
}


// COMPARATIVA 2: DUPLICAR PLANTILLA

func BenchmarkPlantilla_3Mecanicos(b *testing.B) {
	plantilla := []ConfiguracionMecanico{
		{"mecanica", 1},
		{"electrica", 1},
		{"carroceria", 1},
	}
	idCounter := 0

	for i := 0; i < b.N; i++ {
		simularInicio(b, plantilla)
		b.ResetTimer()
		idCounter = 0

		simularCarga(b, &idCounter, numCochesBase, "electrica") // 5 coches
		simularCarga(b, &idCounter, numCochesBase, "mecanica") // 5 coches
		simularCarga(b, &idCounter, numCochesBase, "carroceria") // 5 coches


		esperarFinalizacion(b, numCochesBase * 3) // Espera 15

		b.StopTimer()
	}
}

func BenchmarkPlantilla_6Mecanicos(b *testing.B) {
	plantilla := []ConfiguracionMecanico{
		{"mecanica", 2},
		{"electrica", 2},
		{"carroceria", 2},
	}
	idCounter := 0

	for i := 0; i < b.N; i++ {
		simularInicio(b, plantilla)
		b.ResetTimer()
		idCounter = 0

		simularCarga(b, &idCounter, numCochesBase, "electrica") // 5 coches
		simularCarga(b, &idCounter, numCochesBase, "mecanica") // 5 coches
		simularCarga(b, &idCounter, numCochesBase, "carroceria") // 5 coches


		esperarFinalizacion(b, numCochesBase * 3) // Espera 5

		b.StopTimer()
	}
}

// COMPARATIVA 3: PLANTILLAS DESBALANCEADAS

func BenchmarkPlantilla_3M_1E_1C(b *testing.B) {
	plantilla := []ConfiguracionMecanico{
		{"mecanica", 3},
		{"electrica", 1},
		{"carroceria", 1},
	}
	idCounter := 0

	for i := 0; i < b.N; i++ {
		simularInicio(b, plantilla)
		b.ResetTimer()
		idCounter = 0

		// Carga mixta: 50% mecánica (4), 25% eléctrica (2), 25% carrocería (2). Total: 8
		simularCarga(b, &idCounter, numCochesMixtos/2, "mecanica") // ID: 1-4
		simularCarga(b, &idCounter, numCochesMixtos/4, "electrica") // ID: 5-6
		simularCarga(b, &idCounter, numCochesMixtos/4, "carroceria") // ID: 7-8
		esperarFinalizacion(b, numCochesMixtos) // Espera 8

		b.StopTimer()
	}
}

func BenchmarkPlantilla_1M_3E_3C(b *testing.B) {
	plantilla := []ConfiguracionMecanico{
		{"mecanica", 1},
		{"electrica", 3},
		{"carroceria", 3},
	}
	idCounter := 0

	for i := 0; i < b.N; i++ {
		simularInicio(b, plantilla)
		b.ResetTimer()
		idCounter = 0

		// Carga mixta: 50% mecánica (4), 25% eléctrica (2), 25% carrocería (2). Total: 8
		simularCarga(b, &idCounter, numCochesMixtos/2, "mecanica")
		simularCarga(b, &idCounter, numCochesMixtos/4, "electrica")
		simularCarga(b, &idCounter, numCochesMixtos/4, "carroceria")
		esperarFinalizacion(b, numCochesMixtos) // Espera 8

		b.StopTimer()
	}
}

// TESTS UNITARIOS ADICIONALES

func TestCreacionMecanico(t *testing.T) {
	plantilla := []ConfiguracionMecanico{{"mecanica", 1}}
	simularInicio(t, plantilla)

	// Verificar que hay 1 mecánico
	chResp := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_mecanicos",
		Respuesta:     chResp,
	}

	select {
	case resp := <-chResp:
		if !resp.Exito {
			t.Fatal("No se pudieron obtener los mecánicos")
		}
		mecsList := resp.Datos.([]Mecanico)
		if len(mecsList) != 1 {
			t.Fatalf("Se esperaba 1 mecánico, se obtuvieron %d", len(mecsList))
		}
		t.Logf("Test pasado: 1 mecánico creado correctamente")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout obteniendo mecánicos")
	}
}

func TestPlazasPorMecanico(t *testing.T) {
	plantilla := []ConfiguracionMecanico{{"mecanica", 3}} // 3 mecánicos = 6 plazas
	simularInicio(t, plantilla)

	time.Sleep(300 * time.Millisecond) // Esperar inicialización

	chResp := make(chan RespuestaControl)
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_estado_taller",
		Respuesta:     chResp,
	}

	select {
	case resp := <-chResp:
		if !resp.Exito {
			t.Fatal("No se pudo obtener estado del taller")
		}
		estado := resp.Datos.(map[string]interface{})
		totalPlazas := estado["totalPlazas"].(int)

		esperado := 3 * 2 // 3 mecánicos * 2 plazas cada uno = 6
		if totalPlazas != esperado {
			t.Fatalf("Se esperaban %d plazas, se obtuvieron %d", esperado, totalPlazas)
		}
		t.Logf("Test pasado: 3 mecánicos generan %d plazas (2 por mecánico)", totalPlazas)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout obteniendo estado del taller")
	}
}

func TestEscalamiento15Segundos(t *testing.T) {
	// Usar carrocería (11s) para que con 2 ciclos supere los 15s
	plantilla := []ConfiguracionMecanico{{"carroceria", 1}}
	simularInicio(t, plantilla)

	// Crear una incidencia de carrocería
	// Para este test unitario, simplificamos el proceso manual para evitar colisiones
	id := 1000 
	chResp := make(chan RespuestaControl)
	
	// Crear Cliente, Vehículo, Incidencia con ID único (1000)
	CanalControl <- PeticionControl{TipoOperacion: "crear_cliente", ID: id, Data: Cliente{ID: id, Nombre: "Cliente-Test"}, Respuesta: chResp}
	<-chResp
	CanalControl <- PeticionControl{TipoOperacion: "crear_vehiculo", Matricula: "TEST-15S", Data: Vehiculo{Matricula: "TEST-15S", IDCliente: id, EnTaller: true, Marca: "Test"}, Respuesta: chResp}
	<-chResp
	CanalControl <- PeticionControl{TipoOperacion: "asignar_vehiculo_a_taller", Matricula: "TEST-15S", Respuesta: chResp}
	<-chResp
	CanalControl <- PeticionControl{TipoOperacion: "crear_incidencia", ID: id, Matricula: "TEST-15S", Data: Incidencia{ID: id, Tipo: "carroceria", Estado: "abierta", Prioridad: "media"}, Respuesta: chResp}
	<-chResp
	PeticionesAsignacion <- PeticionTrabajo{IDIncidencia: id}


	// Esperar a que termine (solo 1 trabajo)
	esperarFinalizacion(t, 1)

	// Verificar que la incidencia se cerró
	CanalControl <- PeticionControl{
		TipoOperacion: "obtener_incidencia_por_id",
		ID:            id,
		Respuesta:     chResp,
	}

	select {
	case resp := <-chResp:
		if !resp.Exito {
			t.Fatal("No se pudo obtener la incidencia")
		}
		inc := resp.Datos.(Incidencia)
		if inc.Estado != "cerrada" {
			t.Fatalf("La incidencia debería estar cerrada, estado: %s", inc.Estado)
		}
		t.Logf("Test pasado: Incidencia cerrada tras %.2fs (prioridad: %s)", 
			inc.TiempoAcumulado, inc.Prioridad)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout obteniendo incidencia")
	}
}

func TestColaDeEspera(t *testing.T) {
	plantilla := []ConfiguracionMecanico{{"mecanica", 1}} // Solo 1 mecánico
	simularInicio(t, plantilla)

	// Crear 5 incidencias (más que mecánicos disponibles)
	idCounter := 0
	for i := 1; i <= 5; i++ {
		// Garantizar ID único
		idCounter++
		clienteID := idCounter
		matricula := fmt.Sprintf("COLA-%d", idCounter)
		incidenciaID := idCounter

		chResp := make(chan RespuestaControl)
		CanalControl <- PeticionControl{TipoOperacion: "crear_cliente", ID: clienteID, Data: Cliente{ID: clienteID, Nombre: "Cola-Cliente"}, Respuesta: chResp}
		<-chResp
		CanalControl <- PeticionControl{TipoOperacion: "crear_vehiculo", Matricula: matricula, Data: Vehiculo{Matricula: matricula, IDCliente: clienteID, EnTaller: true, Marca: "Test", Modelo: "Model-Cola"}, Respuesta: chResp}
		<-chResp
		CanalControl <- PeticionControl{TipoOperacion: "asignar_vehiculo_a_taller", Matricula: matricula, Respuesta: chResp}
		<-chResp
		CanalControl <- PeticionControl{TipoOperacion: "crear_incidencia", ID: incidenciaID, Matricula: matricula, Data: Incidencia{ID: incidenciaID, Tipo: "mecanica", Estado: "abierta", Prioridad: "media"}, Respuesta: chResp}
		<-chResp
		PeticionesAsignacion <- PeticionTrabajo{IDIncidencia: incidenciaID}
	}

	// Esperar a que todas terminen
	esperarFinalizacion(t, 5)

	t.Logf("Test pasado: 5 incidencias procesadas con 1 solo mecánico (cola funcionó)")
}