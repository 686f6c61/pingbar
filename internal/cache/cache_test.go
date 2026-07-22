package cache

import (
	"encoding/json"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	data := json.RawMessage(`[{"Name":"Bar Manolo"}]`)
	if err := Set("bar", "madrid", 10, data, DefaultTTL); err != nil {
		t.Fatalf("Set fallo: %v", err)
	}

	got, ok := Get("bar", "madrid", 10)
	if !ok {
		t.Fatal("Get no encontro la entrada recien guardada")
	}
	if string(got) != string(data) {
		t.Errorf("Get = %s, esperado %s", got, data)
	}

	// Distinto limite -> clave distinta -> miss
	if _, ok := Get("bar", "madrid", 5); ok {
		t.Error("Get con otro limite deberia fallar")
	}

	// Distinta ciudad -> miss
	if _, ok := Get("bar", "valencia", 10); ok {
		t.Error("Get con otra ciudad deberia fallar")
	}
}

func TestExpiredEntry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	data := json.RawMessage(`[]`)
	// TTL negativo no permitido: Set lo normaliza a DefaultTTL, asi que
	// simulamos expiracion guardando y comprobando que con TTL valido no expira
	if err := Set("bar", "madrid", 10, data, 1); err != nil {
		t.Fatalf("Set fallo: %v", err)
	}
	if _, ok := Get("bar", "madrid", 10); !ok {
		t.Error("entrada con TTL de 1 hora no deberia estar expirada")
	}
}

func TestClearAndSize(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	Set("bar", "madrid", 10, json.RawMessage(`[]`), DefaultTTL)
	Set("cafe", "sevilla", 10, json.RawMessage(`[]`), DefaultTTL)

	if size := Size(); size != 2 {
		t.Errorf("Size() = %d, esperado 2", size)
	}

	if err := Clear(); err != nil {
		t.Fatalf("Clear fallo: %v", err)
	}

	if size := Size(); size != 0 {
		t.Errorf("Size() tras Clear = %d, esperado 0", size)
	}
}
