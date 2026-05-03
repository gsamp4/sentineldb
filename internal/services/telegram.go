package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"
	"strings"
	"time"
)

type telegramMessagePayload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func Notify(log *logger.Logger, token, chatID string, finding models.Finding) error {
	startedAt := time.Now()

	if strings.TrimSpace(token) == "" {
		err := fmt.Errorf("telegram token is required")
		log.Errorf("telegram notify aborted reason=%s finding_id=%s", "missing token", finding.ID)
		return err
	}

	if strings.TrimSpace(chatID) == "" {
		err := fmt.Errorf("telegram chat_id is required")
		log.Errorf("telegram notify aborted reason=%s finding_id=%s", "missing chat_id", finding.ID)
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	messageText := buildTelegramFindingMessage(finding)
	payload := telegramMessagePayload{
		ChatID: chatID,
		Text:   messageText,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("telegram notify payload serialization failed finding_id=%s err=%v", finding.ID, err)
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Errorf("telegram notify request failed finding_id=%s err=%v elapsed=%s", finding.ID, err, time.Since(startedAt))
		return err
	}
	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Errorf("telegram notify response read failed finding_id=%s status=%d err=%v", finding.ID, resp.StatusCode, readErr)
		return readErr
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("telegram notify rejected finding_id=%s status=%d response=%s elapsed=%s", finding.ID, resp.StatusCode, string(responseBody), time.Since(startedAt))
		return fmt.Errorf("erro na API: status %d: %s", resp.StatusCode, string(responseBody))
	}

	log.Infof("telegram ok finding=%s chat=%s", finding.ID, maskTelegramTarget(chatID))

	return nil
}

func buildTelegramFindingMessage(finding models.Finding) string {
	detail := strings.TrimSpace(formatTelegramFindingDetail(finding.Detail))
	if detail == "" {
		detail = "No additional detail provided."
	}

	return strings.Join([]string{
		"SentinelDB detected a finding update.",
		"",
		fmt.Sprintf("Title: %s", fallbackFindingValue(finding.Title, "Untitled")),
		fmt.Sprintf("Severity: %s", strings.ToUpper(fallbackFindingValue(finding.Severity, "unknown"))),
		fmt.Sprintf("Source: %s", fallbackFindingValue(finding.Source, "unknown")),
		fmt.Sprintf("Asset ID: %s", fallbackFindingValue(finding.AssetID, "unknown")),
		fmt.Sprintf("Run ID: %s", fallbackFindingValue(finding.RunID, "unknown")),
		"",
		"Details:",
		detail,
	}, "\n")
}

func formatTelegramFindingDetail(detail []byte) string {
	trimmed := strings.TrimSpace(string(detail))
	if trimmed == "" || trimmed == "null" {
		return ""
	}

	var values []interface{}
	if err := json.Unmarshal(detail, &values); err == nil && len(values) > 0 {
		lines := make([]string, 0, len(values))
		for _, value := range values {
			lines = append(lines, fmt.Sprintf("- %v", value))
		}
		return strings.Join(lines, "\n")
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, detail, "", "  "); err == nil {
		return pretty.String()
	}

	return trimmed
}

func fallbackFindingValue(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}

	return trimmed
}

func maskTelegramTarget(chatID string) string {
	trimmed := strings.TrimSpace(chatID)
	if len(trimmed) <= 4 {
		return trimmed
	}

	return strings.Repeat("*", len(trimmed)-4) + trimmed[len(trimmed)-4:]
}
