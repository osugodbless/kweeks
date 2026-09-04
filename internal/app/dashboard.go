package app

import (
	"context"
	"fmt"
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

// History returns the unified instructor feed. Order: wallet funding entries,
// then quizzes (with an active room), then paid winners, then rooms.
func (g *Game) History(ctx context.Context, instructorID string) ([]HistoryItem, error) {
	wallet, err := g.store.GetWalletByInstructor(ctx, instructorID)
	if err == nil && wallet != nil {
		txs, _ := g.store.ListWalletTransactions(ctx, wallet.ID)
		out := make([]HistoryItem, 0, len(txs)+4)
		for _, t := range txs {
			kind := string(t.Kind)
			title := t.Note
			meta := ""
			if t.Kind == domain.TxFund {
				title = "Funded wallet · " + fundingKindLabel(t.Note)
			}
			out = append(out, HistoryItem{
				ID: t.ID, At: t.CreatedAt, Type: kind, Title: title,
				AmountNaira: t.Amount.DisplayString(),
				Meta:        meta,
			})
		}
		return out, nil
	}

	// No wallet tx (fresh account): surface quizzes + paid winners as a fallback
	// feed so the History empty state still has honest content.
	quizzes, err := g.store.ListQuizzes(ctx, instructorID)
	if err != nil {
		return []HistoryItem{}, nil
	}
	out := []HistoryItem{}
	quizIDs := make([]string, 0, len(quizzes))
	for i := range quizzes {
		quizIDs = append(quizIDs, quizzes[i].ID)
		out = append(out, HistoryItem{
			ID: quizzes[i].ID, At: quizzes[i].CreatedAt, Type: "quiz",
			Title: "Hosted · " + quizzes[i].Title,
			Meta:  poolSummary(quizzes[i]),
		})
	}
	if claims, err := g.store.ListClaimsByQuizIDs(ctx, quizIDs); err == nil {
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
	}
	return out, nil
}

func fundingKindLabel(note string) string {
	switch note {
	case "Funded wallet · card":
		return "card"
	case "Funded wallet · transfer":
		return "transfer"
	default:
		return "instant credit"
	}
}

func poolSummary(q domain.Quiz) string {
	return formatPoolMeta(q)
}

func formatPoolMeta(q domain.Quiz) string {
	return fmt.Sprintf("%d winner(s) · %s pool", q.WinnerCount, q.Pool.DisplayString())
}
