package seeding

import (
	"context"
	"crypto/md5"
	"unicode"

	"fmt"
	"log"
	"regexp"
	"strings"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

type QualityFilter struct {
	stateService service.QualityStateService
	config       model.QualityCheckConfig
}

func NewQualityFilter(stateService service.QualityStateService, config model.QualityCheckConfig) *QualityFilter {
	return &QualityFilter{
		stateService: stateService,
		config:       config,
	}
}

func (q *QualityFilter) Check(ctx context.Context, event model.MatchEvent, message *model.DraftMessage, bundle *model.ContextBundle) (model.QualityResult, error) {
	text := strings.TrimSpace(message.Text)
	textLower := strings.ToLower(text)
	// ============1.Check length=========
	textLen := len([]rune(text))
	if textLen < q.config.MinLength {
		return model.QualityResult{IsPass: false, Reason: []string{"too short"}, Action: []string{"retry_p1"}}, nil
	}
	if textLen > q.config.MaxLength {
		return model.QualityResult{IsPass: false, Reason: []string{"too long"}, Action: []string{"retry_p1"}}, nil
	}
	// ============2.Check banned words=========
	for _, banned := range q.config.BannedWords {
		if strings.Contains(textLower, strings.ToLower(banned)) {
			log.Printf("⏭️  [QUALITY] banned word: %q", banned)
			return model.QualityResult{IsPass: false, Reason: []string{"banned content"}, Action: []string{"retry_p1"}}, nil
		}
	}
	if !q.checkEventConsistency(event, textLower) {
		log.Printf("⏭️  [QUALITY] inconsistent with event=%s", event.Type)
		return model.QualityResult{IsPass: false, Reason: []string{"inconsistent with event"}, Action: []string{"retry_p1"}}, nil
	}
	// ========== 4. Check Anti-Repeat (Duplicate) ==========
	isPrematch := event.Type == "MATCH_UPCOMING" || event.Minute < 0
	if !isPrematch {
		normalizedText := normalizeMessageForHash(textLower)
		msgHash := hashMessage(normalizedText)
		isDuplicated, err := q.stateService.IsMessageDuplicated(bundle.Match.MatchID, msgHash)
		if err != nil {
			log.Printf("   ⚠️  [QUALITY] Failed to check duplicate: %v", err)
		} else if isDuplicated {
			return model.QualityResult{IsPass: false, Reason: []string{"duplicate"}, Action: []string{"retry_p1"}}, nil
		}
		if err := q.stateService.SaveMessageHash(bundle.Match.MatchID, msgHash, q.config.DedupTTL); err != nil {
			log.Printf("   ⚠️  [QUALITY] Failed to save hash: %v", err)
		}
	}
	// ========== 5. PASS ==========
	return model.QualityResult{IsPass: true, Reason: []string{}, Action: []string{"proceed"}}, nil
}

func (q *QualityFilter) checkEventConsistency(event model.MatchEvent, textLower string) bool {
	switch event.Type {
	case "GOAL":
		// FIX: mở rộng keyword tiếng Việt, tránh false positive từ "vào" quá ngắn
		goalKeywords := []string{
			"goal", "goall", "ghi bàn", "bàn thắng", "vào lưới",
			"nổ súng", "⚽", "vô lưới", "ghi được",
		}
		return containsAny(textLower, goalKeywords)

	case "RED_CARD":
		redCardKeywords := []string{
			"thẻ đỏ", "red card", "đuổi", "off", "bye bye",
		}
		return containsAny(textLower, redCardKeywords)

	case "PENALTY":
		penaltyKeywords := []string{
			"penalty", "phạt đền", "chấm 11m", "pk",
		}
		return containsAny(textLower, penaltyKeywords)
	}
	// Các event khác không yêu cầu keyword cụ thể
	return true
}
func containsAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func normalizeMessageForHash(text string) string {
	s := regexp.MustCompile(`\d+`).ReplaceAllString(text, "")

	// FIX: bỏ emoji đúng cách bằng cách filter theo Unicode range
	// thay vì regex \p{So}\p{Sk} chỉ bắt được một phần
	var b strings.Builder
	for _, r := range s {
		if !isEmoji(r) {
			b.WriteRune(r)
		}
	}
	s = b.String()

	// Normalize khoảng trắng
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return s

}
func isEmoji(r rune) bool {
	return unicode.Is(unicode.So, r) || // Symbol, other (⚽ ⚠ ▶)
		unicode.Is(unicode.Sk, r) || // Symbol, modifier
		(r >= 0x1F300 && r <= 0x1FAFF) || // Miscellaneous Symbols and Pictographs, Emoticons
		(r >= 0x2600 && r <= 0x27BF) || // Misc symbols (☀ ♥)
		(r >= 0xFE00 && r <= 0xFE0F) // Variation selectors
}

func hashMessage(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}
