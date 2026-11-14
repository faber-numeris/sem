package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/faber-numeris/sem/internal/core/domain"
)

func (h *Handler) HandleToken(w http.ResponseWriter, r *http.Request) {
	var req domain.TokenRequest
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		req = domain.TokenRequest{
			GrantType:    body["grant_type"],
			ClientID:     body["client_id"],
			ClientSecret: body["client_secret"],
			Username:     body["username"],
			Password:     body["password"],
			Code:         body["code"],
			RedirectURI:  body["redirect_uri"],
			RefreshToken: body["refresh_token"],
			Assertion:    body["assertion"],
			Scope:        body["scope"],
		}
	} else {
		_ = r.ParseForm()
		req = domain.TokenRequest{
			GrantType:    r.FormValue("grant_type"),
			ClientID:     r.FormValue("client_id"),
			ClientSecret: r.FormValue("client_secret"),
			Username:     r.FormValue("username"),
			Password:     r.FormValue("password"),
			Code:         r.FormValue("code"),
			RedirectURI:  r.FormValue("redirect_uri"),
			RefreshToken: r.FormValue("refresh_token"),
			Assertion:    r.FormValue("assertion"),
			Scope:        r.FormValue("scope"),
		}
	}
	if req.ClientID == "" {
		if u, p, ok := r.BasicAuth(); ok {
			req.ClientID = u
			req.ClientSecret = p
		}
	}
	if r.Host != "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		req.InstanceURL = fmt.Sprintf("%s://%s", scheme, r.Host)
	}

	tok, err := h.svc.IssueToken(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	resp := map[string]any{
		"access_token": tok.AccessToken,
		"instance_url": tok.InstanceURL,
		"id":           tok.IDURL,
		"token_type":   tok.TokenType,
		"issued_at":    strconv.FormatInt(tok.IssuedAt, 10),
		"signature":    tok.Signature,
		"scope":        tok.Scope,
	}
	if tok.RefreshToken != "" {
		resp["refresh_token"] = tok.RefreshToken
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := domain.AuthorizeRequest{
		ResponseType: q.Get("response_type"),
		ClientID:     q.Get("client_id"),
		RedirectURI:  q.Get("redirect_uri"),
		Scope:        q.Get("scope"),
		State:        q.Get("state"),
	}
	res, err := h.svc.Authorize(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
		return
	}
	loc := res.RedirectURI
	if res.Code != "" {
		sep := "?"
		if strings.Contains(loc, "?") {
			sep = "&"
		}
		loc = fmt.Sprintf("%s%scode=%s", loc, sep, res.Code)
		if res.State != "" {
			loc += "&state=" + res.State
		}
	} else if res.AccessToken != "" {
		loc = fmt.Sprintf("%s#access_token=%s&token_type=Bearer", loc, res.AccessToken)
	}
	w.Header().Set("Location", loc)
	writeJSON(w, http.StatusFound, map[string]any{
		"code":         res.Code,
		"access_token": res.AccessToken,
		"redirect_uri": loc,
		"state":        res.State,
	})
}

func (h *Handler) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	tok := r.URL.Query().Get("token")
	if tok == "" {
		tok = r.FormValue("token")
	}
	if err := h.svc.RevokeToken(r.Context(), tok); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":             "invalid_token",
			"error_description": err.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleIntrospect(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	tok := r.FormValue("token")
	res, _ := h.svc.IntrospectToken(r.Context(), tok)
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	baseURL := ""
	if r.Host != "" {
		baseURL = "http://" + r.Host
	}
	info, err := h.svc.GetUserInfo(r.Context(), authHeader, baseURL)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, info)
}
