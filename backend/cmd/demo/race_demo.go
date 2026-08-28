/*
================================================================================
🏆 PLAY HACK - LIVE DEMO PRESENTATION SCRIPT
================================================================================

PRE-DEMO 3-STEP CHECKLIST:
--------------------------------------------------------------------------------
1. Start the API server:
   $ cd backend && go run cmd/api/main.go

2. Seed the database:
   $ cd backend && go run cmd/seed/main.go

3. Run this live demo script:
   $ cd backend && go run cmd/demo/race_demo.go --reset

Options:
   --url     API Base URL (default: "http://localhost:8080")
   --slot    Target Slot UUID (default: auto-discovered from active future unbooked slots)
   --users   Number of simultaneous users (default: 2, range: 2-10)
   --reset   Reset/cancel any existing booking on the target slot before race
================================================================================
*/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// ANSI Color constants
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

type Facility struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SportType string `json:"sport_type"`
}

type Court struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Slot struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}

type BookingDetail struct {
	ID     string `json:"id"`
	SlotID string `json:"slot_id"`
}

type DemoUser struct {
	Email string
	Token string
}

type RaceResult struct {
	UserEmail    string
	StatusCode   int
	Latency      time.Duration
	ResponseBody string
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "API base URL")
	targetSlot := flag.String("slot", "", "Target slot UUID (auto-discovered if empty)")
	numUsers := flag.Int("users", 2, "Number of concurrent users (2 to 10)")
	reset := flag.Bool("reset", false, "Cancel existing booking on target slot before race")
	flag.Parse()

	if *numUsers < 2 {
		*numUsers = 2
	}

	fmt.Println()
	fmt.Printf("%s%s========================================================================%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s%s 🏆 PLAY HACK — ATOMIC CONCURRENCY ENGINE LIVE DEMO %s\n", ColorBold, ColorYellow, ColorReset)
	fmt.Printf("%s%s========================================================================%s\n", ColorBold, ColorCyan, ColorReset)

	// Step 1: Register/Authenticate Demo Users
	fmt.Printf("%s[1/3] Authenticating %d Demo Users...%s\n", ColorYellow, *numUsers, ColorReset)
	users := make([]DemoUser, *numUsers)
	ts := time.Now().Unix() % 10000

	for i := 0; i < *numUsers; i++ {
		email := fmt.Sprintf("judge%d_%d@iitg.ac.in", i+1, ts)
		token, err := authenticateUser(*baseURL, email)
		if err != nil {
			fmt.Printf("%s❌ Error authenticating %s: %v%s\n", ColorRed, email, err, ColorReset)
			fmt.Println("Please make sure the backend server is running on", *baseURL)
			os.Exit(1)
		}
		users[i] = DemoUser{Email: email, Token: token}
		fmt.Printf("   ✓ Authenticated %s\n", email)
	}

	// Step 2: Auto-Discover Target Slot if not specified
	slotID := *targetSlot
	slotInfo := "Target Slot"
	if slotID == "" {
		fmt.Printf("%s[2/3] Auto-discovering active unbooked future court slot...%s\n", ColorYellow, ColorReset)
		discoveredID, info, err := discoverFutureSlot(*baseURL, *reset)
		if err != nil {
			fmt.Printf("%s❌ Slot discovery failed: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}
		slotID = discoveredID
		slotInfo = info
		fmt.Printf("   ✓ Discovered Slot: %s\n", slotInfo)
	} else {
		fmt.Printf("%s[2/3] Target Slot specified: %s%s\n", ColorYellow, slotID, ColorReset)
	}

	// Step 3: Handle --reset flag if requested
	if *reset {
		fmt.Printf("%s[3/3] Resetting target slot bookings...%s\n", ColorYellow, ColorReset)
		for _, u := range users {
			_ = cancelUserBookingForSlot(*baseURL, u.Token, slotID)
		}
		fmt.Println("   ✓ Slot reset check complete.")
	} else {
		fmt.Printf("%s[3/3] Preparing synchronized race...%s\n", ColorYellow, ColorReset)
	}

	// Banner Display
	fmt.Println()
	fmt.Printf("%s========================================================================%s\n", ColorCyan, ColorReset)
	fmt.Printf(" Server URL      : %s\n", *baseURL)
	fmt.Printf(" Target Slot     : %s\n", slotInfo)
	fmt.Printf(" Target Slot ID  : %s\n", slotID)
	fmt.Printf(" Race Mode       : %s%d Simultaneous Workers (Synchronized Barrier)%s\n", ColorBold, *numUsers, ColorReset)
	fmt.Printf("%s========================================================================%s\n", ColorCyan, ColorReset)
	fmt.Printf("%s⚡ Firing %d simultaneous booking requests at the exact same instant...%s\n\n", ColorYellow, *numUsers, ColorReset)

	// Step 4: Synchronized Goroutine Race Barrier
	var wg sync.WaitGroup
	startBarrier := make(chan struct{})
	resultsChan := make(chan RaceResult, *numUsers)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"slot_id":      slotID,
		"player_count": 2,
	})

	for i := 0; i < *numUsers; i++ {
		wg.Add(1)
		user := users[i]

		go func(u DemoUser) {
			defer wg.Done()

			req, _ := http.NewRequest("POST", *baseURL+"/bookings", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+u.Token)

			client := &http.Client{Timeout: 5 * time.Second}

			// Block until barrier release
			<-startBarrier
			startTime := time.Now()

			resp, err := client.Do(req)
			latency := time.Since(startTime)

			if err != nil {
				resultsChan <- RaceResult{UserEmail: u.Email, StatusCode: 500, Latency: latency, ResponseBody: err.Error()}
				return
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)
			resultsChan <- RaceResult{
				UserEmail:    u.Email,
				StatusCode:   resp.StatusCode,
				Latency:      latency,
				ResponseBody: string(respBytes),
			}
		}(user)
	}

	// RELEASE BARRIER SIMULTANEOUSLY
	close(startBarrier)
	wg.Wait()
	close(resultsChan)

	// Step 5: Process & Output Results Live
	var successCount, conflictCount, failureCount int

	for res := range resultsChan {
		latencyMs := float64(res.Latency.Microseconds()) / 1000.0

		if res.StatusCode == http.StatusCreated {
			successCount++
			fmt.Printf(" [%s%.2fms%s] %s %-25s -> %s201 Created%s  — Booking confirmed ✅\n",
				ColorBold, latencyMs, ColorReset, ColorBold, res.UserEmail, ColorGreen, ColorReset)
		} else if res.StatusCode == http.StatusConflict {
			conflictCount++
			fmt.Printf(" [%s%.2fms%s] %s %-25s -> %s409 Conflict%s — Slot already booked ❌\n",
				ColorBold, latencyMs, ColorReset, ColorBold, res.UserEmail, ColorRed, ColorReset)
		} else {
			failureCount++
			fmt.Printf(" [%s%.2fms%s] %s %-25s -> %s%d Error%s    — %s\n",
				ColorBold, latencyMs, ColorReset, ColorBold, res.UserEmail, ColorRed, res.StatusCode, ColorReset, res.ResponseBody)
		}
	}

	// Step 6: Summary Verdict
	fmt.Println()
	fmt.Printf("%s========================================================================%s\n", ColorCyan, ColorReset)

	if successCount == 1 && conflictCount == (*numUsers-1) && failureCount == 0 {
		fmt.Printf("%s%s RESULT: 🎉 VERDICT PASSED — Exactly 1 booking succeeded (201), %d correctly rejected (409). %s\n",
			ColorBold, ColorGreen, conflictCount, ColorReset)
		fmt.Printf("%s         Database single-winner consistency verified under live race conditions!%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s%s RESULT: ❌ VERDICT FAILED — Expected 1x 201 Created & %dx 409 Conflict. Got %d Success, %d Conflicts, %d Errors.%s\n",
			ColorBold, ColorRed, *numUsers-1, successCount, conflictCount, failureCount, ColorReset)
	}
	fmt.Printf("%s========================================================================%s\n\n", ColorCyan, ColorReset)
}

func authenticateUser(baseURL, email string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	reqBody, _ := json.Marshal(map[string]string{"email": email})
	resp, err := client.Post(baseURL+"/auth/request-otp", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var otpResp map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&otpResp)

	code := otpResp["code"]
	if code == "" {
		return "", fmt.Errorf("failed to request OTP for %s: %s", email, resp.Status)
	}

	verifyBody, _ := json.Marshal(map[string]string{"email": email, "code": code})
	vResp, err := client.Post(baseURL+"/auth/verify-otp", "application/json", bytes.NewBuffer(verifyBody))
	if err != nil {
		return "", err
	}
	defer vResp.Body.Close()

	var tokenResp map[string]string
	if err := json.NewDecoder(vResp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	token := tokenResp["token"]
	if token == "" {
		return "", fmt.Errorf("no token returned for %s", email)
	}
	return token, nil
}

func discoverFutureSlot(baseURL string, requireAvailable bool) (string, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL + "/facilities")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var facilities []Facility
	if err := json.NewDecoder(resp.Body).Decode(&facilities); err != nil || len(facilities) == 0 {
		return "", "", fmt.Errorf("no active facilities found")
	}

	datesToTry := []string{
		time.Now().Format("2006-01-02"),
		time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
	}

	now := time.Now()

	for _, facility := range facilities {
		cResp, err := client.Get(fmt.Sprintf("%s/facilities/%s/courts", baseURL, facility.ID))
		if err != nil {
			continue
		}
		var courts []Court
		_ = json.NewDecoder(cResp.Body).Decode(&courts)
		cResp.Body.Close()

		for _, court := range courts {
			for _, dateStr := range datesToTry {
				sResp, err := client.Get(fmt.Sprintf("%s/courts/%s/slots?date=%s", baseURL, court.ID, dateStr))
				if err != nil {
					continue
				}
				var slots []Slot
				_ = json.NewDecoder(sResp.Body).Decode(&slots)
				sResp.Body.Close()

				for _, s := range slots {
					if s.StartTime.After(now) {
						if !requireAvailable || s.Available {
							info := fmt.Sprintf("%s - %s (%s, %s - %s)",
								facility.Name, court.Label,
								s.StartTime.Format("2006-01-02"),
								s.StartTime.Format("15:04"),
								s.EndTime.Format("15:04"))
							return s.ID, info, nil
						}
					}
				}
			}
		}
	}

	return "", "", fmt.Errorf("no available future slots found")
}

func cancelUserBookingForSlot(baseURL, token, slotID string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	req, _ := http.NewRequest("GET", baseURL+"/bookings/mine", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var bookings []BookingDetail
	if err := json.NewDecoder(resp.Body).Decode(&bookings); err != nil {
		return err
	}

	for _, b := range bookings {
		if b.SlotID == slotID {
			delReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/bookings/%s", baseURL, b.ID), nil)
			delReq.Header.Set("Authorization", "Bearer "+token)
			delResp, err := client.Do(delReq)
			if err == nil {
				delResp.Body.Close()
			}
		}
	}
	return nil
}
