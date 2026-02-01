# NeuraPay – AI Financial Agent powered by Liminal & Nim

**Hackathon Starter Kit**  
Proactive AI wealth optimizer that lives in your pocket  
Built with [nim-go-sdk](https://github.com/becomeliminal/nim-go-sdk) + Liminal stablecoin banking APIs

<p align="center">
  <img src="https://via.placeholder.com/1200x400/0d1117/58a6ff?text=NeuraPay+-+Your+AI+Money+Friend" alt="NeuraPay Banner" />
</p>

## ✨ What is NeuraPay?

NeuraPay is a **conversational AI financial agent** that:

- Proactively watches your money 24/7
- Greets you with your current spare cash
- Suggests smart savings moves at optimal moments
- Analyzes your **spending patterns** and **money personality**
- Helps send money, deposit/withdraw to savings — all via natural chat
- Celebrates financial wins and warns before trouble

It uses **Claude** (Anthropic) as the reasoning engine and connects to real **stablecoin banking infrastructure** via Liminal.

## 🚀 Features (Current)

- Real-time balance & savings checking
- Transaction history viewing
- Send money to other users (@tags)
- Deposit / withdraw from high-yield savings vaults
- **Spending analyzer** tool (categories, velocity, trends)
- **Money Personality** analyzer (Reward Seeker, Safety Hoarder, etc.)
- Offline testing mode using `transactions.csv`
- WebSocket-based chat interface (ready for React/Vue frontend)

## 🛠️ Tech Stack

- **Backend**: Go + [nim-go-sdk](https://github.com/becomeliminal/nim-go-sdk)
- **AI**: Claude Sonnet (via Anthropic API)
- **Banking**: [Liminal](https://liminal.cash) stablecoin banking APIs
  - Wallet, savings vaults, P2P transfers, user search...
- **Auth**: JWT passed via WebSocket (handled by nim SDK)
- **Frontend** (not included): any WebSocket client (example: React + Vite)

## Quick Start

### 1. Prerequisites

- Go 1.21+
- Anthropic API key
- (Optional) Liminal developer account / test credentials

### 2. Environment Variables (.env)

```bash
# Required
ANTHROPIC_API_KEY=sk-ant-...

# Optional – defaults provided
LIMINAL_BASE_URL=https://api.liminal.cash
PORT=8080
