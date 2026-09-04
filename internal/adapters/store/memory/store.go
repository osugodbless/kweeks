// Package memory provides an in-memory Store implementation. It is the test
// double AND the demo fallback: no Postgres required to run the show.
package memory

import (
	"context"
	"sync"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Store is a thread-safe in-memory implementation of ports.Store.
type Store struct {
	mu           sync.RWMutex
	quizzes      map[string]*domain.Quiz
	rooms        map[string]*domain.Room
	participants map[string][]*domain.Participant // key roomID
	answers      map[string][]*domain.Answer      // key roomID
	claims       map[string][]*domain.Claim       // key quizID

	instructors map[string]*domain.Instructor          // by id
	sessions    map[string]*domain.Session             // by token
	wallets     map[string]*domain.Wallet              // by id
	walletByIns map[string]string                      // instructorID -> walletID
	walletTx    map[string][]*domain.WalletTransaction // by walletID
}

func New() *Store {
	return &Store{
		quizzes:      map[string]*domain.Quiz{},
		rooms:        map[string]*domain.Room{},
		participants: map[string][]*domain.Participant{},
		answers:      map[string][]*domain.Answer{},
		claims:       map[string][]*domain.Claim{},
		instructors:  map[string]*domain.Instructor{},
		sessions:     map[string]*domain.Session{},
		wallets:      map[string]*domain.Wallet{},
		walletByIns:  map[string]string{},
		walletTx:     map[string][]*domain.WalletTransaction{},
	}
}

func (s *Store) CreateQuiz(ctx context.Context, q *domain.Quiz) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quizzes[q.ID] = cloneQuiz(q)
	return nil
}

func (s *Store) GetQuiz(ctx context.Context, id string) (*domain.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q, ok := s.quizzes[id]
	if !ok {
		return nil, domain.ErrQuizNotFound
	}
	return cloneQuiz(q), nil
}

func (s *Store) ListQuizzes(ctx context.Context, instructorID string) ([]domain.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Quiz
	for _, q := range s.quizzes {
		if q.InstructorID == instructorID {
			out = append(out, *cloneQuiz(q))
		}
	}
	return out, nil
}

func (s *Store) CreateRoom(ctx context.Context, r *domain.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *r
	s.rooms[r.ID] = &cp
	return nil
}

func (s *Store) GetRoom(ctx context.Context, id string) (*domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[id]
	if !ok {
		return nil, domain.ErrRoomNotFound
	}
	cp := *r
	return &cp, nil
}

func (s *Store) SaveRoom(ctx context.Context, r *domain.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rooms[r.ID]; !ok {
		return domain.ErrRoomNotFound
	}
	cp := *r
	s.rooms[r.ID] = &cp
	return nil
}

func (s *Store) ListLiveRooms(ctx context.Context) ([]domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Room
	for _, r := range s.rooms {
		if r.State == domain.RoomLive {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (s *Store) GetRoomByCode(ctx context.Context, code string) (*domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rooms {
		if r.Code == code {
			cp := *r
			return &cp, nil
		}
	}
	return nil, domain.ErrRoomNotFound
}

func (s *Store) ListRoomsByHost(ctx context.Context, hostID string) ([]domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Room
	for _, r := range s.rooms {
		if r.HostID == hostID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (s *Store) LatestLiveRoom(ctx context.Context, quizID string) (*domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var best *domain.Room
	for _, r := range s.rooms {
		if r.QuizID != quizID {
			continue
		}
		if r.State != domain.RoomLobby && r.State != domain.RoomLive {
			continue
		}
		if best == nil || r.StartedAt.After(best.StartedAt) {
			cp := *r
			best = &cp
		}
	}
	if best == nil {
		return nil, domain.ErrRoomNotFound
	}
	return best, nil
}

func (s *Store) JoinParticipant(ctx context.Context, p *domain.Participant) (*domain.Participant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.participants[p.RoomID] {
		if existing.Email == p.Email {
			cp := *existing
			return &cp, domain.ErrDuplicateParticipant
		}
	}
	cp := *p
	s.participants[p.RoomID] = append(s.participants[p.RoomID], &cp)
	return &cp, nil
}

func (s *Store) GetParticipant(ctx context.Context, roomID, email string) (*domain.Participant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.participants[roomID] {
		if p.Email == email {
			cp := *p
			return &cp, nil
		}
	}
	return nil, domain.ErrParticipantNotFound
}

func (s *Store) ListParticipants(ctx context.Context, roomID string) ([]domain.Participant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ps := s.participants[roomID]
	out := make([]domain.Participant, 0, len(ps))
	for _, p := range ps {
		out = append(out, *p)
	}
	return out, nil
}

func (s *Store) RecordAnswer(ctx context.Context, a *domain.Answer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.answers[a.RoomID] = append(s.answers[a.RoomID], cloneAnswer(a))
	return nil
}

func (s *Store) HasAnswered(ctx context.Context, roomID, participantID, questionID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.answers[roomID] {
		if a.ParticipantID == participantID && a.QuestionID == questionID {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) ListAnswers(ctx context.Context, roomID string) ([]domain.Answer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	as := s.answers[roomID]
	out := make([]domain.Answer, 0, len(as))
	for _, a := range as {
		out = append(out, *cloneAnswer(a))
	}
	return out, nil
}

func (s *Store) CreateClaim(ctx context.Context, c *domain.Claim) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.claims[c.QuizID] {
		if existing.Email == c.Email {
			return domain.ErrClaimExists
		}
	}
	cp := *c
	s.claims[c.QuizID] = append(s.claims[c.QuizID], &cp)
	return nil
}

func (s *Store) GetClaimByCode(ctx context.Context, quizID, code string) (*domain.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.claims[quizID] {
		if c.ClaimCode == code {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrClaimNotFound
}

func (s *Store) GetClaimByEmail(ctx context.Context, quizID, email string) (*domain.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.claims[quizID] {
		if c.Email == email {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrClaimNotFound
}

func (s *Store) ListClaims(ctx context.Context, quizID string) ([]domain.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs := s.claims[quizID]
	out := make([]domain.Claim, 0, len(cs))
	for _, c := range cs {
		out = append(out, *c)
	}
	return out, nil
}

func (s *Store) UpdateClaimState(ctx context.Context, id string, to domain.ClaimState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cs := range s.claims {
		for _, c := range cs {
			if c.ID == id {
				if !domain.CanTransition(c.State, to) {
					return domain.ErrInvalidTransition
				}
				c.State = to
				return nil
			}
		}
	}
	return domain.ErrClaimNotFound
}

func (s *Store) Close() error { return nil }

// --- Instructors ---

func (s *Store) CreateInstructor(ctx context.Context, i *domain.Instructor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.instructors {
		if existing.Email == i.Email {
			return domain.ErrEmailTaken
		}
	}
	cp := *i
	s.instructors[cp.ID] = &cp
	return nil
}

func (s *Store) GetInstructorByEmail(ctx context.Context, email string) (*domain.Instructor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, i := range s.instructors {
		if i.Email == email {
			cp := *i
			return &cp, nil
		}
	}
	return nil, domain.ErrInstructorNotFound
}

func (s *Store) GetInstructor(ctx context.Context, id string) (*domain.Instructor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.instructors[id]
	if !ok {
		return nil, domain.ErrInstructorNotFound
	}
	cp := *i
	return &cp, nil
}

// --- Sessions ---

func (s *Store) CreateSession(ctx context.Context, sess *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *sess
	s.sessions[cp.Token] = &cp
	return nil
}

func (s *Store) GetSession(ctx context.Context, token string) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	se, ok := s.sessions[token]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	cp := *se
	return &cp, nil
}

// --- Wallets ---

func (s *Store) CreateWallet(ctx context.Context, w *domain.Wallet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.walletByIns[w.InstructorID]; exists {
		return domain.ErrWalletExists
	}
	cp := *w
	s.wallets[cp.ID] = &cp
	s.walletByIns[cp.InstructorID] = cp.ID
	return nil
}

func (s *Store) GetWalletByInstructor(ctx context.Context, instructorID string) (*domain.Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	wid, ok := s.walletByIns[instructorID]
	if !ok {
		return nil, domain.ErrWalletNotFound
	}
	w := s.wallets[wid]
	if w == nil {
		return nil, domain.ErrWalletNotFound
	}
	cp := *w
	return &cp, nil
}

func (s *Store) ApplyWalletTx(ctx context.Context, walletID string, tx *domain.WalletTransaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.wallets[walletID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	newBal := w.Balance + tx.Amount
	if newBal < 0 {
		return domain.ErrInsufficientBalance
	}
	w.Balance = newBal
	cp := *tx
	s.walletTx[walletID] = append(s.walletTx[walletID], &cp)
	return nil
}

func (s *Store) ListWalletTransactions(ctx context.Context, walletID string) ([]domain.WalletTransaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	txs := s.walletTx[walletID]
	out := make([]domain.WalletTransaction, 0, len(txs))
	for _, t := range txs {
		out = append(out, *t)
	}
	return out, nil
}

func (s *Store) ListClaimsByQuizIDs(ctx context.Context, quizIDs []string) ([]domain.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	want := map[string]bool{}
	for _, id := range quizIDs {
		want[id] = true
	}
	var out []domain.Claim
	for qid, cs := range s.claims {
		if !want[qid] {
			continue
		}
		for _, c := range cs {
			out = append(out, *c)
		}
	}
	return out, nil
}

func cloneQuiz(q *domain.Quiz) *domain.Quiz {
	if q == nil {
		return nil
	}
	cp := *q
	cp.Questions = append([]domain.Question(nil), q.Questions...)
	return &cp
}

func cloneAnswer(a *domain.Answer) *domain.Answer {
	if a == nil {
		return nil
	}
	cp := *a
	return &cp
}
