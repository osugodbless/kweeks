package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
)

// DashboardStat is the wallet landing page headline numbers.
type DashboardStat struct {
	QuizzesHosted  int         `json:"quizzesHosted"`
	PlayersHosted  int         `json:"playersHosted"`
	WinnersPaid    int         `json:"winnersPaid"`
	AvailableNaira string      `json:"availableNaira"`
	Quizzes        []QuizBrief `json:"quizzes"`
}

// QuizBrief is a dashboard quiz row, optionally with its active room.
type QuizBrief struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	PoolNaira     string    `json:"poolNaira"`
	WinnerCount   int       `json:"winnerCount"`
	Pacing        string    `json:"pacing"`
	QuestionCount int       `json:"questionCount"`
	CreatedAt     time.Time `json:"createdAt"`
	RoomCode      string    `json:"roomCode,omitempty"`
	RoomState     string    `json:"state,omitempty"`
	RoomID        string    `json:"roomId,omitempty"`
}

// HistoryItem is one row of the instructor history feed.
type HistoryItem struct {
	ID          string    `json:"id"`
	At          time.Time `json:"at"`
	Type        string    `json:"type"` // fund|quiz|payout|room
	Title       string    `json:"title"`
	AmountNaira string    `json:"amountNaira,omitempty"`
	Meta        string    `json:"meta,omitempty"`
	State       string    `json:"state,omitempty"`
}

// Dashboard assembles instructor headline stats + quiz rows.
func (g *Game) Dashboard(ctx context.Context, instructorID string) (*DashboardStat, error) {
	quizzes, err := g.store.ListQuizzes(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	stat := &DashboardStat{QuizzesHosted: len(quizzes)}

	quizIDs := make([]string, 0, len(quizzes))
	playerSeen := map[string]bool{}
	for i := range quizzes {
		q := &quizzes[i]
		quizIDs = append(quizIDs, q.ID)
		stat.Quizzes = append(stat.Quizzes, QuizBrief{
			ID: q.ID, Title: q.Title, PoolNaira: q.Pool.DisplayString(),
			WinnerCount: q.WinnerCount, Pacing: string(q.Pacing),
			QuestionCount: len(q.Questions), CreatedAt: q.CreatedAt,
		})
		// Player + active-room aggregates.
		if room, err := g.store.LatestLiveRoom(ctx, q.ID); err == nil && room != nil {
			players, _ := g.store.ListParticipants(ctx, room.ID)
			for _, p := range players {
				playerSeen[p.ID] = true
			}
			brief := &stat.Quizzes[len(stat.Quizzes)-1]
			brief.RoomID = room.ID
			brief.RoomCode = room.Code
			brief.RoomState = string(room.State)
		}
	}
	stat.PlayersHosted = len(playerSeen)

	claims, err := g.store.ListClaimsByQuizIDs(ctx, quizIDs)
	if err == nil {
		for _, c := range claims {
			if c.State == domain.ClaimPaid {
				stat.WinnersPaid++
			}
		}
	}
	wallet, err := g.store.GetWalletByInstructor(ctx, instructorID)
	if err == nil && wallet != nil {
		stat.AvailableNaira = wallet.Balance.DisplayString()
	}
	return stat, nil
}

// History returns the unified instructor feed: wallet funding + quiz hosting +
// paid winners, newest first. A fresh instructor with no activity gets an
// empty slice (frontend shows the empty state).
func (g *Game) History(ctx context.Context, instructorID string) ([]HistoryItem, error) {
	quizzes, err := g.store.ListQuizzes(ctx, instructorID)
	if err != nil {
		return []HistoryItem{}, nil
	}
	quizIDs := make([]string, 0, len(quizzes))
	for i := range quizzes {
		quizIDs = append(quizIDs, quizzes[i].ID)
	}
	claims, _ := g.store.ListClaimsByQuizIDs(ctx, quizIDs)

	wallet, _ := g.store.GetWalletByInstructor(ctx, instructorID)
	var walletTx []domain.WalletTransaction
	if wallet != nil {
		walletTx, _ = g.store.ListWalletTransactions(ctx, wallet.ID)
	}

	out := make([]HistoryItem, 0, len(walletTx)+len(quizzes)+len(claims))

	// Wallet funding first (kind fund).
	for _, t := range walletTx {
		out = append(out, HistoryItem{
			ID: t.ID, At: t.CreatedAt, Type: string(t.Kind), Title: t.Note,
			AmountNaira: t.Amount.DisplayString(),
		})
	}
	// Quizzes hosted.
	for i := range quizzes {
		q := quizzes[i]
		out = append(out, HistoryItem{
			ID: q.ID, At: q.CreatedAt, Type: "quiz", Title: "Hosted · " + q.Title,
			Meta: poolSummary(q),
		})
		if room, err := g.store.LatestLiveRoom(ctx, q.ID); err == nil && room != nil {
			// room opened after hosting
		}
	}
	// Paid winners.
	for _, c := range claims {
		if c.State != domain.ClaimPaid {
			continue
		}
		out = append(out, HistoryItem{
			ID: c.ID, At: c.CreatedAt, Type: "payout",
			Title:       "Paid winners · " + c.Email,
			AmountNaira: c.Amount.DisplayString(), State: "Paid",
		})
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out, nil
}

func poolSummary(q domain.Quiz) string {
	return formatPoolMeta(q)
}

func formatPoolMeta(q domain.Quiz) string {
	return fmt.Sprintf("%d winner(s) · %s pool", q.WinnerCount, q.Pool.DisplayString())
}
