package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"

	goredis "github.com/redis/go-redis/v9"
)

type RoomEngineService struct {
	gameRepository  *repopg.GameRepository
	redis           *goredis.Client
	tickInterval    time.Duration
	playRoomService *PlayRoomService
	walletService   *WalletService
	broker          *realtime.Broker
}

func NewRoomEngineService(
	gameRepository *repopg.GameRepository,
	redisClient *goredis.Client,
	playRoomService *PlayRoomService,
	walletService *WalletService,
	broker *realtime.Broker,
	tickInterval time.Duration,
) *RoomEngineService {
	if tickInterval <= 0 {
		tickInterval = time.Second
	}
	return &RoomEngineService{
		gameRepository:  gameRepository,
		redis:           redisClient,
		tickInterval:    tickInterval,
		playRoomService: playRoomService,
		walletService:   walletService,
		broker:          broker,
	}
}

func (s *RoomEngineService) Run(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[engine][panic] stage=run panic=%v stack=%s", r, string(debug.Stack()))
		}
	}()

	if err := s.runTick(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
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
			if err := s.runTick(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Printf("[engine] tick lỗi: %v", err)
				}
			}
		}
	}
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
		acquired, err := s.acquireLock(ctx, lockKey, 3*time.Second)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] không lock được room %s: %v", room.Code, err)
			continue
		}
		if !acquired {
			continue
		}

		createdPeriods, err := s.gameRepository.EnsureRoomPeriods(ctx, room, now)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] ensure period lỗi room=%s err=%v", room.Code, err)
		} else if len(createdPeriods) > 0 {
			if err := s.refreshRoomState(ctx, room.Code, "period.created"); err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
			}
		}
		s.releaseLock(ctx, lockKey)
	}

	openedRooms, err := s.gameRepository.MoveScheduledToOpen(ctx, now)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		log.Printf("[engine] chuyển SCHEDULED->OPEN lỗi: %v", err)
	} else {
		for _, roomCode := range openedRooms {
			if err := s.refreshRoomState(ctx, roomCode, "period.opened"); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("[engine][room.refresh.error] room_code=%s source=period.opened err=%v", roomCode, err)
			}
		}
	}
	lockedRooms, err := s.gameRepository.MoveOpenToLocked(ctx, now)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		log.Printf("[engine] chuyển OPEN->LOCKED lỗi: %v", err)
	} else {
		for _, roomCode := range lockedRooms {
			if err := s.refreshRoomState(ctx, roomCode, "period.locked"); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("[engine][room.refresh.error] room_code=%s source=period.locked err=%v", roomCode, err)
			}
		}
	}

	lockedPeriods, err := s.gameRepository.ListLockedPeriodsForDraw(ctx, now, 200)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return err
	}
	for _, period := range lockedPeriods {
		lockKey := fmt.Sprintf("engine:period:draw:%d", period.ID)
		acquired, err := s.acquireLock(ctx, lockKey, 5*time.Second)
		if err != nil || !acquired {
			if errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}

		draw, err := s.generateDraw(period)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] sinh kết quả lỗi period=%d err=%v", period.ID, err)
			s.releaseLock(ctx, lockKey)
			continue
		}

		if err := s.gameRepository.MarkPeriodDrawn(ctx, period, draw); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] đánh dấu DRAWN lỗi period=%d err=%v", period.ID, err)
		} else if err := s.refreshRoomState(ctx, period.RoomCode, "period.drawn"); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[engine][room.refresh.error] room_code=%s source=period.drawn err=%v", period.RoomCode, err)
		}
		s.releaseLock(ctx, lockKey)
	}

	drawnPeriods, err := s.gameRepository.ListDrawnPeriodsForSettlement(ctx, 200)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return err
	}
	for _, period := range drawnPeriods {
		lockKey := fmt.Sprintf("engine:period:settle:%d", period.ID)
		acquired, err := s.acquireLock(ctx, lockKey, 5*time.Second)
		if err != nil || !acquired {
			if errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		userIDs, err := s.gameRepository.SettlePeriod(ctx, period)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] settlement lỗi period=%d err=%v", period.ID, err)
		} else {
			if err := s.refreshRoomState(ctx, period.RoomCode, "period.settled"); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("[engine][room.refresh.error] room_code=%s source=period.settled err=%v", period.RoomCode, err)
			}
			for _, userID := range userIDs {
				if err := s.publishWalletSummary(ctx, userID, "period.settled"); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("[engine][wallet.refresh.error] user_id=%d source=period.settled err=%v", userID, err)
				}
				// Publish bets update event
				if err := s.publishBetsUpdate(ctx, period.RoomCode, userID); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("[engine][bets.update.error] room_code=%s user_id=%d err=%v", period.RoomCode, userID, err)
				}
			}
		}
		s.releaseLock(ctx, lockKey)
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

func (s *RoomEngineService) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.redis.SetNX(ctx, key, "1", ttl).Result()
}

func (s *RoomEngineService) releaseLock(ctx context.Context, key string) {
	_, _ = s.redis.Del(ctx, key).Result()
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

	sum := d1 + d2 + d3
	bigSmall := "small"
	if sum >= 11 {
		bigSmall = "big"
	}
	oddEven := "even"
	if sum%2 != 0 {
		oddEven = "odd"
	}
	isTriple := d1 == d2 && d2 == d3
	tags := []string{
		fmt.Sprintf("sum_%d", sum),
		bigSmall,
		oddEven,
	}
	if isTriple {
		tags = append(tags, "triple_any")
	}

	result := fmt.Sprintf("%d-%d-%d", d1, d2, d3)
	payload, _ := json.Marshal(map[string]any{
		"game_type":    "k3",
		"dice":         []int{d1, d2, d3},
		"sum":          sum,
		"result":       result,
		"big_small":    bigSmall,
		"odd_even":     oddEven,
		"is_triple":    isTriple,
		"tags":         tags,
		"generated_at": clock.Now(),
	})

	return repopg.DrawResult{
		Result:      result,
		BigSmall:    bigSmall,
		Color:       "-",
		PayloadJSON: payload,
	}
}

func generateLotteryDraw() repopg.DrawResult {
	rng := rand.New(rand.NewSource(clock.Now().UnixNano()))
	digits := make([]int, 5)
	sum := 0
	for i := 0; i < 5; i++ {
		digits[i] = rng.Intn(10)
		sum += digits[i]
	}

	last := digits[4]
	lastBigSmall := "small"
	if last >= 5 {
		lastBigSmall = "big"
	}
	lastOddEven := "even"
	if last%2 != 0 {
		lastOddEven = "odd"
	}
	sumBigSmall := "small"
	if sum >= 23 {
		sumBigSmall = "big"
	}
	sumOddEven := "even"
	if sum%2 != 0 {
		sumOddEven = "odd"
	}

	builder := strings.Builder{}
	for _, digit := range digits {
		builder.WriteString(strconv.Itoa(digit))
	}
	result := builder.String()
	tags := []string{
		fmt.Sprintf("pick5_%s", result),
		fmt.Sprintf("sum_%d", sum),
		fmt.Sprintf("last_%d", last),
		lastBigSmall,
		lastOddEven,
		fmt.Sprintf("sum_%s", sumBigSmall),
		fmt.Sprintf("sum_%s", sumOddEven),
	}
	positionPayload := make(map[string]map[string]any, len(digits))
	for index, digit := range digits {
		position := string(rune('a' + index))
		positionBigSmall := "small"
		if digit >= 5 {
			positionBigSmall = "big"
		}
		positionOddEven := "even"
		if digit%2 != 0 {
			positionOddEven = "odd"
		}

		tags = append(tags,
			fmt.Sprintf("pos_%s_%d", position, digit),
			fmt.Sprintf("pos_%s_%s", position, positionBigSmall),
			fmt.Sprintf("pos_%s_%s", position, positionOddEven),
		)
		positionPayload[strings.ToUpper(position)] = map[string]any{
			"digit":     digit,
			"big_small": positionBigSmall,
			"odd_even":  positionOddEven,
		}
	}

	payload, _ := json.Marshal(map[string]any{
		"game_type":     "lottery",
		"digits":        digits,
		"positions":     positionPayload,
		"sum":           sum,
		"sum_big_small": sumBigSmall,
		"sum_odd_even":  sumOddEven,
		"last_digit":    last,
		"result":        result,
		"big_small":     lastBigSmall,
		"odd_even":      lastOddEven,
		"tags":          tags,
		"generated_at":  clock.Now(),
	})

	return repopg.DrawResult{
		Result:      result,
		BigSmall:    lastBigSmall,
		Color:       "-",
		PayloadJSON: payload,
	}
}
