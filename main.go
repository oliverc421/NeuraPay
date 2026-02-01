// Hackathon Starter: Complete AI Financial Agent
// Build intelligent financial tools with nim-go-sdk + Liminal banking APIs
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
	"math"
	"sort"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/executor"
	"github.com/becomeliminal/nim-go-sdk/server"
	"github.com/becomeliminal/nim-go-sdk/tools"
	"github.com/joho/godotenv"
)

func main() {
	// ============================================================================
	// CONFIGURATION
	// ============================================================================
	// Load .env file if it exists (optional - will use system env vars if not found)
	_ = godotenv.Load()

	// Load configuration from environment variables
	// Create a .env file or export these in your shell

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("❌ ANTHROPIC_API_KEY environment variable is required")
	}

	liminalBaseURL := os.Getenv("LIMINAL_BASE_URL")
	if liminalBaseURL == "" {
		liminalBaseURL = "https://api.liminal.cash"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ============================================================================
	// LIMINAL EXECUTOR SETUP
	// ============================================================================
	// The HTTPExecutor handles all API calls to Liminal banking services.
	// Authentication is handled automatically via JWT tokens passed from the
	// frontend login flow (email/OTP). No API key needed!

	liminalExecutor := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
		BaseURL: liminalBaseURL,
	})
	log.Println("✅ Liminal API configured")

	// ============================================================================
	// SERVER SETUP
	// ============================================================================
	// Create the nim-go-sdk server with Claude AI
	// The server handles WebSocket connections and manages conversations
	// Authentication is automatic: JWT tokens from the login flow are extracted
	// from WebSocket connections and forwarded to Liminal API calls

	srv, err := server.New(server.Config{
		AnthropicKey:    anthropicKey,
		SystemPrompt:    hackathonSystemPrompt,
		Model:           "claude-sonnet-4-20250514",
		MaxTokens:       4096,
		LiminalExecutor: liminalExecutor, // SDK automatically handles JWT extraction and forwarding
	})
	if err != nil {
		log.Fatal(err)
	}

	// ============================================================================
	// ADD LIMINAL BANKING TOOLS
	// ============================================================================
	// These are the 9 core Liminal tools that give your AI access to real banking:
	//
	// READ OPERATIONS (no confirmation needed):
	//   1. get_balance - Check wallet balance
	//   2. get_savings_balance - Check savings positions and APY
	//   3. get_vault_rates - Get current savings rates
	//   4. get_transactions - View transaction history
	//   5. get_profile - Get user profile info
	//   6. search_users - Find users by display tag
	//
	// WRITE OPERATIONS (require user confirmation):
	//   7. send_money - Send money to another user
	//   8. deposit_savings - Deposit funds into savings
	//   9. withdraw_savings - Withdraw funds from savings

	srv.AddTools(tools.LiminalTools(liminalExecutor)...)
	log.Println("✅ Added 9 Liminal banking tools")

	// ============================================================================
	// ADD CUSTOM TOOLS
	// ============================================================================
	// This is where you'll add your hackathon project's custom tools!
	// Below is an example spending analyzer tool to get you started.

	srv.AddTool(createSpendingAnalyzerTool(liminalExecutor))
	log.Println("✅ Added custom spending analyzer tool")

	// TODO: Add more custom tools here!
	// Examples:
	//   - Savings goal tracker
	//   - Budget alerts
	//   - Spending category analyzer
	//   - Bill payment predictor
	//   - Cash flow forecaster

	srv.AddTool(createMoneyPersonality(liminalExecutor))
    log.Println("✅ Added Money Personality analyzer")
	
	srv.AddTool(createCSVTransactionsTool())
	log.Println("✅ Added CSV transactions reader (for testing)")
	// ============================================================================
	// START SERVER
	// ============================================================================

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 Hackathon Starter Server Running")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("Ready for connections! Start your frontend with: cd frontend && npm run dev")
	log.Println()

	if err := srv.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ============================================================================
// CSV TRANSACTION LOADER
// ============================================================================
// This function loads transactions from a CSV file for offline testing.
// It looks for "transactions.csv" in the current directory.
//
// CSV Format:
// timestamp,type,amount,currency,counterparty,description,category,balance_after
//
// Returns a slice of transaction maps compatible with the Liminal API format.

func loadTransactionsFromCSV(filepath string) ([]map[string]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var transactions []map[string]interface{}

	// Read all rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV row: %w", err)
		}

		// Create transaction map
		tx := make(map[string]interface{})
		for i, value := range record {
			if i >= len(header) {
				break
			}
			
			columnName := header[i]
			
			// Parse specific fields appropriately
			switch columnName {
			case "amount", "balance_after":
				if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					tx[columnName] = floatVal
				} else {
					tx[columnName] = value
				}
			case "timestamp":
				tx[columnName] = value
			default:
				tx[columnName] = value
			}
		}

		transactions = append(transactions, tx)
	}

	log.Printf("✅ Loaded %d transactions from CSV file", len(transactions))
	return transactions, nil
}

// ============================================================================
// SYSTEM PROMPT
// ============================================================================
// This prompt defines your AI agent's personality and behavior
// Customize this to match your hackathon project's focus!

const hackathonSystemPrompt = `You are NeuraPay, a proactive AI wealth optimizer. You don't wait to be asked - 
you actively monitor finances and suggest optimal moves. You're like having 
a smart friend who's really good with money watching your back 24/7.

PROACTIVE BEHAVIORS:
- Greet users with their current spare cash amount
- Suggest savings moves at optimal moments
- Celebrate interest earnings and milestones
- Warn about low balances before they happen

WHAT YOU DO:
You help users manage their money using Liminal's stablecoin banking platform. You can check balances, review transactions, send money, and manage savings - all through natural conversation.

CONVERSATIONAL STYLE:
- Be warm, friendly, and conversational - not robotic
- Use casual language when appropriate, but stay professional about money
- Ask clarifying questions when something is unclear
- Remember context from earlier in the conversation
- Explain things simply without being condescending

WHEN TO USE TOOLS:
- Use tools immediately for simple queries ("what's my balance?")
- For actions, gather all required info first ("send $50 to @alice")
- Always confirm before executing money movements
- Don't use tools for general questions about how things work

MONEY MOVEMENT RULES (IMPORTANT):
- ALL money movements require explicit user confirmation
- Show a clear summary before confirming:
  * send_money: "Send $50 USD to @alice"
  * deposit_savings: "Deposit $100 USD into savings"
  * withdraw_savings: "Withdraw $50 USD from savings"
- Never assume amounts or recipients
- Always use the exact currency the user specified

AVAILABLE BANKING TOOLS:
- Check wallet balance (get_balance)
- Check savings balance and APY (get_savings_balance)
- View savings rates (get_vault_rates)
- View transaction history (get_transactions)
- Get profile info (get_profile)
- Search for users (search_users)
- Send money (send_money) - requires confirmation
- Deposit to savings (deposit_savings) - requires confirmation
- Withdraw from savings (withdraw_savings) - requires confirmation

TESTING/DEMO TOOLS:
- Read CSV transactions (get_csv_transactions) - for offline testing with transactions.csv

CUSTOM ANALYTICAL TOOLS:
- Analyze spending patterns (analyze_spending)
- Discover your Money Personality (analyze_money_personality)

TIPS FOR GREAT INTERACTIONS:
- Proactively suggest relevant actions ("Want me to move some to savings?")
- Explain the "why" behind suggestions
- Celebrate financial wins ("Nice! Your savings earned $5 this month!")
- Be encouraging about savings goals
- Make finance feel less intimidating

MONEY PERSONALITY INSIGHTS:
When users want to understand their financial psychology, use analyze_money_personality.
This isn't just data - it reveals behavioral patterns and provides personalized strategies.
Make it feel like a revelation: "Let me analyze your financial DNA..."

Remember: You're here to make banking delightful and help users build better financial habits!`

// ============================================================================
// CUSTOM TOOL: SPENDING ANALYZER
// ============================================================================
// This is an example custom tool that demonstrates how to:
// 1. Define tool parameters with JSON schema
// 2. Call other Liminal tools from within your tool OR load from CSV
// 3. Process and analyze the data
// 4. Return useful insights
//
// Use this as a template for your own hackathon tools!

func createSpendingAnalyzerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("analyze_spending").
		Description("Analyze the user's spending patterns over a specified time period. Returns insights about spending velocity, categories, and trends.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"days": tools.IntegerProperty("Number of days to analyze (default: 30)"),
			"use_csv": tools.BooleanProperty("Use local CSV file instead of API (for testing, default: false)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				Days   int  `json:"days"`
				UseCSV bool `json:"use_csv"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Default to 30 days if not specified
			if params.Days == 0 {
				params.Days = 30
			}

			var transactions []map[string]interface{}

			// STEP 1: Fetch transaction data (from CSV or API)
			if params.UseCSV {
				// Load from CSV file for testing
				csvTransactions, err := loadTransactionsFromCSV("transactions.csv")
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to load CSV: %v", err),
					}, nil
				}
				transactions = csvTransactions
			} else {
				// Fetch from Liminal API
				txRequest := map[string]interface{}{
					"limit": 100, // Get up to 100 transactions
				}
				txRequestJSON, _ := json.Marshal(txRequest)

				txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
					UserID:    toolParams.UserID,
					Tool:      "get_transactions",
					Input:     txRequestJSON,
					RequestID: toolParams.RequestID,
				})
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
					}, nil
				}

				if !txResponse.Success {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
					}, nil
				}

				// Parse transaction data from API response
				var txData map[string]interface{}
				if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
					if txArray, ok := txData["transactions"].([]interface{}); ok {
						for _, tx := range txArray {
							if txMap, ok := tx.(map[string]interface{}); ok {
								transactions = append(transactions, txMap)
							}
						}
					}
				}
			}

			// STEP 2: Analyze the data
			analysis := analyzeTransactions(transactions, params.Days)

			// STEP 3: Return insights
			result := map[string]interface{}{
				"period_days":        params.Days,
				"total_transactions": len(transactions),
				"analysis":           analysis,
				"data_source":        map[string]bool{"csv": params.UseCSV, "api": !params.UseCSV},
				"generated_at":       time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// analyzeTransactions processes transaction data and returns insights
func analyzeTransactions(transactions []map[string]interface{}, days int) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{
			"summary": "No transactions found in the specified period",
		}
	}

	// Calculate basic metrics
	var totalSpent, totalReceived float64
	var spendCount, receiveCount int
	categorySpending := make(map[string]float64)

	// Analyze each transaction
	for _, tx := range transactions {
		txType, _ := tx["type"].(string)
		amount, _ := tx["amount"].(float64)
		category, _ := tx["category"].(string)

		switch txType {
		case "send":
			totalSpent += amount
			spendCount++
			if category != "" {
				categorySpending[category] += amount
			}
		case "receive":
			totalReceived += amount
			receiveCount++
		}
	}

	avgDailySpend := totalSpent / float64(days)
	
	// Find top spending categories
	type categoryTotal struct {
		Category string
		Amount   float64
	}
	var categories []categoryTotal
	for cat, amt := range categorySpending {
		categories = append(categories, categoryTotal{Category: cat, Amount: amt})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Amount > categories[j].Amount
	})

	// Build category breakdown
	topCategories := make(map[string]string)
	for i, cat := range categories {
		if i >= 5 { // Top 5 categories
			break
		}
		topCategories[cat.Category] = fmt.Sprintf("$%.2f", cat.Amount)
	}

	return map[string]interface{}{
		"total_spent":       fmt.Sprintf("%.2f", totalSpent),
		"total_received":    fmt.Sprintf("%.2f", totalReceived),
		"net_cashflow":      fmt.Sprintf("%.2f", totalReceived-totalSpent),
		"spend_count":       spendCount,
		"receive_count":     receiveCount,
		"avg_daily_spend":   fmt.Sprintf("%.2f", avgDailySpend),
		"velocity":          calculateVelocity(spendCount, days),
		"top_categories":    topCategories,
		"insights": []string{
			fmt.Sprintf("You made %d spending transactions over %d days", spendCount, days),
			fmt.Sprintf("Average daily spend: $%.2f", avgDailySpend),
			fmt.Sprintf("Net cash flow: $%.2f", totalReceived-totalSpent),
			"Consider setting up savings goals to build financial cushion",
		},
	}
}

// calculateVelocity determines spending frequency
func calculateVelocity(transactionCount, days int) string {
	txPerWeek := float64(transactionCount) / float64(days) * 7

	switch {
	case txPerWeek < 2:
		return "low"
	case txPerWeek < 7:
		return "moderate"
	default:
		return "high"
	}
}

// ============================================================================
// CUSTOM TOOL: MONEY PERSONALITY ANALYZER
// ============================================================================

type PersonalityScore struct {
	Name  string
	Value float64
}

type PersonalityArchetype struct {
	Type       string
	Emoji      string
	Confidence float64
	Traits     []string
	Triggers   []string
	Strategies []string
	FunFact    string
}

func createMoneyPersonality(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("analyze_money_personality").
		Description("Discover your Money Personality - a psychological profile of your spending and saving behaviors. Reveals behavioral patterns, triggers, and personalized strategies.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"use_csv": tools.BooleanProperty("Use local CSV file instead of API (for testing, default: false)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				UseCSV bool `json:"use_csv"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			var transactions []map[string]interface{}

			// Fetch transaction data (from CSV or API)
			if params.UseCSV {
				csvTransactions, err := loadTransactionsFromCSV("transactions.csv")
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to load CSV: %v", err),
					}, nil
				}
				transactions = csvTransactions
			} else {
				txRequest := map[string]interface{}{"limit": 100}
				txRequestJSON, _ := json.Marshal(txRequest)

				txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
					UserID:    toolParams.UserID,
					Tool:      "get_transactions",
					Input:     txRequestJSON,
					RequestID: toolParams.RequestID,
				})
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
					}, nil
				}

				if !txResponse.Success {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
					}, nil
				}

				var txData map[string]interface{}
				if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
					if txArray, ok := txData["transactions"].([]interface{}); ok {
						for _, tx := range txArray {
							if txMap, ok := tx.(map[string]interface{}); ok {
								transactions = append(transactions, txMap)
							}
						}
					}
				}
			}

			if len(transactions) < 10 {
				return &core.ToolResult{
					Success: false,
					Error:   "Need at least 10 transactions for accurate personality analysis",
				}, nil
			}

			// Calculate personality scores
			scores := calculatePersonalityScores(transactions)
			archetype := matchArchetype(scores)

			result := map[string]interface{}{
				"personality_type": archetype.Type,
				"emoji":            archetype.Emoji,
				"confidence":       fmt.Sprintf("%.0f%%", archetype.Confidence*100),
				"traits":           archetype.Traits,
				"behavioral_triggers": archetype.Triggers,
				"personalized_strategies": archetype.Strategies,
				"fun_fact":         archetype.FunFact,
				"raw_scores":       scores,
				"data_source":      map[string]bool{"csv": params.UseCSV, "api": !params.UseCSV},
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

func calculatePersonalityScores(transactions []map[string]interface{}) map[string]float64 {
	scores := make(map[string]float64)
	
	var amounts []float64
	var balances []float64
	incomeCount := 0
	totalIncome := 0.0
	totalSpend := 0.0
	savingsTransactions := 0
	
	categorySpend := make(map[string]float64)
	
	for _, tx := range transactions {
		txType, _ := tx["type"].(string)
		amount, _ := tx["amount"].(float64)
		category, _ := tx["category"].(string)
		balance, _ := tx["balance_after"].(float64)
		
		if txType == "send" {
			amounts = append(amounts, amount)
			totalSpend += amount
			categorySpend[category] += amount
			
			if category == "savings" {
				savingsTransactions++
			}
		} else if txType == "receive" {
			incomeCount++
			totalIncome += amount
		}
		
		if balance > 0 {
			balances = append(balances, balance)
		}
	}
	
	// 1. Transaction Velocity (0-100)
	txPerWeek := float64(len(transactions)) / 4.0 // Assuming ~4 weeks of data
	scores["transaction_velocity"] = math.Min(txPerWeek*10, 100)
	
	// 2. Amount Distribution (0-100) - measures consistency
	if len(amounts) > 0 {
		variance := calculateVariance(amounts)
		mean := calculateMean(amounts)
		cv := 0.0
		if mean > 0 {
			cv = (math.Sqrt(variance) / mean) * 100
		}
		scores["amount_distribution"] = math.Min(cv, 100)
	}
	
	// 3. Balance Comfort (0-100)
	if len(balances) > 0 {
		avgBalance := calculateMean(balances)
		minBalance := balances[0]
		for _, b := range balances {
			if b < minBalance {
				minBalance = b
			}
		}
		bufferRatio := 0.0
		if avgBalance > 0 {
			bufferRatio = (minBalance / avgBalance) * 100
		}
		scores["balance_comfort"] = math.Min(bufferRatio, 100)
	}
	
	// 4. Savings Affinity (0-100)
	savingsRate := 0.0
	if len(transactions) > 0 {
		savingsRate = (float64(savingsTransactions) / float64(len(transactions))) * 100 * 3 // Amplify
	}
	scores["savings_affinity"] = math.Min(savingsRate, 100)
	
	// 5. Income Response (0-100) - spending surge after income
	scores["income_response"] = 50.0 // Placeholder - would need temporal analysis
	
	return scores
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := calculateMean(values)
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	return variance / float64(len(values))
}

func matchArchetype(scores map[string]float64) PersonalityArchetype {
	type archetype struct {
		name       string
		emoji      string
		matcher    func(map[string]float64) float64
		traits     []string
		triggers   []string
		strategies []string
		funFact    string
	}
	
	archetypes := []archetype{
		{
			name:  "The Reward Seeker",
			emoji: "🎉",
			matcher: func(s map[string]float64) float64 {
				return s["transaction_velocity"]*0.4 + (100-s["savings_affinity"])*0.3 + s["income_response"]*0.3
			},
			traits: []string{
				"You spend to celebrate and feel good",
				"Money is a tool for experiences and pleasure",
				"High transaction frequency - lots of small treats",
				"Impulsive but not reckless",
			},
			triggers: []string{
				"Income hits = immediate 'treat yourself' urge",
				"Stress or bad day triggers comfort spending",
				"Social occasions: primary spending driver",
			},
			strategies: []string{
				"Auto-save 20% BEFORE you see your paycheck (out of sight, out of mind)",
				"Keep a visible 'celebration budget' so treats don't feel restricted",
				"Gamify savings: every $500 saved = unlock a $50 reward",
				"Schedule 'mini celebrations' that cost $0 (movie night at home, etc.)",
			},
			funFact: "Reward Seekers save 47% more when savings feel like 'winning' rather than 'restricting'. Your brain needs the dopamine hit!",
		},
		{
			name:  "The Safety Hoarder",
			emoji: "🛡️",
			matcher: func(s map[string]float64) float64 {
				return s["balance_comfort"]*0.4 + (100-s["transaction_velocity"])*0.3 + s["savings_affinity"]*0.3
			},
			traits: []string{
				"You maintain a high balance buffer at all times",
				"Low transaction frequency - you think before spending",
				"Money anxiety drives conservative behavior",
				"'What if' scenarios dominate your financial decisions",
			},
			triggers: []string{
				"Balance dipping below comfort threshold triggers stress",
				"Unexpected expenses cause disproportionate anxiety",
				"You delay purchases waiting for 'the right time'",
			},
			strategies: []string{
				"Calculate your TRUE minimum (3 months expenses) and relax about the rest",
				"Move excess beyond safety threshold to high-yield savings",
				"Set up 'if-then' rules: IF balance > $X, THEN auto-move to savings",
				"Track what you DON'T spend vs what you do (flip the anxiety narrative)",
			},
			funFact: "Safety Hoarders often sit on $5,000+ earning 0% interest when their actual safety threshold is $2,000. You're losing $200+/year to fear!",
		},
		{
			name:  "The Impulse Optimizer",
			emoji: "⚡",
			matcher: func(s map[string]float64) float64 {
				return s["transaction_velocity"]*0.4 + (100-s["amount_distribution"])*0.3 + (100-s["savings_affinity"])*0.3
			},
			traits: []string{
				"High transaction frequency - many small purchases",
				"Convenience over cost is your philosophy",
				"You optimize for time and ease, not dollars",
				"Spending is habitual and automatic",
			},
			triggers: []string{
				"Daily coffee/food runs add up to 30% of spending",
				"One-click purchase features are dangerous",
				"'Just this once' happens 5+ times per week",
			},
			strategies: []string{
				"Add friction: 24-hour delay for purchases over $25",
				"Round-up savings: auto-save the 'change' from each transaction",
				"Batch purchases: weekly grocery trip instead of daily stops",
				"Make saving the path of least resistance (auto-transfer on payday)",
			},
			funFact: "Impulse Optimizers spend 40% more on convenience purchases than they estimate. Your $4 coffee habit is actually $8/day when you count the muffin!",
		},
		{
			name:  "The Cyclical Spender",
			emoji: "🌊",
			matcher: func(s map[string]float64) float64 {
				return s["amount_distribution"]*0.4 + s["income_response"]*0.3 + (100-s["balance_comfort"])*0.3
			},
			traits: []string{
				"Boom-bust spending cycles dominate your pattern",
				"Large irregular transactions mixed with quiet periods",
				"Emotional state drives financial decisions",
				"Balance swings wildly month to month",
			},
			triggers: []string{
				"Stress or celebration both trigger spending sprees",
				"'Flush with cash' feeling leads to overshooting",
				"Low balance periods create panic and restriction",
			},
			strategies: []string{
				"Income smoothing: divide monthly income into weekly 'paychecks'",
				"Create artificial scarcity: move money OUT immediately",
				"Separate accounts: one for bills, one for discretionary, one for savings",
				"Track cycles and predict them (you're more regular than you think)",
			},
			funFact: "Cyclical Spenders have the most to gain from automation. Smoothing your income into weekly distributions can cut overspending by 60%!",
		},
		{
			name:  "The Strategic Planner",
			emoji: "🎯",
			matcher: func(s map[string]float64) float64 {
				return (100-s["amount_distribution"])*0.3 + s["savings_affinity"]*0.3 + (100-s["income_response"])*0.2 + s["balance_comfort"]*0.2
			},
			traits: []string{
				"Consistent, predictable spending patterns",
				"High savings rate without much effort",
				"You're already optimized - low variation in behavior",
				"Natural financial discipline",
			},
			triggers: []string{
				"Rare - you don't have strong triggers",
				"Unusual expenses are planned and budgeted",
				"You think ahead and avoid surprises",
			},
			strategies: []string{
				"Maximize interest arbitrage - you have the discipline",
				"Explore tax optimization and advanced strategies",
				"Consider investing surplus rather than just saving",
				"Help others - your natural skills could benefit friends",
			},
			funFact: "Strategic Planners are rare (only 12% of people). Your challenge isn't saving more - it's not becoming too rigid. Allow yourself some spontaneity!",
		},
	}

	// Score each archetype
	bestMatch := archetypes[0]
	bestScore := 0.0

	for _, archetype := range archetypes {
		score := archetype.matcher(scores)
		if score > bestScore {
			bestScore = score
			bestMatch = archetype
		}
	}

	// Calculate confidence (how much better is best match vs second best)
	var sortedScores []float64
	for _, archetype := range archetypes {
		sortedScores = append(sortedScores, archetype.matcher(scores))
	}
	sort.Float64s(sortedScores)
	
	confidence := 0.7 // Default confidence
	if len(sortedScores) >= 2 {
		diff := sortedScores[len(sortedScores)-1] - sortedScores[len(sortedScores)-2]
		confidence = math.Min(0.5+diff/100, 0.95)
	}

	return PersonalityArchetype{
		Type:       bestMatch.name,
		Emoji:      bestMatch.emoji,
		Confidence: confidence,
		Traits:     bestMatch.traits,
		Triggers:   bestMatch.triggers,
		Strategies: bestMatch.strategies,
		FunFact:    bestMatch.funFact,
	}
}

// ============================================================================
// CUSTOM TOOL: CSV TRANSACTIONS READER
// ============================================================================
// This tool reads transactions directly from CSV for testing/demo purposes
// It's like get_transactions but works offline with local data

func createCSVTransactionsTool() core.Tool {
	return tools.New("get_csv_transactions").
		Description("Read transactions from the local transactions.csv file. Use this for testing when API is unavailable or you want to use demo data.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"limit": tools.IntegerProperty("Maximum number of transactions to return (default: 50)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Limit int `json:"limit"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			if params.Limit == 0 {
				params.Limit = 50
			}

			// Load transactions from CSV
			transactions, err := loadTransactionsFromCSV("transactions.csv")
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to load transactions from CSV: %v", err),
				}, nil
			}

			// Limit results if requested
			if len(transactions) > params.Limit {
				transactions = transactions[:params.Limit]
			}

			result := map[string]interface{}{
				"transactions": transactions,
				"count":        len(transactions),
				"source":       "csv",
				"file":         "transactions.csv",
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}




// ============================================================================
// HACKATHON IDEAS
// ============================================================================
// Here are some ideas for custom tools you could build:
//
// 1. SAVINGS GOAL TRACKER
//    - Track progress toward savings goals
//    - Calculate how long until goal is reached
//    - Suggest optimal deposit amounts
//
// 2. BUDGET ANALYZER
//    - Set spending limits by category
//    - Alert when approaching limits
//    - Compare actual vs. planned spending
//
// 3. RECURRING PAYMENT DETECTOR
//    - Identify subscription payments
//    - Warn about upcoming bills
//    - Suggest savings opportunities
//
// 4. CASH FLOW FORECASTER
//    - Predict future balance based on patterns
//    - Identify potential low balance periods
//    - Suggest when to save vs. spend
//
// 5. SMART SAVINGS ADVISOR
//    - Analyze spare cash available
//    - Recommend savings deposits
//    - Calculate interest projections
//
// 6. SPENDING INSIGHTS
//    - Categorize spending automatically
//    - Compare to typical user patterns
//    - Highlight unusual activity
//
// 7. FINANCIAL HEALTH SCORE
//    - Calculate overall financial wellness
//    - Track improvements over time
//    - Provide actionable recommendations
//
// 8. PEER COMPARISON (anonymous)
//    - Compare savings rate to anonymized peers
//    - Show percentile rankings
//    - Motivate better habits
//
// 9. TAX ESTIMATION
//    - Track potential tax obligations
//    - Suggest amounts to set aside
//    - Generate tax reports
//
// 10. EMERGENCY FUND BUILDER
//     - Calculate needed emergency fund size
//     - Track progress toward goal
//     - Suggest automated savings plan
//
// ============================================================================
