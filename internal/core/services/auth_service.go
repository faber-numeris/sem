package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/faber-numeris/sem/internal/core/domain"
)

func (s *Service) IssueToken(ctx context.Context, req domain.TokenRequest) (*domain.OAuthToken, error) {
	var user *domain.SalesforceUser
	clientID := req.ClientID
	if clientID == "" {
		clientID = "default_client_id"
	}
	scope := req.Scope
	if scope == "" {
		scope = "api full refresh_token openid"
	}

	switch req.GrantType {
	case "client_credentials":
		client, err := s.repo.GetClient(ctx, clientID)
		if err != nil || !client.IsActive || (req.ClientSecret != "" && client.ClientSecret != req.ClientSecret) {
			return nil, domain.ErrInvalidClient
		}
		user, err = s.repo.GetDefaultUser(ctx)
		if err != nil {
			return nil, err
		}

	case "password":
		if req.Username != "" {
			u, err := s.repo.GetUserByUsername(ctx, req.Username)
			if err != nil {
				return nil, domain.ErrInvalidGrant
			}
			user = u
		} else {
			u, err := s.repo.GetDefaultUser(ctx)
			if err != nil {
				return nil, err
			}
			user = u
		}

	case "authorization_code":
		if req.Code == "" {
			return nil, domain.ErrInvalidGrant
		}
		authCode, err := s.repo.GetAuthCode(ctx, req.Code)
		if err != nil || authCode.IsUsed || time.Now().After(authCode.ExpiresAt) {
			return nil, domain.ErrInvalidGrant
		}
		if err := s.repo.MarkAuthCodeUsed(ctx, req.Code); err != nil {
			return nil, err
		}
		clientID = authCode.ClientID
		if authCode.Scope != "" {
			scope = authCode.Scope
		}
		user, err = s.repo.GetUserByID(ctx, authCode.UserID)
		if err != nil {
			return nil, err
		}

	case "refresh_token":
		if req.RefreshToken == "" {
			return nil, domain.ErrInvalidGrant
		}
		prevToken, err := s.repo.GetTokenByRefresh(ctx, req.RefreshToken)
		if err != nil || prevToken.IsRevoked {
			return nil, domain.ErrInvalidGrant
		}
		clientID = prevToken.ClientID
		scope = prevToken.Scope
		user, err = s.repo.GetUserByID(ctx, prevToken.UserID)
		if err != nil {
			return nil, err
		}

	case "urn:ietf:params:oauth:grant-type:jwt-bearer":
		if req.Assertion == "" {
			return nil, domain.ErrInvalidGrant
		}
		u, err := s.repo.GetDefaultUser(ctx)
		if err != nil {
			return nil, err
		}
		user = u

	default:
		return nil, domain.ErrInvalidGrant
	}

	instURL := req.InstanceURL
	if instURL == "" {
		instURL = s.instanceURL
	}
	instURL = strings.TrimRight(instURL, "/")

	now := time.Now().UTC()
	accessToken := fmt.Sprintf("%s!%s", user.OrganizationID, randomHex(24))
	refreshToken := fmt.Sprintf("5Aep861%s", randomHex(20))
	idURL := fmt.Sprintf("%s/id/%s/%s", instURL, user.OrganizationID, user.ID)

	tok := &domain.OAuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ClientID:     clientID,
		UserID:       user.ID,
		Scope:        scope,
		InstanceURL:  instURL,
		IDURL:        idURL,
		Signature:    randomHex(20),
		IssuedAt:     now.UnixMilli(),
		ExpiresAt:    now.Add(2 * time.Hour),
	}
	return s.repo.CreateToken(ctx, tok)
}

func (s *Service) Authorize(ctx context.Context, req domain.AuthorizeRequest) (*domain.AuthorizeResult, error) {
	if req.ClientID == "" {
		return nil, domain.ErrInvalidRequest
	}
	client, err := s.repo.GetClient(ctx, req.ClientID)
	if err != nil || !client.IsActive {
		return nil, domain.ErrInvalidClient
	}
	user, err := s.repo.GetDefaultUser(ctx)
	if err != nil {
		return nil, err
	}
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = client.RedirectURI
	}
	scope := req.Scope
	if scope == "" {
		scope = client.Scopes
	}

	switch req.ResponseType {
	case "code", "":
		codeStr := "aPrx" + randomHex(16)
		_, err := s.repo.CreateAuthCode(ctx, &domain.OAuthAuthCode{
			Code:        codeStr,
			ClientID:    client.ClientID,
			UserID:      user.ID,
			RedirectURI: redirectURI,
			Scope:       scope,
			ExpiresAt:   time.Now().UTC().Add(15 * time.Minute),
		})
		if err != nil {
			return nil, err
		}
		return &domain.AuthorizeResult{
			Code:        codeStr,
			RedirectURI: redirectURI,
			State:       req.State,
		}, nil

	case "token":
		tok, err := s.IssueToken(ctx, domain.TokenRequest{
			GrantType: "client_credentials",
			ClientID:  client.ClientID,
			Scope:     scope,
		})
		if err != nil {
			return nil, err
		}
		return &domain.AuthorizeResult{
			AccessToken: tok.AccessToken,
			RedirectURI: redirectURI,
			State:       req.State,
		}, nil

	default:
		return nil, domain.ErrInvalidRequest
	}
}

func (s *Service) RevokeToken(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return domain.ErrInvalidRequest
	}
	if _, err := s.repo.GetTokenByAny(ctx, token); err != nil {
		return domain.ErrInvalidToken
	}
	return s.repo.RevokeToken(ctx, token)
}

func (s *Service) IntrospectToken(ctx context.Context, token string) (*domain.IntrospectResult, error) {
	if strings.TrimSpace(token) == "" {
		return &domain.IntrospectResult{Active: false}, nil
	}
	tok, err := s.repo.GetTokenByAny(ctx, token)
	if err != nil || tok.IsRevoked || time.Now().After(tok.ExpiresAt) {
		return &domain.IntrospectResult{Active: false}, nil
	}
	user, err := s.repo.GetUserByID(ctx, tok.UserID)
	username := ""
	if err == nil {
		username = user.Username
	}
	return &domain.IntrospectResult{
		Active:    true,
		Scope:     tok.Scope,
		ClientID:  tok.ClientID,
		Username:  username,
		TokenType: tok.TokenType,
		Exp:       tok.ExpiresAt.Unix(),
		Iat:       tok.IssuedAt / 1000,
		Sub:       tok.IDURL,
	}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, accessToken, baseURL string) (*domain.UserInfoResult, error) {
	cleanToken := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(accessToken, "Bearer "), "bearer "))
	var user *domain.SalesforceUser
	if cleanToken != "" {
		if tok, err := s.repo.GetTokenByAccess(ctx, cleanToken); err == nil {
			if tok.IsRevoked || time.Now().After(tok.ExpiresAt) {
				return nil, domain.ErrInvalidToken
			}
			user, _ = s.repo.GetUserByID(ctx, tok.UserID)
		}
	}
	if user == nil {
		u, err := s.repo.GetDefaultUser(ctx)
		if err != nil {
			return nil, err
		}
		user = u
	}

	inst := baseURL
	if inst == "" {
		inst = s.instanceURL
	}
	inst = strings.TrimRight(inst, "/")

	return &domain.UserInfoResult{
		Sub:               fmt.Sprintf("%s/id/%s/%s", inst, user.OrganizationID, user.ID),
		UserID:            user.ID,
		OrganizationID:    user.OrganizationID,
		PreferredUsername: user.Username,
		Nickname:          user.Nickname,
		Name:              user.Name,
		Email:             user.Email,
		EmailVerified:     true,
		GivenName:         user.FirstName,
		FamilyName:        user.LastName,
		Zoneinfo:          "America/New_York",
		Photos: map[string]string{
			"picture":   user.PictureURL,
			"thumbnail": user.PictureURL,
		},
		Profile:   fmt.Sprintf("%s/%s", inst, user.ProfileID),
		Picture:   user.PictureURL,
		Urls:      map[string]string{"rest": inst + "/services/data/v{version}/", "sobjects": inst + "/services/data/v{version}/sobjects/"},
		Active:    user.IsActive,
		UserType:  user.UserType,
		Language:  user.Language,
		Locale:    user.Locale,
		UTCOffset: user.UTCOffset,
		UpdatedAt: user.LastModifiedDate.UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}
