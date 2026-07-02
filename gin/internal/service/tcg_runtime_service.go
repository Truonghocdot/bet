package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gin/internal/domain/auth"
	"gin/internal/domain/game"
	"gin/internal/domain/user"
	gateclient "gin/internal/integration/gate"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/id"
	goredis "github.com/redis/go-redis/v9"
)

const tcgActiveProductTTL = 24 * time.Hour

type TCGRuntimeConfig struct {
	Enabled                     bool
	DefaultLanguage             string
	DefaultGameMode             string
	DefaultIPAddress            string
	DefaultWebPlatform          string
	DefaultMobilePlatform       string
	DefaultLotteryLobbyGameCode string
	LaunchReturnURL             string
	DefaultCurrency             string
}

type TCGLaunchRequest struct {
	ProductType int    `json:"product_type"`
	GameType    string `json:"game_type"`
	GameCode    string `json:"game_code"`
	Name        string `json:"name"`
}

type TCGLaunchResponse struct {
	Message           string `json:"message"`
	GameURL           string `json:"game_url"`
	ProductType       int    `json:"product_type"`
	GameType          string `json:"game_type"`
	TransferredAmount string `json:"transferred_amount"`
	Source            string `json:"source"`
}

type TCGCloseActiveResponse struct {
	Message     string `json:"message"`
	ProductType int    `json:"product_type"`
	GameType    string `json:"game_type"`
	Swept       bool   `json:"swept"`
	ReferenceNo string `json:"reference_no,omitempty"`
}

type tcgActiveProductState struct {
	ProductType     int       `json:"product_type"`
	GameType        string    `json:"game_type"`
	LastTransferRef string    `json:"last_transfer_ref,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TCGRuntimeService struct {
	userRepository   *repopg.UserRepository
	walletRepository *repopg.WalletRepository
	walletService    *WalletService
	tcg              *gateclient.TCGClient
	redis            *goredis.Client
	config           TCGRuntimeConfig
}

func NewTCGRuntimeService(
	userRepository *repopg.UserRepository,
	walletRepository *repopg.WalletRepository,
	walletService *WalletService,
	tcg *gateclient.TCGClient,
	redis *goredis.Client,
	config TCGRuntimeConfig,
) *TCGRuntimeService {
	return &TCGRuntimeService{
		userRepository:   userRepository,
		walletRepository: walletRepository,
		walletService:    walletService,
		tcg:              tcg,
		redis:            redis,
		config:           config,
	}
}

func (s *TCGRuntimeService) LaunchGame(ctx context.Context, userID int64, request TCGLaunchRequest, meta game.ProviderGameLaunchMeta) (TCGLaunchResponse, error) {
	if !s.config.Enabled || s.tcg == nil {
		return TCGLaunchResponse{}, fmt.Errorf("tcg launch chưa sẵn sàng")
	}
	if userID == 0 {
		return TCGLaunchResponse{}, ErrUnauthorized
	}
	if request.ProductType <= 0 {
		return TCGLaunchResponse{}, fmt.Errorf("thiếu product_type")
	}
	if strings.TrimSpace(request.GameType) == "" {
		return TCGLaunchResponse{}, fmt.Errorf("thiếu game_type")
	}
	if strings.TrimSpace(request.GameCode) == "" {
		return TCGLaunchResponse{}, fmt.Errorf("thiếu game_code")
	}

	profile, err := s.userRepository.FindProfileByUserID(ctx, userID)
	if err != nil {
		return TCGLaunchResponse{}, err
	}

	username := buildTCGUsername(profile.User.Name, profile.User.Phone)
	if username == "" {
		return TCGLaunchResponse{}, fmt.Errorf("không thể tạo username tcg")
	}
	if err := s.ensureTCGPlayer(ctx, profile, username); err != nil {
		return TCGLaunchResponse{}, err
	}

	activeState, _ := s.loadActiveState(ctx, userID)
	if activeState != nil && activeState.ProductType != request.ProductType {
		if _, err := s.closeActiveProduct(ctx, userID, username, *activeState); err != nil {
			return TCGLaunchResponse{}, err
		}
		activeState = nil
	}

	transferredAmount := "0"
	if activeState == nil || activeState.ProductType != request.ProductType {
		wallet, err := s.walletRepository.FindByUserAndUnit(ctx, userID, user.WalletUnitVND)
		if err != nil {
			return TCGLaunchResponse{}, err
		}

		if isPositiveNumeric(wallet.Balance) {
			referenceNo := buildTCGReference("FT")
			transferResponse, transferErr := s.tcg.Transfer(ctx, gateclient.TransferTCGRequest{
				Username:    username,
				ProductType: request.ProductType,
				FundType:    "1",
				Amount:      wallet.Balance,
				ReferenceNo: referenceNo,
			})
			if transferErr != nil || transferResponse.Status != 0 {
				statusResponse, statusErr := s.tcg.CheckTransferStatus(ctx, gateclient.TransferStatusTCGRequest{
					ProductType: request.ProductType,
					ReferenceNo: referenceNo,
				})
				if statusErr != nil || strings.ToUpper(strings.TrimSpace(statusResponse.TransactionStatus)) != "SUCCESS" {
					if transferErr != nil {
						return TCGLaunchResponse{}, transferErr
					}
					if strings.TrimSpace(transferResponse.ErrorDesc) != "" {
						return TCGLaunchResponse{}, fmt.Errorf("%s", transferResponse.ErrorDesc)
					}
					if strings.TrimSpace(statusResponse.ErrorDesc) != "" {
						return TCGLaunchResponse{}, fmt.Errorf("%s", statusResponse.ErrorDesc)
					}
					return TCGLaunchResponse{}, fmt.Errorf("không thể chuyển tiền vào ví game")
				}
			}
			transferredAmount = wallet.Balance
			if err := s.walletRepository.DebitByUserAndUnit(ctx, userID, user.WalletUnitVND, wallet.Balance, "tcg_transfer", fmt.Sprintf("Chuyển sang TCG product %d", request.ProductType)); err != nil {
				return TCGLaunchResponse{}, err
			}
			activeState = &tcgActiveProductState{
				ProductType:     request.ProductType,
				GameType:        strings.TrimSpace(request.GameType),
				LastTransferRef: referenceNo,
				UpdatedAt:       clock.Now(),
			}
			_ = s.saveActiveState(ctx, userID, *activeState)
			if s.walletService != nil {
				_ = s.walletService.PublishSummary(ctx, userID)
			}
		}
	}

	launchResponse, launchErr := s.tcg.LaunchGame(ctx, gateclient.LaunchGameTCGRequest{
		Username:    username,
		ProductType: request.ProductType,
		IPAddress:   resolveTCGIPAddress(meta.IP, s.config.DefaultIPAddress),
		Platform:    resolveTCGPlatform(meta.UserAgent, s.config.DefaultMobilePlatform, s.config.DefaultWebPlatform),
		GameMode:    defaultString(strings.TrimSpace(s.config.DefaultGameMode), "1"),
		GameCode:    strings.TrimSpace(request.GameCode),
		Language:    strings.TrimSpace(s.config.DefaultLanguage),
		Nickname:    strings.TrimSpace(profile.User.Name),
		BackURL:     strings.TrimSpace(s.config.LaunchReturnURL),
	})
	if launchErr != nil {
		if activeState != nil {
			activeState.ProductType = request.ProductType
			activeState.GameType = strings.TrimSpace(request.GameType)
			activeState.UpdatedAt = clock.Now()
			_ = s.saveActiveState(ctx, userID, *activeState)
		}
		if strings.TrimSpace(launchResponse.ErrorDesc) != "" {
			return TCGLaunchResponse{}, fmt.Errorf("%s", launchResponse.ErrorDesc)
		}
		return TCGLaunchResponse{}, launchErr
	}
	if strings.TrimSpace(launchResponse.GameURL) == "" {
		return TCGLaunchResponse{}, fmt.Errorf("tcg không trả game_url")
	}

	if activeState == nil {
		activeState = &tcgActiveProductState{
			ProductType: request.ProductType,
			GameType:    strings.TrimSpace(request.GameType),
			UpdatedAt:   clock.Now(),
		}
	}
	activeState.ProductType = request.ProductType
	activeState.GameType = strings.TrimSpace(request.GameType)
	activeState.UpdatedAt = clock.Now()
	_ = s.saveActiveState(ctx, userID, *activeState)

	return TCGLaunchResponse{
		Message:           "Khởi động game thành công",
		GameURL:           strings.TrimSpace(launchResponse.GameURL),
		ProductType:       request.ProductType,
		GameType:          strings.TrimSpace(request.GameType),
		TransferredAmount: transferredAmount,
		Source:            "tcg_api",
	}, nil
}

func (s *TCGRuntimeService) CloseActive(ctx context.Context, userID int64) (TCGCloseActiveResponse, error) {
	if !s.config.Enabled || s.tcg == nil {
		return TCGCloseActiveResponse{Message: "TCG chưa bật", Swept: false}, nil
	}
	if userID == 0 {
		return TCGCloseActiveResponse{}, ErrUnauthorized
	}

	state, ok := s.loadActiveState(ctx, userID)
	if !ok || state == nil {
		return TCGCloseActiveResponse{Message: "Không có product đang hoạt động", Swept: false}, nil
	}

	profile, err := s.userRepository.FindProfileByUserID(ctx, userID)
	if err != nil {
		return TCGCloseActiveResponse{}, err
	}

	username := buildTCGUsername(profile.User.Name, profile.User.Phone)
	if username == "" {
		return TCGCloseActiveResponse{}, fmt.Errorf("không thể tạo username tcg")
	}
	if err := s.ensureTCGPlayer(ctx, profile, username); err != nil {
		return TCGCloseActiveResponse{}, err
	}

	return s.closeActiveProduct(ctx, userID, username, *state)
}

func (s *TCGRuntimeService) closeActiveProduct(ctx context.Context, userID int64, username string, state tcgActiveProductState) (TCGCloseActiveResponse, error) {
	balanceBeforeSweep := "0"
	if balanceResponse, balanceErr := s.tcg.GetBalance(ctx, gateclient.BalanceTCGRequest{
		Username:    username,
		ProductType: state.ProductType,
	}); balanceErr == nil {
		balanceBeforeSweep = strings.TrimSpace(balanceResponse.Balance)
	}

	referenceNo := buildTCGReference("FOA")
	response, err := s.tcg.TransferOutAll(ctx, gateclient.TransferTCGRequest{
		Username:    username,
		ProductType: state.ProductType,
		ReferenceNo: referenceNo,
	})
	if err != nil || response.Status != 0 {
		statusResponse, statusErr := s.tcg.CheckTransferStatus(ctx, gateclient.TransferStatusTCGRequest{
			ProductType: state.ProductType,
			ReferenceNo: referenceNo,
		})
		if statusErr != nil || strings.ToUpper(strings.TrimSpace(statusResponse.TransactionStatus)) != "SUCCESS" {
			if err != nil {
				return TCGCloseActiveResponse{}, err
			}
			if strings.TrimSpace(response.ErrorDesc) != "" {
				return TCGCloseActiveResponse{}, fmt.Errorf("%s", response.ErrorDesc)
			}
			if strings.TrimSpace(statusResponse.ErrorDesc) != "" {
				return TCGCloseActiveResponse{}, fmt.Errorf("%s", statusResponse.ErrorDesc)
			}
			return TCGCloseActiveResponse{}, fmt.Errorf("không thể rút tiền khỏi ví game")
		}
	}

	_ = s.deleteActiveState(ctx, userID)
	if isPositiveNumeric(balanceBeforeSweep) {
		if err := s.walletRepository.CreditByUserAndUnit(ctx, userID, user.WalletUnitVND, balanceBeforeSweep, "tcg_transfer_out_all", fmt.Sprintf("Rút từ TCG product %d", state.ProductType)); err != nil {
			return TCGCloseActiveResponse{}, err
		}
	}
	if s.walletService != nil {
		_ = s.walletService.PublishSummary(ctx, userID)
	}
	return TCGCloseActiveResponse{
		Message:     "Đã đóng product đang hoạt động",
		ProductType: state.ProductType,
		GameType:    state.GameType,
		Swept:       true,
		ReferenceNo: referenceNo,
	}, nil
}

func (s *TCGRuntimeService) loadActiveState(ctx context.Context, userID int64) (*tcgActiveProductState, bool) {
	if s.redis == nil {
		return nil, false
	}
	raw, err := s.redis.Get(ctx, s.activeStateKey(userID)).Result()
	if err != nil {
		return nil, false
	}

	var state tcgActiveProductState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, false
	}
	return &state, true
}

func (s *TCGRuntimeService) saveActiveState(ctx context.Context, userID int64, state tcgActiveProductState) error {
	if s.redis == nil {
		return nil
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, s.activeStateKey(userID), payload, tcgActiveProductTTL).Err()
}

func (s *TCGRuntimeService) deleteActiveState(ctx context.Context, userID int64) error {
	if s.redis == nil {
		return nil
	}
	return s.redis.Del(ctx, s.activeStateKey(userID)).Err()
}

func (s *TCGRuntimeService) activeStateKey(userID int64) string {
	return fmt.Sprintf("tcg:active-product:%d", userID)
}

func buildTCGReference(prefix string) string {
	return fmt.Sprintf("TCG%s%s", strings.ToUpper(strings.TrimSpace(prefix)), id.New())
}

func (s *TCGRuntimeService) ensureTCGPlayer(ctx context.Context, profile auth.UserProfile, username string) error {
	return s.tcg.RegisterPlayer(ctx, gateclient.RegisterTCGPlayerRequest{
		Username: strings.TrimSpace(username),
		Password: buildTCGPassword("", username),
		Currency: strings.TrimSpace(s.config.DefaultCurrency),
	})
}

func resolveTCGIPAddress(ip string, fallback string) string {
	trimmed := strings.TrimSpace(ip)
	if trimmed != "" {
		return trimmed
	}
	return defaultString(strings.TrimSpace(fallback), "127.0.0.1")
}

func resolveTCGPlatform(userAgent string, mobile string, desktop string) string {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if strings.Contains(ua, "android") || strings.Contains(ua, "iphone") || strings.Contains(ua, "mobile") || strings.Contains(ua, "ipad") {
		return defaultString(strings.TrimSpace(mobile), "html5")
	}
	return defaultString(strings.TrimSpace(desktop), "html5-desktop")
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func isPositiveNumeric(value string) bool {
	rat := new(big.Rat)
	if _, ok := rat.SetString(strings.TrimSpace(value)); !ok {
		return false
	}
	return rat.Cmp(new(big.Rat)) > 0
}
