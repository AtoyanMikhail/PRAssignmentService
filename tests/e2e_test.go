package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestE2EHealthCheck(t *testing.T) {
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", result["status"])
	}
}

func TestE2EReassignment(t *testing.T) {
	teamName := fmt.Sprintf("reassign-team-%d", time.Now().Unix())

	payload := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "reassign-user1", "username": "Reassign User 1"},
			{"user_id": "reassign-user2", "username": "Reassign User 2"},
			{"user_id": "reassign-user3", "username": "Reassign User 3"},
			{"user_id": "reassign-user4", "username": "Reassign User 4"},
		},
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	_ = resp.Body.Close()

	prID := fmt.Sprintf("reassign-pr-%d", time.Now().Unix())
	prPayload := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "Reassignment Test PR",
		"author_id":         "reassign-user1",
	}

	resp, err = postJSON(baseURL+"/pullRequest/create", prPayload)
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("Failed to create PR: status %d: %s", resp.StatusCode, string(body))
	}

	var prResult map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&prResult)
	_ = resp.Body.Close()

	reviewers := prResult["assigned_reviewers"].([]interface{})
	if len(reviewers) == 0 {
		t.Fatal("No reviewers assigned")
	}

	oldReviewer := reviewers[0].(string)

	t.Run("ReassignReviewer", func(t *testing.T) {
		reassignPayload := map[string]interface{}{
			"pull_request_id": prID,
			"old_user_id":     oldReviewer,
		}

		resp, err := postJSON(baseURL+"/pullRequest/reassign", reassignPayload)
		if err != nil {
			t.Fatalf("Failed to reassign reviewer: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["status"] != "reassigned" {
			t.Errorf("Expected status 'reassigned', got %v", result["status"])
		}
	})
}

func TestE2ETeamOperations(t *testing.T) {
	teamName := fmt.Sprintf("team-ops-%d", time.Now().Unix())

	payload := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "team-user1", "username": "Team User 1"},
			{"user_id": "team-user2", "username": "Team User 2"},
		},
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	_ = resp.Body.Close()

	resp, err = client.Get(baseURL + "/team/get?team_name=" + teamName)
	if err != nil {
		t.Fatalf("Failed to get team: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var teamResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&teamResult); err != nil {
		t.Fatalf("Failed to decode team response: %v", err)
	}

	if teamResult["team_name"] != teamName {
		t.Errorf("Expected team name %s, got %v", teamName, teamResult["team_name"])
	}

	members := teamResult["members"].([]interface{})
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestE2EUserOperations(t *testing.T) {
	teamName := fmt.Sprintf("user-ops-%d", time.Now().Unix())
	payload := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "user-ops-1", "username": "User Ops 1"},
		},
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	setActivePayload := map[string]interface{}{
		"user_id":   "user-ops-1",
		"is_active": false,
	}

	resp, err = postJSON(baseURL+"/users/setIsActive", setActivePayload)
	if err != nil {
		t.Fatalf("Failed to set user inactive: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	var user map[string]interface{}
	if u, ok := result["user"].(map[string]interface{}); ok {
		user = u
	} else {
		user = result
	}

	if user["is_active"] != false {
		t.Errorf("Expected user to be inactive, got %v", user["is_active"])
	}
}

func TestE2EPullRequestLifecycle(t *testing.T) {
	teamName := fmt.Sprintf("pr-lifecycle-%d", time.Now().Unix())

	payload := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "pr-user1", "username": "PR User 1"},
			{"user_id": "pr-user2", "username": "PR User 2"},
			{"user_id": "pr-user3", "username": "PR User 3"},
		},
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	prID := fmt.Sprintf("pr-lifecycle-%d", time.Now().Unix())
	prPayload := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "Lifecycle Test PR",
		"author_id":         "pr-user1",
	}

	resp, err = postJSON(baseURL+"/pullRequest/create", prPayload)
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}

	var prResult map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&prResult)
	_ = resp.Body.Close()

	if prResult["status"] != "OPEN" {
		t.Errorf("Expected PR status OPEN, got %v", prResult["status"])
	}

	reviewers := prResult["assigned_reviewers"].([]interface{})
	if len(reviewers) == 0 {
		t.Fatal("No reviewers assigned")
	}

	mergePayload := map[string]interface{}{
		"pull_request_id": prID,
	}

	resp, err = postJSON(baseURL+"/pullRequest/merge", mergePayload)
	if err != nil {
		t.Fatalf("Failed to merge PR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var mergeResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mergeResult); err != nil {
		t.Fatalf("Failed to decode merge response: %v", err)
	}

	var pr map[string]interface{}
	if p, ok := mergeResult["pr"].(map[string]interface{}); ok {
		pr = p
	} else {
		pr = mergeResult
	}

	if pr["status"] != "MERGED" {
		t.Errorf("Expected PR status MERGED, got %v", pr["status"])
	}
}

func TestE2EUserReviews(t *testing.T) {
	teamName := fmt.Sprintf("reviews-%d", time.Now().Unix())

	payload := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "review-user1", "username": "Review User 1"},
			{"user_id": "review-user2", "username": "Review User 2"},
		},
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	prID := fmt.Sprintf("review-pr-%d", time.Now().Unix())
	prPayload := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "Review Test PR",
		"author_id":         "review-user1",
	}

	resp, err = postJSON(baseURL+"/pullRequest/create", prPayload)
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	resp.Body.Close()

	resp, err = client.Get(baseURL + "/users/getReview?user_id=review-user2")
	if err != nil {
		t.Fatalf("Failed to get user reviews: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	var prs []interface{}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if resultMap["user_id"] != "review-user2" {
			t.Errorf("Expected user_id review-user2, got %v", resultMap["user_id"])
		}
		if pullRequests, exists := resultMap["pull_requests"]; exists {
			prs = pullRequests.([]interface{})
		}
	} else if resultArray, ok := result.([]interface{}); ok {
		prs = resultArray
	}

	if len(prs) == 0 {
		t.Error("Expected at least one PR for reviewer")
	}
}

func TestE2EStatistics(t *testing.T) {
	teamName := fmt.Sprintf("stats-%d", time.Now().Unix())

	payload := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "stats-user1", "username": "Stats User 1"},
			{"user_id": "stats-user2", "username": "Stats User 2"},
		},
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	for i := 1; i <= 3; i++ {
		prPayload := map[string]interface{}{
			"pull_request_id":   fmt.Sprintf("stats-pr-%d-%d", time.Now().Unix(), i),
			"pull_request_name": fmt.Sprintf("Stats PR %d", i),
			"author_id":         "stats-user1",
		}

		resp, err := postJSON(baseURL+"/pullRequest/create", prPayload)
		if err != nil {
			t.Fatalf("Failed to create PR %d: %v", i, err)
		}
		resp.Body.Close()
	}

	resp, err = client.Get(baseURL + "/statistics/assignments")
	if err != nil {
		t.Fatalf("Failed to get assignments statistics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var assignmentsResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&assignmentsResult); err != nil {
		t.Fatalf("Failed to decode assignments response: %v", err)
	}

	if len(assignmentsResult) == 0 {
		t.Error("Expected assignments statistics, got empty response")
	}

	resp, err = client.Get(baseURL + "/statistics/workload")
	if err != nil {
		t.Fatalf("Failed to get workload statistics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var workloadResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&workloadResult); err != nil {
		t.Fatalf("Failed to decode workload response: %v", err)
	}

	workload := workloadResult["workload"].([]interface{})
	if len(workload) == 0 {
		t.Error("Expected workload statistics, got empty array")
	}
}

func TestE2ETeamDeactivation(t *testing.T) {
	teamName := fmt.Sprintf("deactivate-%d", time.Now().Unix())

	payload := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "deactivate-user1", "username": "Deactivate User 1"},
			{"user_id": "deactivate-user2", "username": "Deactivate User 2"},
		},
	}

	resp, err := postJSON(baseURL+"/team/add", payload)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	prID := fmt.Sprintf("deactivate-pr-%d", time.Now().Unix())
	prPayload := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "Deactivate Test PR",
		"author_id":         "deactivate-user1",
	}

	resp, err = postJSON(baseURL+"/pullRequest/create", prPayload)
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	resp.Body.Close()

	deactivatePayload := map[string]interface{}{
		"team_name": teamName,
	}

	resp, err = postJSON(baseURL+"/team/deactivate", deactivatePayload)
	if err != nil {
		t.Fatalf("Failed to deactivate team: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, hasDeactivated := result["deactivated_users"]; !hasDeactivated {
		t.Errorf("Expected 'deactivated_users' field in response, got: %v", result)
	}

	if _, hasReassigned := result["reassigned_prs"]; !hasReassigned {
		t.Errorf("Expected 'reassigned_prs' field in response, got: %v", result)
	}
}

func TestE2ELoadBalancing(t *testing.T) {
	teamName := fmt.Sprintf("balance-team-%d", time.Now().Unix())

	members := []map[string]interface{}{}
	for i := 1; i <= 5; i++ {
		members = append(members, map[string]interface{}{
			"user_id":  fmt.Sprintf("balance-user%d", i),
			"username": fmt.Sprintf("Balance User %d", i),
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

	reviewerCount := make(map[string]int)
	for i := 1; i <= 10; i++ {
		prPayload := map[string]interface{}{
			"pull_request_id":   fmt.Sprintf("balance-pr-%d-%d", time.Now().Unix(), i),
			"pull_request_name": fmt.Sprintf("Load Balance PR %d", i),
			"author_id":         "balance-user1",
		}

		resp, err := postJSON(baseURL+"/pullRequest/create", prPayload)
		if err != nil {
			t.Fatalf("Failed to create PR %d: %v", i, err)
		}

		var result map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&result)
		_ = resp.Body.Close()

		reviewers := result["assigned_reviewers"].([]interface{})
		for _, r := range reviewers {
			reviewer := r.(string)
			reviewerCount[reviewer]++
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Logf("Reviewer distribution: %v", reviewerCount)

	if count, exists := reviewerCount["balance-user1"]; exists && count > 0 {
		t.Errorf("Author should not be assigned as reviewer, but has %d assignments", count)
	}

	expectedReviewers := []string{"balance-user2", "balance-user3", "balance-user4", "balance-user5"}
	for _, reviewer := range expectedReviewers {
		if count, exists := reviewerCount[reviewer]; !exists || count == 0 {
			t.Errorf("Reviewer %s should have assignments, but has %d", reviewer, count)
		}
	}

	minCount, maxCount := 999, 0
	for _, count := range reviewerCount {
		if count < minCount {
			minCount = count
		}
		if count > maxCount {
			maxCount = count
		}
	}

	if maxCount-minCount > 5 {
		t.Errorf("Load imbalance too high: min=%d, max=%d, diff=%d", minCount, maxCount, maxCount-minCount)
	}
}
