package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/domain/vo"
	authmodel "github.com/cia-da-vacina/crm/backend/internal/module/auth/model"
	usermodel "github.com/cia-da-vacina/crm/backend/internal/module/user/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/jwt"
	"github.com/google/uuid"
)

type LoginInput struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
}

type RefreshInput struct {
	RefreshToken string
	IP           string
	UserAgent    string
}

// LoginOutput.User propositalmente não inclui unit_ids — o contrato define
// que o payload de login não carrega unidades; o BFF chama GET /me em
// seguida para obtê-las (docs/BACKEND-CONTRACT.md §1).
type LoginOutput struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
	User         usermodel.User `json:"user"`
}

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (entity.User, error)
	GetUserByID(ctx context.Context, id string) (entity.User, error)
	GetUnitIDsByUserID(ctx context.Context, userID string) ([]string, error)
	CreateRefreshToken(ctx context.Context, rt entity.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
}

type UseCase struct {
	repo Repository
	jwt  *jwt.Service
}

func New(repo Repository, jwt *jwt.Service) *UseCase {
	return &UseCase{repo: repo, jwt: jwt}
}

func (uc *UseCase) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	user, err := uc.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return LoginOutput{}, apperrors.NewUnauthorizedError("")
		}
		return LoginOutput{}, apperrors.NewDatabaseError(err)
	}

	if !vo.VerifyPassword(input.Password, user.PasswordHash, vo.DefaultPasswordConfig()) {
		return LoginOutput{}, apperrors.NewUnauthorizedError("")
	}

	if !user.Active {
		return LoginOutput{}, apperrors.NewForbiddenError("user is inactive")
	}

	return uc.issueSession(ctx, user, input.IP, input.UserAgent)
}

func (uc *UseCase) Refresh(ctx context.Context, input RefreshInput) (LoginOutput, error) {
	rt, err := uc.repo.GetRefreshTokenByHash(ctx, hashToken(input.RefreshToken))
	if err != nil {
		return LoginOutput{}, apperrors.NewUnauthorizedError("")
	}

	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		return LoginOutput{}, apperrors.NewUnauthorizedError("")
	}

	// Revoga antes de emitir o novo par — qualquer erro aqui aborta o refresh
	// em vez de arriscar dois refresh tokens válidos simultâneos.
	if err := uc.repo.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return LoginOutput{}, err
	}

	user, err := uc.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return LoginOutput{}, apperrors.NewUnauthorizedError("")
		}
		return LoginOutput{}, apperrors.NewDatabaseError(err)
	}

	if !user.Active {
		return LoginOutput{}, apperrors.NewUnauthorizedError("")
	}

	return uc.issueSession(ctx, user, input.IP, input.UserAgent)
}

// Logout revoga todas as sessões (refresh tokens) do usuário autenticado.
// Ver nota em repository.RevokeAllUserRefreshTokens sobre por que é "todas"
// e não uma sessão específica. Idempotente: chamar de novo não é erro.
func (uc *UseCase) Logout(ctx context.Context, userID string) error {
	return uc.repo.RevokeAllUserRefreshTokens(ctx, userID)
}

func (uc *UseCase) issueSession(ctx context.Context, user entity.User, ip, userAgent string) (LoginOutput, error) {
	unitIDs, err := uc.repo.GetUnitIDsByUserID(ctx, user.ID)
	if err != nil {
		return LoginOutput{}, apperrors.NewDatabaseError(err)
	}

	accessToken, expiresIn, err := uc.signUserToken(user, unitIDs)
	if err != nil {
		return LoginOutput{}, err
	}

	rawToken, rt, err := uc.newRefreshToken(user.ID, ip, userAgent)
	if err != nil {
		return LoginOutput{}, err
	}

	if err := uc.repo.CreateRefreshToken(ctx, rt); err != nil {
		return LoginOutput{}, apperrors.NewDatabaseError(err)
	}

	return LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		ExpiresIn:    expiresIn,
		User: usermodel.User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      string(user.Role),
			Active:    user.Active,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (uc *UseCase) signUserToken(user entity.User, unitIDs []string) (string, int64, error) {
	expiration := uc.jwt.Expiration()
	reg, err := uc.jwt.NewRegisteredClaims(user.ID, expiration)
	if err != nil {
		return "", 0, err
	}
	claims := authmodel.UserClaims{
		RegisteredClaims: reg,
		Role:             string(user.Role),
		UnitIDs:          unitIDs,
	}
	signed, err := uc.jwt.Sign(claims)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(expiration.Seconds()), nil
}

func (uc *UseCase) newRefreshToken(userID, ip, userAgent string) (string, entity.RefreshToken, error) {
	rawToken, err := generateOpaqueToken()
	if err != nil {
		return "", entity.RefreshToken{}, err
	}

	rt := entity.RefreshToken{
		ID:        uuid.Must(uuid.NewV7()).String(),
		UserID:    userID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(uc.jwt.RefreshExpiration()),
		IP:        ip,
		UserAgent: userAgent,
	}

	return rawToken, rt, nil
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
