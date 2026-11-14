package postgres

import (
	"context"

	"github.com/faber-numeris/sem/internal/adapters/outbound/postgres/gen"
	"github.com/faber-numeris/sem/internal/core/domain"
)

func (r *Repository) GetClient(ctx context.Context, clientID string) (*domain.OAuthClient, error) {
	c, err := r.queries.GetOAuthClientByClientID(ctx, clientID)
	if err != nil {
		return nil, domain.ErrInvalidClient
	}
	return &domain.OAuthClient{
		ID:           c.ID,
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Name:         c.Name,
		RedirectURI:  c.RedirectUri,
		Scopes:       c.Scopes,
		IsActive:     c.IsActive,
	}, nil
}

func mapUser(u gen.SalesforceSalesforceUser) *domain.SalesforceUser {
	return &domain.SalesforceUser{
		ID:               u.ID,
		OrganizationID:   u.OrganizationID,
		Username:         u.Username,
		Email:            u.Email,
		PasswordHash:     u.PasswordHash,
		Name:             u.Name,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		Nickname:         u.Nickname,
		PictureURL:       u.PictureUrl,
		ProfileID:        u.ProfileID,
		UserType:         u.UserType,
		Language:         u.Language,
		Locale:           u.Locale,
		UTCOffset:        u.UtcOffset,
		IsActive:         u.IsActive,
		CreatedDate:      u.CreatedDate,
		LastModifiedDate: u.LastModifiedDate,
	}
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*domain.SalesforceUser, error) {
	u, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapUser(u), nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*domain.SalesforceUser, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapUser(u), nil
}

func (r *Repository) GetDefaultUser(ctx context.Context) (*domain.SalesforceUser, error) {
	u, err := r.queries.GetDefaultUser(ctx)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapUser(u), nil
}

func (r *Repository) CreateAuthCode(ctx context.Context, code *domain.OAuthAuthCode) (*domain.OAuthAuthCode, error) {
	created, err := r.queries.CreateAuthCode(ctx, gen.CreateAuthCodeParams{
		Code:        code.Code,
		ClientID:    code.ClientID,
		UserID:      code.UserID,
		RedirectUri: code.RedirectURI,
		Scope:       code.Scope,
		ExpiresAt:   code.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return &domain.OAuthAuthCode{
		Code:        created.Code,
		ClientID:    created.ClientID,
		UserID:      created.UserID,
		RedirectURI: created.RedirectUri,
		Scope:       created.Scope,
		IsUsed:      created.IsUsed,
		ExpiresAt:   created.ExpiresAt,
	}, nil
}

func (r *Repository) GetAuthCode(ctx context.Context, code string) (*domain.OAuthAuthCode, error) {
	c, err := r.queries.GetAuthCode(ctx, code)
	if err != nil {
		return nil, domain.ErrInvalidGrant
	}
	return &domain.OAuthAuthCode{
		Code:        c.Code,
		ClientID:    c.ClientID,
		UserID:      c.UserID,
		RedirectURI: c.RedirectUri,
		Scope:       c.Scope,
		IsUsed:      c.IsUsed,
		ExpiresAt:   c.ExpiresAt,
	}, nil
}

func (r *Repository) MarkAuthCodeUsed(ctx context.Context, code string) error {
	return r.queries.MarkAuthCodeUsed(ctx, code)
}

func mapToken(t gen.SalesforceOauthToken) *domain.OAuthToken {
	rt := ""
	if t.RefreshToken != nil {
		rt = *t.RefreshToken
	}
	return &domain.OAuthToken{
		ID:           t.ID,
		AccessToken:  t.AccessToken,
		RefreshToken: rt,
		TokenType:    t.TokenType,
		ClientID:     t.ClientID,
		UserID:       t.UserID,
		Scope:        t.Scope,
		InstanceURL:  t.InstanceUrl,
		IDURL:        t.IDUrl,
		Signature:    t.Signature,
		IssuedAt:     t.IssuedAt,
		ExpiresAt:    t.ExpiresAt,
		IsRevoked:    t.IsRevoked,
	}
}

func (r *Repository) CreateToken(ctx context.Context, token *domain.OAuthToken) (*domain.OAuthToken, error) {
	var rtPtr *string
	if token.RefreshToken != "" {
		rt := token.RefreshToken
		rtPtr = &rt
	}
	created, err := r.queries.CreateOAuthToken(ctx, gen.CreateOAuthTokenParams{
		AccessToken:  token.AccessToken,
		RefreshToken: rtPtr,
		TokenType:    token.TokenType,
		ClientID:     token.ClientID,
		UserID:       token.UserID,
		Scope:        token.Scope,
		InstanceUrl:  token.InstanceURL,
		IDUrl:        token.IDURL,
		Signature:    token.Signature,
		IssuedAt:     token.IssuedAt,
		ExpiresAt:    token.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return mapToken(created), nil
}

func (r *Repository) GetTokenByAccess(ctx context.Context, accessToken string) (*domain.OAuthToken, error) {
	t, err := r.queries.GetOAuthTokenByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	return mapToken(t), nil
}

func (r *Repository) GetTokenByRefresh(ctx context.Context, refreshToken string) (*domain.OAuthToken, error) {
	t, err := r.queries.GetOAuthTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidGrant
	}
	return mapToken(t), nil
}

func (r *Repository) GetTokenByAny(ctx context.Context, token string) (*domain.OAuthToken, error) {
	t, err := r.queries.GetOAuthTokenByAnyToken(ctx, token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	return mapToken(t), nil
}

func (r *Repository) RevokeToken(ctx context.Context, token string) error {
	return r.queries.RevokeOAuthToken(ctx, token)
}
