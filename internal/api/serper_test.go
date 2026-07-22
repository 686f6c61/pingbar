package api

import (
	"testing"
	"time"
)

func TestNormalizeTime(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"9:00", "09:00"},
		{"09:00", "09:00"},
		{"9:5", "09:05"},
		{"21:30", "21:30"},
		{"9", "09:00"},
		{"9.30", "09:30"},
		{" 10:00 ", "10:00"},
	}

	for _, tt := range tests {
		if got := normalizeTime(tt.in); got != tt.want {
			t.Errorf("normalizeTime(%q) = %q, esperado %q", tt.in, got, tt.want)
		}
	}
}

func TestExtractHoursFromText(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		weekday time.Weekday
		want    string
	}{
		{"rango simple", "Horario: 10:00 - 22:00", time.Monday, "10:00 - 22:00"},
		{"de a", "Abrimos de 9:30 a 21:00 todos los dias", time.Monday, "09:30 - 21:00"},
		{"formato h", "Abierto de 9h a 21h", time.Monday, "09:00 - 21:00"},
		{"dia con horas", "lunes 10:00 a 20:00", time.Monday, "10:00 - 20:00"},
		{"dia con tilde", "Sábado de 10:00 a 14:00", time.Saturday, "10:00 - 14:00"},
		{"prioriza dia actual martes", "lunes: 09:00 - 14:00 martes: 10:00 - 20:00", time.Tuesday, "10:00 - 20:00"},
		{"prioriza dia actual lunes", "lunes: 09:00 - 14:00 martes: 10:00 - 20:00", time.Monday, "09:00 - 14:00"},
		{"24 horas", "Farmacia abierta 24 horas", time.Monday, "Abierto 24 horas"},
		{"sin horario", "El mejor bar de la ciudad", time.Monday, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractHoursFromText(tt.in, tt.weekday); got != tt.want {
				t.Errorf("extractHoursFromText(%q, %v) = %q, esperado %q", tt.in, tt.weekday, got, tt.want)
			}
		})
	}
}

func TestNormalizeForCompare(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Málaga", "malaga"},
		{"CORUÑA", "coruna"},
		{"madrid", "madrid"},
		{"Cádiz, España", "cadiz, espana"},
	}
	for _, tt := range tests {
		if got := normalizeForCompare(tt.in); got != tt.want {
			t.Errorf("normalizeForCompare(%q) = %q, esperado %q", tt.in, got, tt.want)
		}
	}
}

func TestIsOpenAt(t *testing.T) {
	at := func(hour, min int) time.Time {
		return time.Date(2026, 7, 22, hour, min, 0, 0, time.UTC)
	}

	tests := []struct {
		name  string
		hours string
		now   time.Time
		want  bool
	}{
		{"abierto dentro de horario", "10:00 - 22:00", at(15, 0), true},
		{"cerrado antes de abrir", "10:00 - 22:00", at(9, 59), false},
		{"cerrado tras el cierre", "10:00 - 22:00", at(22, 0), false},
		{"justo al abrir", "10:00 - 22:00", at(10, 0), true},
		{"horario nocturno abierto", "20:00 - 02:00", at(1, 0), true},
		{"horario nocturno cerrado", "20:00 - 02:00", at(12, 0), false},
		{"24 horas", "Abierto 24 horas", at(4, 0), true},
		{"sin horario reconocible", "consultar web", at(12, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOpenAt(tt.hours, tt.now); got != tt.want {
				t.Errorf("isOpenAt(%q, %v) = %v, esperado %v", tt.hours, tt.now, got, tt.want)
			}
		})
	}
}
