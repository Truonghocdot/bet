package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	repopg "gin/internal/repository/postgres"

	goredis "github.com/redis/go-redis/v9"
)

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
	if err := s.runTick(ctx); err != nil {
		log.Printf("[engine] tick lỗi ban đầu: %v", err)
	}

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.runTick(ctx); err != nil {
				log.Printf("[engine] tick lỗi: %v", err)
			}
		}
	}
}

func (s *RoomEngineService) runTick(ctx context.Context) error {
	now := time.Now()
	rooms, err := s.gameRepository.ListRooms(ctx)
	if err != nil {
		return err
	}

	for _, room := range rooms {
		lockKey := fmt.Sprintf("engine:room:ensure:%s", room.Code)
		acquired, err := s.acquireLock(ctx, lockKey, 3*time.Second)
		if err != nil {
			log.Printf("[engine] không lock được room %s: %v", room.Code, err)
			continue
		}
		if !acquired {
			continue
		}

		if err := s.gameRepository.EnsureRoomPeriods(ctx, room, now); err != nil {
			log.Printf("[engine] ensure period lỗi room=%s err=%v", room.Code, err)
		}
		s.releaseLock(ctx, lockKey)
	}

	if _, err := s.gameRepository.MoveScheduledToOpen(ctx, now); err != nil {
		log.Printf("[engine] chuyển SCHEDULED->OPEN lỗi: %v", err)
	}
	if _, err := s.gameRepository.MoveOpenToLocked(ctx, now); err != nil {
		log.Printf("[engine] chuyển OPEN->LOCKED lỗi: %v", err)
	}

	lockedPeriods, err := s.gameRepository.ListLockedPeriodsForDraw(ctx, now, 200)
	if err != nil {
		return err
	}
	for _, period := range lockedPeriods {
		lockKey := fmt.Sprintf("engine:period:draw:%d", period.ID)
		acquired, err := s.acquireLock(ctx, lockKey, 5*time.Second)
		if err != nil || !acquired {
			continue
		}

		draw, err := s.generateDraw(period.GameType)
		if err != nil {
			log.Printf("[engine] sinh kết quả lỗi period=%d err=%v", period.ID, err)
			s.releaseLock(ctx, lockKey)
			continue
		}

		if err := s.gameRepository.MarkPeriodDrawn(ctx, period, draw); err != nil {
			log.Printf("[engine] đánh dấu DRAWN lỗi period=%d err=%v", period.ID, err)
		}
		s.releaseLock(ctx, lockKey)
	}

	drawnPeriods, err := s.gameRepository.ListDrawnPeriodsForSettlement(ctx, 200)
	if err != nil {
		return err
	}
	for _, period := range drawnPeriods {
		lockKey := fmt.Sprintf("engine:period:settle:%d", period.ID)
		acquired, err := s.acquireLock(ctx, lockKey, 5*time.Second)
		if err != nil || !acquired {
			continue
		}
		if err := s.gameRepository.SettlePeriod(ctx, period); err != nil {
			log.Printf("[engine] settlement lỗi period=%d err=%v", period.ID, err)
		}
		s.releaseLock(ctx, lockKey)
	}

	return nil
}

func (s *RoomEngineService) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.redis.SetNX(ctx, key, "1", ttl).Result()
}

func (s *RoomEngineService) releaseLock(ctx context.Context, key string) {
	_, _ = s.redis.Del(ctx, key).Result()
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
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
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
		"generated_at": time.Now(),
	})

	return repopg.DrawResult{
		Result:      strconv.Itoa(number),
		BigSmall:    bigSmall,
		Color:       color,
		PayloadJSON: payload,
	}
}

func generateK3Draw() repopg.DrawResult {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
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
		"generated_at": time.Now(),
	})

	return repopg.DrawResult{
		Result:      result,
		BigSmall:    bigSmall,
		Color:       "-",
		PayloadJSON: payload,
	}
}

func generateLotteryDraw() repopg.DrawResult {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
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
		"generated_at": time.Now(),
	})

	return repopg.DrawResult{
		Result:      result,
		BigSmall:    bigSmall,
		Color:       "-",
		PayloadJSON: payload,
	}
}
