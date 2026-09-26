package app

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Evento cancelado não se edita nem se cancela de novo; sala desativada
// não aceita reserva nova mas não trava a edição de um evento antigo;
// sala com reserva futura não é excluída; ocupação de sala inexistente.
func TestCalendarRules(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	_, ana := h.user("nexus-user")
	base := time.Now().UTC().Add(72 * time.Hour).Truncate(time.Hour)
	iso := func(d time.Duration) string { return base.Add(d).Format(time.RFC3339) }
	evBody := func(title, room string, from, to time.Duration) string {
		b := `{"title":"` + title + `","starts_at":"` + iso(from) + `","ends_at":"` + iso(to) + `"`
		if room != "" {
			b += `,"room_id":"` + room + `"`
		}
		return b + `}`
	}
	room := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/calendar/rooms", admin, `{"name":"Auditório `+uuid.NewString()[:5]+`"}`))
	ev := data[eventResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/calendar/events", ana, evBody("Reunião", room.ID, 0, time.Hour)))

	// Sala desativada depois: editar o evento na mesma sala continua ok;
	// reserva nova (ou troca para ela) não.
	h.expect(http.StatusOK, http.MethodPut, "/api/v1/calendar/rooms/"+room.ID, admin, `{"name":"Auditório (reforma)","active":false}`)
	h.expect(http.StatusOK, http.MethodPut, "/api/v1/calendar/events/"+ev.ID, ana, evBody("Reunião renomeada", room.ID, 0, time.Hour))
	h.expect(http.StatusConflict, http.MethodPost, "/api/v1/calendar/events", ana, evBody("Nova", room.ID, 2*time.Hour, 3*time.Hour))
	outro := data[eventResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/calendar/events", ana, evBody("Sem sala", "", 0, time.Hour)))
	h.expect(http.StatusConflict, http.MethodPut, "/api/v1/calendar/events/"+outro.ID, ana, evBody("Sem sala", room.ID, 4*time.Hour, 5*time.Hour))
	h.expect(http.StatusNotFound, http.MethodPost, "/api/v1/calendar/events", ana, evBody("Sala fantasma", uuid.NewString(), 0, time.Hour))
	if list := h.expect(http.StatusOK, http.MethodGet, "/api/v1/calendar/rooms", ana, "").Body.String(); strings.Contains(list, room.ID) {
		t.Fatal("sala inativa não aparece para quem reserva")
	}
	if list := h.expect(http.StatusOK, http.MethodGet, "/api/v1/calendar/rooms", admin, "").Body.String(); !strings.Contains(list, room.ID) {
		t.Fatal("a gestão vê as salas inativas")
	}

	// Sala com reserva futura não é excluída (desative); sem reserva, sai.
	h.expect(http.StatusConflict, http.MethodDelete, "/api/v1/calendar/rooms/"+room.ID, admin, "")
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/calendar/events/"+ev.ID+"/cancel", ana, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/calendar/rooms/"+room.ID, admin, "")
	h.expect(http.StatusNotFound, http.MethodDelete, "/api/v1/calendar/rooms/"+room.ID, admin, "")

	// Cancelado: nem edição, nem segundo cancelamento.
	h.expect(http.StatusConflict, http.MethodPut, "/api/v1/calendar/events/"+ev.ID, ana, evBody("Ressuscitar", "", 0, time.Hour))
	h.expect(http.StatusConflict, http.MethodPost, "/api/v1/calendar/events/"+ev.ID+"/cancel", ana, "")
	var cancels int
	if err := h.d.DB.QueryRow(context.Background(), `SELECT count(*) FROM audit_logs WHERE action = 'calendar.event.cancelled' AND resource_id = $1`,
		ev.ID).Scan(&cancels); err != nil || cancels != 1 {
		t.Fatalf("um único cancelamento auditado: %d %v", cancels, err)
	}
	_, beto := h.user("nexus-user")
	h.expect(http.StatusForbidden, http.MethodDelete, "/api/v1/calendar/events/"+ev.ID, beto, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/calendar/events/"+ev.ID, ana, "")
	h.expect(http.StatusNotFound, http.MethodDelete, "/api/v1/calendar/events/"+ev.ID, ana, "")

	// Consultas: período máximo, sala inexistente e parâmetros malformados.
	window := strings.ReplaceAll("from="+iso(0)+"&to="+iso(time.Hour), "+", "%2B")
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/calendar/rooms/"+uuid.NewString()+"/busy?"+window, ana, "")
	long := strings.ReplaceAll("from="+iso(0)+"&to="+base.Add(100*24*time.Hour).Format(time.RFC3339), "+", "%2B")
	for _, p := range []string{"/api/v1/calendar/events?", "/api/v1/calendar/public/events?"} {
		h.expect(http.StatusUnprocessableEntity, http.MethodGet, p+long, ana, "")
		h.expect(http.StatusBadRequest, http.MethodGet, p+"from=ontem", ana, "")
		h.expect(http.StatusBadRequest, http.MethodGet, p+"to=amanha", ana, "")
	}
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/calendar/events", ana, "") // padrão: mês corrente
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/calendar/events/x", ""},
		{http.MethodPut, "/api/v1/calendar/events/x", evBody("x", "", 0, time.Hour)},
		{http.MethodPost, "/api/v1/calendar/events/x/cancel", ""},
		{http.MethodDelete, "/api/v1/calendar/events/x", ""},
		{http.MethodGet, "/api/v1/calendar/rooms/x/busy", ""},
		{http.MethodPut, "/api/v1/calendar/rooms/x", `{"name":"x"}`},
		{http.MethodDelete, "/api/v1/calendar/rooms/x", ""},
		{http.MethodGet, "/api/v1/calendar/events?room_id=x", ""},
		{http.MethodGet, "/api/v1/calendar/rooms/" + uuid.NewString() + "/busy?from=x", ""},
		{http.MethodPost, "/api/v1/calendar/events", `{`},
		{http.MethodPut, "/api/v1/calendar/events/" + outro.ID, `{`},
		{http.MethodPost, "/api/v1/calendar/rooms", `{`},
	} {
		tok := ana
		if strings.Contains(c.path, "/rooms") && c.method != http.MethodGet {
			tok = admin
		}
		h.expect(http.StatusBadRequest, c.method, c.path, tok, c.body)
	}
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/calendar/events/"+uuid.NewString(), ana, "")
	h.expect(http.StatusNotFound, http.MethodPut, "/api/v1/calendar/events/"+uuid.NewString(), ana, evBody("x", "", 0, time.Hour))
	h.expect(http.StatusNotFound, http.MethodPut, "/api/v1/calendar/rooms/"+uuid.NewString(), admin, `{"name":"Fantasma"}`)
	dup := "Sala dup " + uuid.NewString()[:5]
	h.expect(http.StatusCreated, http.MethodPost, "/api/v1/calendar/rooms", admin, `{"name":"`+dup+`"}`)
	h.expect(http.StatusConflict, http.MethodPost, "/api/v1/calendar/rooms", admin, `{"name":"`+dup+`"}`)
}
