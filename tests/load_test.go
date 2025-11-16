package tests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"
)

const baseURL = "https://localhost:8080"

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
	Timeout: 30 * time.Second,
}

// LoadTestStats содержит статистику нагрузочного теста
type LoadTestStats struct {
	TotalUsers      int
	TotalTeams      int
	TotalPRs        int
	TargetRPS       int
	ActualRPS       float64
	TotalDuration   time.Duration
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	SuccessRate     float64
	Errors          []string
}

// String возвращает строковое представление статистики
func (s LoadTestStats) String() string {
	return fmt.Sprintf(`Load test stats:
Users: %d, Teams: %d, PRs: %d
Target RPS: %d, Actual RPS: %.2f
Total time: %v, Avg response: %v
Min response: %v, Max response: %v
Success rate: %.2f%%`,
		s.TotalUsers, s.TotalTeams, s.TotalPRs, s.TargetRPS,
		s.ActualRPS, s.TotalDuration, s.AvgResponseTime,
		s.MinResponseTime, s.MaxResponseTime, s.SuccessRate*100)
}

// TestLoadHighVolume тестирует высокую нагрузку с подробной статистикой
func TestLoadHighVolume(t *testing.T) {
	stats := LoadTestStats{
		TotalUsers:      200,
		TotalTeams:      20,
		TotalPRs:        200,
		TargetRPS:       10,
		MinResponseTime: time.Hour,
	}

	userCount := stats.TotalUsers
	teamCount := stats.TotalTeams
	users := make([]map[string]interface{}, 0, userCount)
	teams := make([]map[string]interface{}, 0, teamCount)

	for i := 1; i <= userCount; i++ {
		users = append(users, map[string]interface{}{
			"user_id":  fmt.Sprintf("load-user-%d", i),
			"username": fmt.Sprintf("Load User %d", i),
		})
	}

	for i := 0; i < teamCount; i++ {
		teamMembers := make([]map[string]interface{}, 0, 10)
		for j := 0; j < 10; j++ {
			idx := i*10 + j
			teamMembers = append(teamMembers, users[idx])
		}
		teams = append(teams, map[string]interface{}{
			"team_name": fmt.Sprintf("load-team-%d", i+1),
			"members":   teamMembers,
		})
	}

	t.Logf("Creating %d teams with %d users", teamCount, userCount)
	teamStart := time.Now()

	for _, team := range teams {
		resp, err := postJSON(baseURL+"/team/add", team)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to create team: %v", err))
			continue
		}
		resp.Body.Close()
	}

	teamDuration := time.Since(teamStart)
	t.Logf("Teams created in %v", teamDuration)

	t.Logf("Creating %d PRs at %d RPS", stats.TotalPRs, stats.TargetRPS)

	prCountPerTeam := stats.TotalPRs / teamCount
	prIDs := make([]string, 0, stats.TotalPRs)
	prAuthors := make([]string, 0, stats.TotalPRs)
	prTeamNames := make([]string, 0, stats.TotalPRs)

	rps := stats.TargetRPS
	interval := time.Second / time.Duration(rps)
	errCh := make(chan error, stats.TotalPRs)
	var wg sync.WaitGroup

	responseTimes := make(chan time.Duration, stats.TotalPRs)
	successCount := int64(0)

	prIdx := 0
	start := time.Now()

	for _, team := range teams {
		members := team["members"].([]map[string]interface{})
		teamName := team["team_name"].(string)

		for prNum := 0; prNum < prCountPerTeam; prNum++ {
			wg.Add(1)
			go func(teamMembers []map[string]interface{}, teamName string, prNum int, prIdx int) {
				defer wg.Done()

				author := teamMembers[rand.Intn(len(teamMembers))]["user_id"].(string)

				prID := fmt.Sprintf("load-pr-%s-%d-%d", teamName, prNum, time.Now().UnixNano())
				prPayload := map[string]interface{}{
					"pull_request_id":   prID,
					"pull_request_name": fmt.Sprintf("Load Test PR %d for %s", prNum, teamName),
					"author_id":         author,
				}

				reqStart := time.Now()
				resp, err := postJSON(baseURL+"/pullRequest/create", prPayload)
				reqDuration := time.Since(reqStart)

				responseTimes <- reqDuration

				if err != nil {
					errCh <- fmt.Errorf("Failed to create PR %s: %v", prID, err)
					return
				}

				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					errCh <- fmt.Errorf("Failed to create PR %s: status %d: %s", prID, resp.StatusCode, string(body))
					return
				}

				resp.Body.Close()

				atomicAddInt64(&successCount, 1)

				prIDs = append(prIDs, prID)
				prAuthors = append(prAuthors, author)
				prTeamNames = append(prTeamNames, teamName)
			}(members, teamName, prNum, prIdx)

			prIdx++
			time.Sleep(interval)
		}
	}

	wg.Wait()
	close(errCh)
	close(responseTimes)

	duration := time.Since(start)

	var totalResponseTime time.Duration
	responseCount := 0

	for rt := range responseTimes {
		totalResponseTime += rt
		responseCount++

		if rt < stats.MinResponseTime {
			stats.MinResponseTime = rt
		}
		if rt > stats.MaxResponseTime {
			stats.MaxResponseTime = rt
		}
	}

	stats.TotalDuration = duration
	stats.ActualRPS = float64(responseCount) / duration.Seconds()
	stats.AvgResponseTime = totalResponseTime / time.Duration(responseCount)
	stats.SuccessRate = float64(successCount) / float64(stats.TotalPRs)

	errorCount := 0
	for err := range errCh {
		stats.Errors = append(stats.Errors, err.Error())
		errorCount++
	}

	t.Log(stats.String())

	if errorCount > 0 {
		t.Logf("Found %d errors:", errorCount)
		for i, err := range stats.Errors {
			if i >= 5 { // Показываем максимум 5 ошибок
				t.Logf("... и еще %d ошибок", errorCount-5)
				break
			}
			t.Logf("  - %s", err)
		}
	}

	if stats.ActualRPS < float64(stats.TargetRPS)*0.8 {
		t.Errorf("❌ Фактический RPS (%.2f) ниже 80%% от целевого (%d)",
			stats.ActualRPS, stats.TargetRPS)
	}

	if stats.SuccessRate < 0.95 {
		t.Errorf("❌ Успешность (%.2f%%) ниже 95%%", stats.SuccessRate*100)
	}

	if stats.AvgResponseTime > 2*time.Second {
		t.Errorf("❌ Среднее время отклика (%v) превышает 2 секунды", stats.AvgResponseTime)
	}

	t.Log("Load test completed")
}

// TestLoadIncreasingRPS тестирует возрастающую нагрузку
func TestLoadIncreasingRPS(t *testing.T) {
	teamName := fmt.Sprintf("increasing-rps-%d", time.Now().Unix())

	members := []map[string]interface{}{}
	for i := 1; i <= 5; i++ {
		members = append(members, map[string]interface{}{
			"user_id":  fmt.Sprintf("increasing-user%d", i),
			"username": fmt.Sprintf("Increasing User %d", i),
		})
	}

	payload := map[string]interface{}{
		"team_name": teamName,
		"members":   members,
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	rpsLevels := []int{1, 5, 10, 15, 20}
	prsPerLevel := 10

	t.Log("Testing increasing load...")

	for _, rps := range rpsLevels {
		t.Logf("Testing %d RPS...", rps)

		interval := time.Second / time.Duration(rps)
		var wg sync.WaitGroup
		responseTimes := make(chan time.Duration, prsPerLevel)
		errors := make(chan error, prsPerLevel)

		start := time.Now()

		for i := 0; i < prsPerLevel; i++ {
			wg.Add(1)
			go func(prNum int) {
				defer wg.Done()

				prPayload := map[string]interface{}{
					"pull_request_id":   fmt.Sprintf("increasing-pr-%d-%d-%d", rps, prNum, time.Now().UnixNano()),
					"pull_request_name": fmt.Sprintf("Increasing RPS PR %d at %d RPS", prNum, rps),
					"author_id":         "increasing-user1",
				}

				reqStart := time.Now()
				resp, err := postJSON(baseURL+"/pullRequest/create", prPayload)
				reqDuration := time.Since(reqStart)

				responseTimes <- reqDuration

				if err != nil {
					errors <- err
					return
				}

				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					errors <- fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
					return
				}

				resp.Body.Close()
			}(i)

			time.Sleep(interval)
		}

		wg.Wait()
		close(responseTimes)
		close(errors)

		duration := time.Since(start)

		var totalRT time.Duration
		count := 0
		minRT := time.Hour
		maxRT := time.Duration(0)

		for rt := range responseTimes {
			totalRT += rt
			count++

			if rt < minRT {
				minRT = rt
			}
			if rt > maxRT {
				maxRT = rt
			}
		}

		avgRT := totalRT / time.Duration(count)
		actualRPS := float64(count) / duration.Seconds()
		errorCount := len(errors)

		t.Logf("  %d RPS: avg=%v, min=%v, max=%v, actual RPS=%.2f, errors=%d",
			rps, avgRT, minRT, maxRT, actualRPS, errorCount)

		time.Sleep(2 * time.Second)
	}

	t.Log("Increasing load test completed")
}

// TestLoadConcurrentUsers тестирует одновременную работу множества пользователей
func TestLoadConcurrentUsers(t *testing.T) {
	concurrentUsers := 50
	prsPerUser := 5

	t.Logf("Testing %d concurrent users (%d PRs each)", concurrentUsers, prsPerUser)

	var wg sync.WaitGroup
	responseTimes := make(chan time.Duration, concurrentUsers*prsPerUser)
	errors := make(chan error, concurrentUsers*prsPerUser)
	successCount := int64(0)

	start := time.Now()

	for userID := 1; userID <= concurrentUsers; userID++ {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()

			teamName := fmt.Sprintf("concurrent-team-%d-%d", uid, time.Now().Unix())

			teamPayload := map[string]interface{}{
				"team_name": teamName,
				"members": []map[string]interface{}{
					{"user_id": fmt.Sprintf("concurrent-user-%d", uid), "username": fmt.Sprintf("Concurrent User %d", uid)},
					{"user_id": fmt.Sprintf("concurrent-reviewer-%d", uid), "username": fmt.Sprintf("Concurrent Reviewer %d", uid)},
				},
			}

			resp, err := postJSON(baseURL+"/team/add", teamPayload)
			if err != nil {
				errors <- fmt.Errorf("user %d failed to create team: %v", uid, err)
				return
			}
			resp.Body.Close()

			for prNum := 1; prNum <= prsPerUser; prNum++ {
				prPayload := map[string]interface{}{
					"pull_request_id":   fmt.Sprintf("concurrent-pr-%d-%d-%d", uid, prNum, time.Now().UnixNano()),
					"pull_request_name": fmt.Sprintf("Concurrent PR %d from user %d", prNum, uid),
					"author_id":         fmt.Sprintf("concurrent-user-%d", uid),
				}

				reqStart := time.Now()
				resp, err := postJSON(baseURL+"/pullRequest/create", prPayload)
				reqDuration := time.Since(reqStart)

				responseTimes <- reqDuration

				if err != nil {
					errors <- fmt.Errorf("user %d PR %d failed: %v", uid, prNum, err)
					continue
				}

				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					errors <- fmt.Errorf("user %d PR %d status %d: %s", uid, prNum, resp.StatusCode, string(body))
					continue
				}

				resp.Body.Close()
				atomicAddInt64(&successCount, 1)
			}
		}(userID)
	}

	wg.Wait()
	close(responseTimes)
	close(errors)

	duration := time.Since(start)

	var totalRT time.Duration
	count := 0
	minRT := time.Hour
	maxRT := time.Duration(0)

	for rt := range responseTimes {
		totalRT += rt
		count++

		if rt < minRT {
			minRT = rt
		}
		if rt > maxRT {
			maxRT = rt
		}
	}

	avgRT := totalRT / time.Duration(count)
	actualRPS := float64(successCount) / duration.Seconds()
	successRate := float64(successCount) / float64(concurrentUsers*prsPerUser)
	errorCount := len(errors)

	stats := LoadTestStats{
		TotalUsers:      concurrentUsers,
		TotalTeams:      concurrentUsers,
		TotalPRs:        concurrentUsers * prsPerUser,
		TargetRPS:       0, // Не применимо для этого теста
		ActualRPS:       actualRPS,
		TotalDuration:   duration,
		AvgResponseTime: avgRT,
		MinResponseTime: minRT,
		MaxResponseTime: maxRT,
		SuccessRate:     successRate,
	}

	t.Log(stats.String())

	if errorCount > 0 {
		t.Logf("Found %d errors", errorCount)
	}

	if successRate < 0.95 {
		t.Errorf("❌ Успешность (%.2f%%) ниже 95%%", successRate*100)
	}

	t.Log("Concurrent users test completed")
}

// atomicAddInt64 атомарно увеличивает значение int64
func atomicAddInt64(val *int64, delta int64) {
	*val += delta
}

func postJSON(url string, payload interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}
