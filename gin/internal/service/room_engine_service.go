package service

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"

	goredis "github.com/redis/go-redis/v9"
)

var releaseRoomEngineLockScript = goredis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	end
	return 0
`)

func isContextDone(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

type RoomEngineService struct {
	gameRepository    *repopg.GameRepository
	redis             *goredis.Client
	tickInterval      time.Duration
	settlementEnabled bool
	playRoomService   *PlayRoomService
	walletService     *WalletService
	broker            *realtime.Broker
}

func NewRoomEngineService(
	gameRepository *repopg.GameRepository,
	redisClient *goredis.Client,
	playRoomService *PlayRoomService,
	walletService *WalletService,
	broker *realtime.Broker,
	tickInterval time.Duration,
	settlementEnabled bool,
) *RoomEngineService {
	if tickInterval <= 0 {
		tickInterval = time.Second
	}
	return &RoomEngineService{
		gameRepository:    gameRepository,
		redis:             redisClient,
		tickInterval:      tickInterval,
		settlementEnabled: settlementEnabled,
		playRoomService:   playRoomService,
		walletService:     walletService,
		broker:            broker,
	}
}

func (s *RoomEngineService) Run(ctx context.Context) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("[engine][panic] stage=run panic=%v stack=%s", recovered, string(debug.Stack()))
			err = fmt.Errorf("room engine panic: %v", recovered)
		}
	}()
	if !s.settlementEnabled {
		log.Printf("[engine][settlement.disabled]")
	}

	if err := s.runTickSafely(ctx); err != nil {
		if !isContextDone(err) {
			log.Printf("[engine] tick lỗi ban đầu: %v", err)
		}
	}

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[engine][stop] reason=context_canceled")
			return nil
		case <-ticker.C:
			if err := s.runTickSafely(ctx); err != nil {
				if !isContextDone(err) {
					log.Printf("[engine] tick lỗi: %v", err)
				}
			}
		}
	}
}

func (s *RoomEngineService) runTickSafely(ctx context.Context) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("[engine][panic] stage=tick panic=%v stack=%s", recovered, string(debug.Stack()))
			err = fmt.Errorf("room engine tick panic: %v", recovered)
		}
	}()

	return s.runTick(ctx)
}

func (s *RoomEngineService) runTick(ctx context.Context) error {
	now := clock.Now()
	rooms, err := s.gameRepository.ListRooms(ctx)
	if err != nil {
		log.Printf("[engine][room.list.error] err=%v", err)
		return err
	}
	if len(rooms) == 0 {
		rooms = defaultEngineRooms()
		log.Printf("[engine] game_rooms trống, dùng catalog mặc định để bootstrap period")
	}

	for _, room := range rooms {
		lockKey := fmt.Sprintf("engine:room:ensure:%s", room.Code)
		acquired, err := s.withLock(ctx, lockKey, 3*time.Second, func() error {
			createdPeriods, ensureErr := s.gameRepository.EnsureRoomPeriods(ctx, room, now)
			if ensureErr != nil {
				if isContextDone(ensureErr) {
					return ensureErr
				}
				log.Printf("[engine] ensure period lỗi room=%s err=%v", room.Code, ensureErr)
				return nil
			}
			if len(createdPeriods) > 0 {
				if refreshErr := s.refreshRoomState(ctx, room.Code, "period.created"); refreshErr != nil {
					if isContextDone(refreshErr) {
						return refreshErr
					}
				}
			}
			return nil
		})
		if err != nil {
			if isContextDone(err) {
				return err
			}
			log.Printf("[engine] không lock được room %s: %v", room.Code, err)
			continue
		}
		if !acquired {
			continue
		}
	}

	openedRooms, err := s.gameRepository.MoveScheduledToOpen(ctx, now)
	if err != nil {
		if isContextDone(err) {
			return err
		}
		log.Printf("[engine] chuyển SCHEDULED->OPEN lỗi: %v", err)
	} else {
		for _, roomCode := range openedRooms {
			if err := s.refreshRoomState(ctx, roomCode, "period.opened"); err != nil && !isContextDone(err) {
				log.Printf("[engine][room.refresh.error] room_code=%s source=period.opened err=%v", roomCode, err)
			}
		}
	}
	lockedRooms, err := s.gameRepository.MoveOpenToLocked(ctx, now)
	if err != nil {
		if isContextDone(err) {
			return err
		}
		log.Printf("[engine] chuyển OPEN->LOCKED lỗi: %v", err)
	} else {
		for _, roomCode := range lockedRooms {
			if err := s.refreshRoomState(ctx, roomCode, "period.locked"); err != nil && !isContextDone(err) {
				log.Printf("[engine][room.refresh.error] room_code=%s source=period.locked err=%v", roomCode, err)
			}
		}
	}

	lockedPeriods, err := s.gameRepository.ListLockedPeriodsForDraw(ctx, now, 200)
	if err != nil {
		if isContextDone(err) {
			return err
		}
		return err
	}
	for _, period := range lockedPeriods {
		lockKey := fmt.Sprintf("engine:period:draw:%d", period.ID)
		acquired, err := s.withLock(ctx, lockKey, 5*time.Second, func() error {
			draw, drawErr := s.generateDraw(period)
			if drawErr != nil {
				if isContextDone(drawErr) {
					return drawErr
				}
				log.Printf("[engine] sinh kết quả lỗi period=%d err=%v", period.ID, drawErr)
				return nil
			}

			if markErr := s.gameRepository.MarkPeriodDrawn(ctx, period, draw); markErr != nil {
				if isContextDone(markErr) {
					return markErr
				}
				log.Printf("[engine] đánh dấu DRAWN lỗi period=%d err=%v", period.ID, markErr)
			} else if refreshErr := s.refreshRoomState(ctx, period.RoomCode, "period.drawn"); refreshErr != nil && !isContextDone(refreshErr) {
				log.Printf("[engine][room.refresh.error] room_code=%s source=period.drawn err=%v", period.RoomCode, refreshErr)
			} else if isContextDone(refreshErr) {
				return refreshErr
			}
			return nil
		})
		if err != nil {
			if isContextDone(err) {
				return err
			}
			log.Printf("[engine][lock.acquire.error] key=%s err=%v", lockKey, err)
			continue
		}
		if !acquired {
			continue
		}
	}

	return s.settleDrawnPeriods(ctx)
}

func (s *RoomEngineService) settleDrawnPeriods(ctx context.Context) error {
	if !s.settlementEnabled {
		return nil
	}

	drawnPeriods, err := s.gameRepository.ListDrawnPeriodsForSettlement(ctx, 200)
	if err != nil {
		if isContextDone(err) {
			return err
		}
		return err
	}
	for _, period := range drawnPeriods {
		lockKey := fmt.Sprintf("engine:period:settle:%d", period.ID)
		acquired, err := s.withLock(ctx, lockKey, 5*time.Second, func() error {
			userIDs, settleErr := s.gameRepository.SettlePeriod(ctx, period)
			if settleErr != nil {
				if isContextDone(settleErr) {
					return settleErr
				}
				log.Printf("[engine] settlement lỗi period=%d err=%v", period.ID, settleErr)
				if recordErr := s.gameRepository.RecordPeriodSettlementFailure(ctx, period.ID, settleErr); recordErr != nil {
					log.Printf("[engine][period.settle.failure_record.error] period_id=%d err=%v", period.ID, recordErr)
					if isContextDone(recordErr) {
						return recordErr
					}
				}
				return nil
			}

			if refreshErr := s.refreshRoomState(ctx, period.RoomCode, "period.settled"); refreshErr != nil {
				if isContextDone(refreshErr) {
					return refreshErr
				}
				log.Printf("[engine][room.refresh.error] room_code=%s source=period.settled err=%v", period.RoomCode, refreshErr)
			}
			for _, userID := range userIDs {
				if publishErr := s.publishWalletSummary(ctx, userID, "period.settled"); publishErr != nil {
					if isContextDone(publishErr) {
						return publishErr
					}
					log.Printf("[engine][wallet.refresh.error] user_id=%d source=period.settled err=%v", userID, publishErr)
				}
				if publishErr := s.publishBetsUpdate(ctx, period.RoomCode, userID); publishErr != nil {
					if isContextDone(publishErr) {
						return publishErr
					}
					log.Printf("[engine][bets.update.error] room_code=%s user_id=%d err=%v", period.RoomCode, userID, publishErr)
				}
			}
			return nil
		})
		if err != nil {
			if isContextDone(err) {
				return err
			}
			log.Printf("[engine][lock.acquire.error] key=%s err=%v", lockKey, err)
			continue
		}
		if !acquired {
			continue
		}
	}

	return nil
}

func defaultEngineRooms() []repopg.GameRoomRecord {
	return []repopg.GameRoomRecord{
		{Code: "wingo_30s", GameType: 1, DurationSeconds: 30, BetCutoffSeconds: 5, Status: 1, SortOrder: 1},
		{Code: "wingo_1m", GameType: 1, DurationSeconds: 60, BetCutoffSeconds: 5, Status: 1, SortOrder: 2},
		{Code: "wingo_3m", GameType: 1, DurationSeconds: 180, BetCutoffSeconds: 5, Status: 1, SortOrder: 3},
		{Code: "wingo_5m", GameType: 1, DurationSeconds: 300, BetCutoffSeconds: 5, Status: 1, SortOrder: 4},
		{Code: "k3_1m", GameType: 2, DurationSeconds: 60, BetCutoffSeconds: 5, Status: 1, SortOrder: 5},
		{Code: "k3_3m", GameType: 2, DurationSeconds: 180, BetCutoffSeconds: 5, Status: 1, SortOrder: 6},
		{Code: "k3_5m", GameType: 2, DurationSeconds: 300, BetCutoffSeconds: 5, Status: 1, SortOrder: 7},
		{Code: "k3_10m", GameType: 2, DurationSeconds: 600, BetCutoffSeconds: 5, Status: 1, SortOrder: 8},
		{Code: "lottery_1m", GameType: 3, DurationSeconds: 60, BetCutoffSeconds: 5, Status: 1, SortOrder: 9},
		{Code: "lottery_3m", GameType: 3, DurationSeconds: 180, BetCutoffSeconds: 5, Status: 1, SortOrder: 10},
		{Code: "lottery_5m", GameType: 3, DurationSeconds: 300, BetCutoffSeconds: 5, Status: 1, SortOrder: 11},
		{Code: "lottery_10m", GameType: 3, DurationSeconds: 600, BetCutoffSeconds: 5, Status: 1, SortOrder: 12},
	}
}

func (s *RoomEngineService) refreshRoomState(ctx context.Context, roomCode string, source string) error {
	if s.playRoomService == nil || strings.TrimSpace(roomCode) == "" {
		return nil
	}
	_, err := s.playRoomService.RefreshRoomState(ctx, roomCode, source)
	return err
}

func (s *RoomEngineService) publishWalletSummary(ctx context.Context, userID int64, source string) error {
	if s.walletService == nil || userID == 0 {
		return nil
	}
	if err := s.walletService.PublishSummary(ctx, userID); err != nil {
		log.Printf("[realtime][wallet.publish.error] user_id=%d source=%s err=%v", userID, source, err)
		return err
	}
	return nil
}

func (s *RoomEngineService) publishBetsUpdate(ctx context.Context, roomCode string, userID int64) error {
	if s.broker == nil || roomCode == "" || userID == 0 {
		return nil
	}

	// Publish event để notify client bets được updated
	topic := realtime.PlayRoomBetsTopic(roomCode, userID)
	payload := map[string]any{
		"type":    "settlement",
		"message": "Bets have been updated",
	}

	if err := s.broker.Publish(ctx, topic, "bets.updated", payload); err != nil {
		log.Printf("[realtime][bets.update.error] room_code=%s user_id=%d err=%v", roomCode, userID, err)
		return err
	}
	return nil
}

func (s *RoomEngineService) withLock(ctx context.Context, key string, ttl time.Duration, operation func() error) (bool, error) {
	token, acquired, err := s.acquireLock(ctx, key, ttl)
	if err != nil || !acquired {
		return acquired, err
	}
	defer s.releaseLock(ctx, key, token)

	return true, operation()
}

func (s *RoomEngineService) acquireLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	tokenBytes := make([]byte, 16)
	if _, err := cryptorand.Read(tokenBytes); err != nil {
		return "", false, fmt.Errorf("generate lock token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	acquired, err := s.redis.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	return token, acquired, nil
}

func (s *RoomEngineService) releaseLock(ctx context.Context, key, token string) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(token) == "" {
		return
	}

	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	if _, err := releaseRoomEngineLockScript.Run(releaseCtx, s.redis, []string{key}, token).Result(); err != nil {
		log.Printf("[engine][lock.release.error] key=%s err=%v", key, err)
	}
}

func (s *RoomEngineService) generateDraw(period repopg.GamePeriodRecord) (repopg.DrawResult, error) {
	if len(period.ManualResultJSON) > 0 {
		var manualResult repopg.DrawResult
		if err := json.Unmarshal(period.ManualResultJSON, &manualResult); err == nil && manualResult.Result != "" {
			return manualResult, nil
		}
	}

	switch period.GameType {
	case 1:
		return generateWingoDraw(), nil
	case 2:
		return generateK3Draw(), nil
	case 3:
		return generateLotteryDraw(), nil
	default:
		return repopg.DrawResult{}, fmt.Errorf("game_type không hỗ trợ: %d", period.GameType)
	}
}

func generateWingoDraw() repopg.DrawResult {
	rng := rand.New(rand.NewSource(clock.Now().UnixNano()))
	number := rng.Intn(10)
	bigSmall := "small"
	if number >= 5 {
		bigSmall = "big"
	}
	oddEven := "even"
	if number%2 != 0 {
		oddEven = "odd"
	}
	primaryColor := "red"
	if number%2 != 0 {
		primaryColor = "green"
	}

	color := primaryColor
	tags := []string{
		fmt.Sprintf("number_%d", number),
		bigSmall,
		oddEven,
	}
	if number == 0 {
		color = "red_violet"
		tags = append(tags, "red", "violet")
	} else if number == 5 {
		color = "green_violet"
		tags = append(tags, "green", "violet")
	} else {
		tags = append(tags, primaryColor)
	}

	payload, _ := json.Marshal(map[string]any{
		"game_type":    "wingo",
		"number":       number,
		"result":       strconv.Itoa(number),
		"big_small":    bigSmall,
		"odd_even":     oddEven,
		"color":        color,
		"tags":         tags,
		"generated_at": clock.Now(),
	})

	return repopg.DrawResult{
		Result:      strconv.Itoa(number),
		BigSmall:    bigSmall,
		Color:       color,
		PayloadJSON: payload,
	}
}

func generateK3Draw() repopg.DrawResult {
	rng := rand.New(rand.NewSource(clock.Now().UnixNano()))
	d1 := rng.Intn(6) + 1
	d2 := rng.Intn(6) + 1
	d3 := rng.Intn(6) + 1

	outcome, ok := game.BuildK3Outcome([]int{d1, d2, d3})
	if !ok {
		outcome = game.K3Outcome{
			Dice:     []int{d1, d2, d3},
			Sum:      d1 + d2 + d3,
			Result:   fmt.Sprintf("%d-%d-%d", d1, d2, d3),
			BigSmall: "small",
			OddEven:  "even",
			Tags:     []string{fmt.Sprintf("sum_%d", d1+d2+d3), "small", "even"},
		}
	}
	payload, _ := json.Marshal(map[string]any{
		"game_type":    "k3",
		"dice":         outcome.Dice,
		"sum":          outcome.Sum,
		"result":       outcome.Result,
		"big_small":    outcome.BigSmall,
		"odd_even":     outcome.OddEven,
		"is_triple":    outcome.IsTriple,
		"tags":         outcome.Tags,
		"generated_at": clock.Now(),
	})

	return repopg.DrawResult{
		Result:      outcome.Result,
		BigSmall:    outcome.BigSmall,
		Color:       "-",
		PayloadJSON: payload,
	}
}

func generateLotteryDraw() repopg.DrawResult {
	rng := rand.New(rand.NewSource(clock.Now().UnixNano()))
	digits := make([]int, 5)
	for i := 0; i < 5; i++ {
		digits[i] = rng.Intn(10)
	}

	outcome, ok := game.BuildLotteryOutcome(digits)
	if !ok {
		outcome = game.LotteryOutcome{
			Digits:      digits,
			Result:      "00000",
			BigSmall:    "small",
			OddEven:     "even",
			SumBigSmall: "small",
			SumOddEven:  "even",
			Tags:        []string{"pick5_00000", "sum_0", "last_0", "small", "even", "sum_small", "sum_even"},
		}
	}
	payload, _ := json.Marshal(map[string]any{
		"game_type":     "lottery",
		"digits":        outcome.Digits,
		"positions":     outcome.Positions,
		"sum":           outcome.Sum,
		"sum_big_small": outcome.SumBigSmall,
		"sum_odd_even":  outcome.SumOddEven,
		"last_digit":    outcome.LastDigit,
		"result":        outcome.Result,
		"big_small":     outcome.BigSmall,
		"odd_even":      outcome.OddEven,
		"tags":          outcome.Tags,
		"generated_at":  clock.Now(),
	})

	return repopg.DrawResult{
		Result:      outcome.Result,
		BigSmall:    outcome.BigSmall,
		Color:       "-",
		PayloadJSON: payload,
	}
}
