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

	"gin/internal/support/clock"
	repopg "gin/internal/repository/postgres"

	goredis "github.com/redis/go-redis/v9"
)

const roomPeriodEventsChannel = "game:room:period-events:v1"

type RoomEngineService struct {
	gameRepository *repopg.GameRepository
	redis          *goredis.Client
	tickInterval   time.Duration
}

func NewRoomEngineService(gameRepository *repopg.GameRepository, redisClient *goredis.Client, tickInterval time.Duration) *RoomEngineService {
	if tickInterval <= 0 {
		tickInterval = time.Second
	}
	return &RoomEngineService{
		gameRepository: gameRepository,
		redis:          redisClient,
		tickInterval:   tickInterval,
	}
}

func (s *RoomEngineService) Run(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[engine][panic] stage=run panic=%v stack=%s", r, string(debug.Stack()))
		}
	}()

	log.Printf("[engine][start] tick_interval=%s", s.tickInterval)
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
			log.Printf("[engine][tick.start] at_vn=%s", clock.Now().Format(time.RFC3339Nano))
			if err := s.runTick(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Printf("[engine] tick lỗi: %v", err)
				}
			}
			log.Printf("[engine][tick.done] at_vn=%s", clock.Now().Format(time.RFC3339Nano))
		}
	}
}

func (s *RoomEngineService) runTick(ctx context.Context) error {
	now := clock.Now()
	log.Printf("[engine][tick.exec] now_vn=%s", now.Format(time.RFC3339Nano))
	rooms, err := s.gameRepository.ListRooms(ctx)
	if err != nil {
		log.Printf("[engine][room.list.error] err=%v", err)
		return err
	}
	log.Printf("[engine][room.list.done] count=%d", len(rooms))
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
		log.Printf("[engine][room.lock.acquired] room_code=%s key=%s", room.Code, lockKey)
		if !acquired {
			log.Printf("[engine][room.lock.skip] room_code=%s reason=already_locked", room.Code)
			continue
		}

		createdPeriods, err := s.gameRepository.EnsureRoomPeriods(ctx, room, now)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] ensure period lỗi room=%s err=%v", room.Code, err)
		} else if len(createdPeriods) > 0 {
			log.Printf("[engine] room=%s đã sinh %d kỳ mới", room.Code, len(createdPeriods))
			for _, period := range createdPeriods {
				if err := s.publishPeriodCreated(ctx, period); err != nil {
					if errors.Is(err, context.Canceled) {
						return err
					}
					log.Printf("[engine] publish period lỗi room=%s period=%s err=%v", period.RoomCode, period.PeriodNo, err)
				}
			}
		} else {
			log.Printf("[engine][period.ensure.none] room_code=%s", room.Code)
		}
		s.releaseLock(ctx, lockKey)
		log.Printf("[engine][room.lock.released] room_code=%s key=%s", room.Code, lockKey)
	}

	openedCount, err := s.gameRepository.MoveScheduledToOpen(ctx, now)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		log.Printf("[engine] chuyển SCHEDULED->OPEN lỗi: %v", err)
	} else {
		log.Printf("[engine][period.transition] from=SCHEDULED to=OPEN count=%d", openedCount)
	}
	lockedCount, err := s.gameRepository.MoveOpenToLocked(ctx, now)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		log.Printf("[engine] chuyển OPEN->LOCKED lỗi: %v", err)
	} else {
		log.Printf("[engine][period.transition] from=OPEN to=LOCKED count=%d", lockedCount)
	}

	lockedPeriods, err := s.gameRepository.ListLockedPeriodsForDraw(ctx, now, 200)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return err
	}
	log.Printf("[engine][period.draw.queue] count=%d", len(lockedPeriods))
	for _, period := range lockedPeriods {
		lockKey := fmt.Sprintf("engine:period:draw:%d", period.ID)
		acquired, err := s.acquireLock(ctx, lockKey, 5*time.Second)
		if err != nil || !acquired {
			if errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}

		draw, err := s.generateDraw(period.GameType)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] sinh kết quả lỗi period=%d err=%v", period.ID, err)
			s.releaseLock(ctx, lockKey)
			continue
		}
		log.Printf("[engine][period.draw.generated] period_id=%d room_code=%s period_no=%s result=%s", period.ID, period.RoomCode, period.PeriodNo, draw.Result)

		if err := s.gameRepository.MarkPeriodDrawn(ctx, period, draw); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] đánh dấu DRAWN lỗi period=%d err=%v", period.ID, err)
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
	log.Printf("[engine][period.settle.queue] count=%d", len(drawnPeriods))
	for _, period := range drawnPeriods {
		lockKey := fmt.Sprintf("engine:period:settle:%d", period.ID)
		acquired, err := s.acquireLock(ctx, lockKey, 5*time.Second)
		if err != nil || !acquired {
			if errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		if err := s.gameRepository.SettlePeriod(ctx, period); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("[engine] settlement lỗi period=%d err=%v", period.ID, err)
		}
		s.releaseLock(ctx, lockKey)
		log.Printf("[engine][period.settle.lock.released] period_id=%d room_code=%s", period.ID, period.RoomCode)
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

func (s *RoomEngineService) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.redis.SetNX(ctx, key, "1", ttl).Result()
}

func (s *RoomEngineService) releaseLock(ctx context.Context, key string) {
	_, _ = s.redis.Del(ctx, key).Result()
}

func (s *RoomEngineService) publishPeriodCreated(ctx context.Context, period repopg.GamePeriodRecord) error {
	if s.redis == nil {
		log.Printf("[engine][period.publish.skip] period_id=%d room_code=%s reason=redis_nil", period.ID, period.RoomCode)
		return nil
	}

	payload, err := json.Marshal(map[string]any{
		"event":       "period.created",
		"room_code":   period.RoomCode,
		"period_id":   period.ID,
		"period_no":   period.PeriodNo,
		"game_type":   period.GameType,
		"status":      period.Status,
		"open_at":     period.OpenAt,
		"bet_lock_at": period.BetLockAt,
		"draw_at":     period.DrawAt,
		"published_at": clock.Now(),
	})
	if err != nil {
		log.Printf("[engine][period.publish.error] period_id=%d room_code=%s stage=marshal err=%v", period.ID, period.RoomCode, err)
		return err
	}

	if err := s.redis.Publish(ctx, roomPeriodEventsChannel, payload).Err(); err != nil {
		log.Printf("[engine][period.publish.error] period_id=%d room_code=%s stage=redis_publish err=%v", period.ID, period.RoomCode, err)
		return err
	}

	log.Printf("[engine][period.publish.done] period_id=%d room_code=%s channel=%s", period.ID, period.RoomCode, roomPeriodEventsChannel)
	return nil
}

func (s *RoomEngineService) generateDraw(gameType int) (repopg.DrawResult, error) {
	switch gameType {
	case 1:
		return generateWingoDraw(), nil
	case 2:
		return generateK3Draw(), nil
	case 3:
		return generateLotteryDraw(), nil
	default:
		return repopg.DrawResult{}, fmt.Errorf("game_type không hỗ trợ: %d", gameType)
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
	color := "red"
	if number%2 != 0 {
		color = "green"
	}
	if number == 0 || number == 5 {
		color = "violet"
	}

	tags := []string{
		fmt.Sprintf("number_%d", number),
		bigSmall,
		oddEven,
		color,
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
	bigSmall := "small"
	if last >= 5 {
		bigSmall = "big"
	}
	oddEven := "even"
	if last%2 != 0 {
		oddEven = "odd"
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
		bigSmall,
		oddEven,
	}

	payload, _ := json.Marshal(map[string]any{
		"game_type":    "lottery",
		"digits":       digits,
		"sum":          sum,
		"last_digit":   last,
		"result":       result,
		"big_small":    bigSmall,
		"odd_even":     oddEven,
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
