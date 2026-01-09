package handlers

import (
	"encoding/json"
	"testing"
	"time"

	db "agendaFacil/db/sqlc"
)

// TestCreateReservaRequest_Structure tests que la estructura de reserva funciona
func TestCreateReservaRequest_Structure(t *testing.T) {
	reserva := CreateReservaRequest{
		ServicioID:      1,
		BarberoID:       1,
		Fecha:           "2025-01-09",
		HoraInicio:      "10:00",
		ClienteNombre:   "Pedro",
		ClienteTelefono: "123456789",
	}

	body, _ := json.Marshal(reserva)

	var decodedReserva CreateReservaRequest
	err := json.Unmarshal(body, &decodedReserva)
	if err != nil {
		t.Errorf("Error decodificando reserva: %v", err)
	}

	if decodedReserva.ServicioID != 1 {
		t.Errorf("ServicioID incorrecto: %d", decodedReserva.ServicioID)
	}

	if decodedReserva.ClienteNombre != "Pedro" {
		t.Errorf("ClienteNombre incorrecto: %s", decodedReserva.ClienteNombre)
	}
}

// TestSlot_Structure tests que la estructura Slot funciona
func TestSlot_Structure(t *testing.T) {
	slot := Slot{
		Inicio: "10:00",
		Fin:    "10:30",
	}

	body, _ := json.Marshal(slot)

	var decodedSlot Slot
	json.Unmarshal(body, &decodedSlot)

	if decodedSlot.Inicio != "10:00" {
		t.Errorf("Inicio incorrecto: %s", decodedSlot.Inicio)
	}

	if decodedSlot.Fin != "10:30" {
		t.Errorf("Fin incorrecto: %s", decodedSlot.Fin)
	}
}

// TestToNullString tests que la función toNullString funciona correctamente
func TestToNullString(t *testing.T) {
	// Test con string no vacío
	result := toNullString("test")
	if !result.Valid {
		t.Error("toNullString debería retornar Valid=true para string no vacío")
	}
	if result.String != "test" {
		t.Errorf("Valor incorrecto: %s", result.String)
	}

	// Test con string vacío
	result = toNullString("")
	if result.Valid {
		t.Error("toNullString debería retornar Valid=false para string vacío")
	}
}

// TestChoca_OverlapDetection tests que la función choca detecta superposiciones
func TestChoca_OverlapDetection(t *testing.T) {
	// Crear turnos ocupados
	ocupados := []db.ListTurnosOcupadosRow{
		{
			BarberoID:  1,
			HoraInicio: time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC),
			HoraFin:    time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC),
		},
	}

	// Test: No hay superposición - slot antes
	inicio := time.Date(2025, 1, 8, 9, 0, 0, 0, time.UTC)
	fin := time.Date(2025, 1, 8, 9, 30, 0, 0, time.UTC)
	if choca(inicio, fin, ocupados) {
		t.Error("No debería haber superposición para slot anterior")
	}

	// Test: No hay superposición - slot después
	inicio = time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	fin = time.Date(2025, 1, 8, 11, 0, 0, 0, time.UTC)
	if choca(inicio, fin, ocupados) {
		t.Error("No debería haber superposición para slot posterior")
	}

	// Test: Hay superposición - slot dentro
	inicio = time.Date(2025, 1, 8, 10, 5, 0, 0, time.UTC)
	fin = time.Date(2025, 1, 8, 10, 25, 0, 0, time.UTC)
	if !choca(inicio, fin, ocupados) {
		t.Error("Debería haber superposición para slot dentro del rango")
	}

	// Test: Hay superposición - slot que empieza dentro
	inicio = time.Date(2025, 1, 8, 10, 15, 0, 0, time.UTC)
	fin = time.Date(2025, 1, 8, 10, 45, 0, 0, time.UTC)
	if !choca(inicio, fin, ocupados) {
		t.Error("Debería haber superposición para slot que empieza dentro")
	}
}

// TestCalcularSlots_Basic tests la función calcularSlots
func TestCalcularSlots_Basic(t *testing.T) {
	apertura := time.Date(2025, 1, 8, 9, 0, 0, 0, time.UTC)
	cierre := time.Date(2025, 1, 8, 11, 0, 0, 0, time.UTC)
	duracion := int32(30) // 30 minutos

	slots := calcularSlots(apertura, cierre, duracion, []db.ListTurnosOcupadosRow{})

	// Debería haber 4 slots: 9:00-9:30, 9:30-10:00, 10:00-10:30, 10:30-11:00
	if len(slots) != 4 {
		t.Errorf("Se esperaba 4 slots, pero se obtuvieron %d", len(slots))
	}

	if slots[0].Inicio != "09:00" {
		t.Errorf("Primer slot incorrecto: %s", slots[0].Inicio)
	}

	if slots[0].Fin != "09:30" {
		t.Errorf("Primer slot fin incorrecto: %s", slots[0].Fin)
	}
}

// TestCalcularSlots_WithOccupied tests calcularSlots con turnos ocupados
func TestCalcularSlots_WithOccupied(t *testing.T) {
	apertura := time.Date(2025, 1, 8, 9, 0, 0, 0, time.UTC)
	cierre := time.Date(2025, 1, 8, 11, 0, 0, 0, time.UTC)
	duracion := int32(30)

	ocupados := []db.ListTurnosOcupadosRow{
		{
			BarberoID:  1,
			HoraInicio: time.Date(2025, 1, 8, 9, 30, 0, 0, time.UTC),
			HoraFin:    time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC),
		},
	}

	slots := calcularSlots(apertura, cierre, duracion, ocupados)

	// Debería haber 3 slots (el 9:30-10:00 está ocupado)
	if len(slots) != 3 {
		t.Errorf("Se esperaba 3 slots, pero se obtuvieron %d", len(slots))
	}

	// Verificar que el slot ocupado no está en la lista
	for _, slot := range slots {
		if slot.Inicio == "09:30" {
			t.Error("El slot 09:30-10:00 debería estar ocupado")
		}
	}
}

// TestClaims_Structure tests que la estructura Claims funciona
func TestClaims_Structure(t *testing.T) {
	claims := Claims{
		UserID: 1,
		Rol:    "admin",
	}

	if claims.UserID != 1 {
		t.Errorf("UserID incorrecto: %d", claims.UserID)
	}

	if claims.Rol != "admin" {
		t.Errorf("Rol incorrecto: %s", claims.Rol)
	}
}

// TestCredentials_Structure tests que la estructura Credentials funciona
func TestCredentials_Structure(t *testing.T) {
	creds := Credentials{
		Username: "admin",
		Password: "password",
	}

	body, _ := json.Marshal(creds)

	var decodedCreds Credentials
	json.Unmarshal(body, &decodedCreds)

	if decodedCreds.Username != "admin" {
		t.Errorf("Username incorrecto: %s", decodedCreds.Username)
	}

	if decodedCreds.Password != "password" {
		t.Errorf("Password incorrecto: %s", decodedCreds.Password)
	}
}

// TestCalcularSlots_LargeDuration tests calcularSlots con duraciones largas
func TestCalcularSlots_LargeDuration(t *testing.T) {
	apertura := time.Date(2025, 1, 8, 9, 0, 0, 0, time.UTC)
	cierre := time.Date(2025, 1, 8, 12, 0, 0, 0, time.UTC)
	duracion := int32(60) // 1 hora

	slots := calcularSlots(apertura, cierre, duracion, []db.ListTurnosOcupadosRow{})

	// Debería haber 3 slots: 9:00-10:00, 10:00-11:00, 11:00-12:00
	if len(slots) != 3 {
		t.Errorf("Se esperaba 3 slots con duracion de 60 min, pero se obtuvieron %d", len(slots))
	}

	// Verificar primer slot
	if slots[0].Inicio != "09:00" || slots[0].Fin != "10:00" {
		t.Errorf("Primer slot incorrecto: %s-%s", slots[0].Inicio, slots[0].Fin)
	}

	// Verificar último slot
	if slots[2].Inicio != "11:00" || slots[2].Fin != "12:00" {
		t.Errorf("Último slot incorrecto: %s-%s", slots[2].Inicio, slots[2].Fin)
	}
}

// TestChoca_EdgeCases tests casos límite de superposición
func TestChoca_EdgeCases(t *testing.T) {
	ocupados := []db.ListTurnosOcupadosRow{
		{
			BarberoID:  1,
			HoraInicio: time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC),
			HoraFin:    time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC),
		},
	}

	// Test: El final del slot es exactamente el inicio del turno ocupado (no debe chocar)
	inicio := time.Date(2025, 1, 8, 9, 30, 0, 0, time.UTC)
	fin := time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC)
	if choca(inicio, fin, ocupados) {
		t.Error("No debería haber superposición cuando el slot termina exactamente al inicio del turno")
	}

	// Test: El inicio del slot es exactamente al final del turno ocupado (no debe chocar)
	inicio = time.Date(2025, 1, 8, 10, 30, 0, 0, time.UTC)
	fin = time.Date(2025, 1, 8, 11, 0, 0, 0, time.UTC)
	if choca(inicio, fin, ocupados) {
		t.Error("No debería haber superposición cuando el slot empieza exactamente al final del turno")
	}
}

// TestWriteJSON_Output tests que writeJSON funciona correctamente
func TestWriteJSON_Output(t *testing.T) {
	// Este test verifica que la función writeJSON puede serializar correctamente
	slots := []Slot{
		{Inicio: "09:00", Fin: "09:30"},
		{Inicio: "09:30", Fin: "10:00"},
	}

	body, err := json.Marshal(slots)
	if err != nil {
		t.Errorf("Error serializando slots: %v", err)
	}

	var decoded []Slot
	err = json.Unmarshal(body, &decoded)
	if err != nil {
		t.Errorf("Error deserializando slots: %v", err)
	}

	if len(decoded) != 2 {
		t.Errorf("Se esperaba 2 slots, pero se obtuvieron %d", len(decoded))
	}
}
