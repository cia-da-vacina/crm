package httpapi

import (
	"net/http"

	"github.com/cia-vacina/crm/backend/internal/domain"
)

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.Health)

	mux.HandleFunc("POST /api/v1/auth/login", s.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.Refresh)

	auth := Auth(s.Tokens)
	adminOnly := RequireRoles(domain.RoleAdmin)

	mux.Handle("POST /api/v1/auth/logout", auth(http.HandlerFunc(s.Logout)))
	mux.Handle("GET /api/v1/me", auth(http.HandlerFunc(s.Me)))

	mux.Handle("GET /api/v1/units", auth(http.HandlerFunc(s.ListUnits)))
	mux.Handle("GET /api/v1/units/{unitId}", auth(http.HandlerFunc(s.GetUnit)))
	mux.Handle("POST /api/v1/units", auth(adminOnly(http.HandlerFunc(s.CreateUnit))))
	mux.Handle("PATCH /api/v1/units/{unitId}", auth(adminOnly(http.HandlerFunc(s.UpdateUnit))))

	mux.Handle("GET /api/v1/users", auth(adminOnly(http.HandlerFunc(s.ListUsers))))
	mux.Handle("POST /api/v1/users", auth(adminOnly(http.HandlerFunc(s.CreateUser))))
	mux.Handle("GET /api/v1/users/{userId}", auth(adminOnly(http.HandlerFunc(s.GetUser))))
	mux.Handle("PATCH /api/v1/users/{userId}", auth(adminOnly(http.HandlerFunc(s.UpdateUser))))
	mux.Handle("DELETE /api/v1/users/{userId}", auth(adminOnly(http.HandlerFunc(s.DeleteUser))))
	mux.Handle("PUT /api/v1/users/{userId}/units", auth(adminOnly(http.HandlerFunc(s.SetUserUnits))))

	// CRM POC
	mux.Handle("GET /api/v1/inbox", auth(http.HandlerFunc(s.ListInbox)))
	mux.Handle("GET /api/v1/conversations/{conversationId}", auth(http.HandlerFunc(s.GetConversation)))
	mux.Handle("GET /api/v1/conversations/{conversationId}/messages", auth(http.HandlerFunc(s.ListMessages)))
	mux.Handle("POST /api/v1/conversations/{conversationId}/messages", auth(http.HandlerFunc(s.SendMessage)))
	mux.Handle("POST /api/v1/conversations/{conversationId}/claim", auth(http.HandlerFunc(s.ClaimConversation)))
	mux.Handle("PATCH /api/v1/conversations/{conversationId}/pipeline", auth(http.HandlerFunc(s.UpdatePipeline)))
	mux.Handle("GET /api/v1/followups", auth(http.HandlerFunc(s.ListFollowUps)))
	mux.Handle("POST /api/v1/followups/{followUpId}/complete", auth(http.HandlerFunc(s.CompleteFollowUp)))
	mux.Handle("GET /api/v1/pops", auth(http.HandlerFunc(s.ListPops)))
	writePops := RequireRoles(domain.RoleAdmin, domain.RoleSupervisor)
	mux.Handle("POST /api/v1/pops", auth(writePops(http.HandlerFunc(s.CreatePop))))
	mux.Handle("PATCH /api/v1/pops/{popId}", auth(writePops(http.HandlerFunc(s.UpdatePop))))
	mux.Handle("DELETE /api/v1/pops/{popId}", auth(writePops(http.HandlerFunc(s.DeletePop))))
	mux.Handle("GET /api/v1/loss-reasons", auth(http.HandlerFunc(s.ListLossReasons)))
	mux.Handle("GET /api/v1/dashboard/summary", auth(http.HandlerFunc(s.DashboardSummary)))
	mux.Handle("GET /api/v1/settings/whatsapp", auth(adminOnly(http.HandlerFunc(s.GetWhatsAppSettings))))
	mux.Handle("PUT /api/v1/settings/whatsapp", auth(adminOnly(http.HandlerFunc(s.PutWhatsAppSettings))))

	return Recover(CORS(RequestID(Logger(mux))))
}
