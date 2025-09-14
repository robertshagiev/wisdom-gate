package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Protocol commands
const (
	REQ = "REQ" // Запрос challenge
	CHL = "CHL" // Challenge от сервера
	RES = "RES" // Решение challenge
	QOT = "QOT" // Цитата от сервера
	ERR = "ERR" // Ошибка
)

type Message struct {
	Command string
	Body    string
}

func (m *Message) String() string {
	if m.Body == "" {
		return m.Command
	}
	return fmt.Sprintf("%s %d |%s", m.Command, len(m.Body), m.Body)
}

func WriteMessage(w net.Conn, msg *Message) error {
	_, err := fmt.Fprintf(w, "%s\n", msg.String())
	return err
}

func ReadMessage(r *bufio.Reader) (*Message, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)

	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 1 {
		return &Message{Command: parts[0]}, nil
	}

	bodyPart := parts[1]
	bodyParts := strings.SplitN(bodyPart, " |", 2)
	if len(bodyParts) != 2 {
		return nil, fmt.Errorf("invalid message format")
	}

	var length int
	_, err = fmt.Sscanf(bodyParts[0], "%d", &length)
	if err != nil {
		return nil, fmt.Errorf("invalid body length: %w", err)
	}

	body := bodyParts[1]
	if len(body) != length {
		return nil, fmt.Errorf("body length mismatch: expected %d, got %d", length, len(body))
	}

	return &Message{
		Command: parts[0],
		Body:    body,
	}, nil
}

type HashcashHeader struct {
	Version    int
	Difficulty int
	ExpiresAt  int64
	Subject    string
	Algorithm  string
	Nonce      string
	Counter    int64
}

func (h *HashcashHeader) String() string {
	if h.Counter == 0 {
		return fmt.Sprintf("%d:%d:%d:%s:%s:%s",
			h.Version, h.Difficulty, h.ExpiresAt, h.Subject, h.Algorithm, h.Nonce)
	}

	// Counter должен быть в base64 формате для сервера
	counterBase64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", h.Counter)))
	return fmt.Sprintf("%d:%d:%d:%s:%s:%s:%s",
		h.Version, h.Difficulty, h.ExpiresAt, h.Subject, h.Algorithm, h.Nonce, counterBase64)
}

func ParseHashcashHeader(header string) (*HashcashHeader, error) {
	parts := strings.Split(header, ":")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid header format")
	}

	var version, difficulty int
	var expiresAt int64
	var subject, algorithm, nonce string

	if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	if _, err := fmt.Sscanf(parts[1], "%d", &difficulty); err != nil {
		return nil, fmt.Errorf("invalid difficulty: %w", err)
	}

	if _, err := fmt.Sscanf(parts[2], "%d", &expiresAt); err != nil {
		return nil, fmt.Errorf("invalid expiresAt: %w", err)
	}

	// Subject может содержать IP:PORT, поэтому объединяем части
	subject = parts[3]
	if len(parts) > 6 {
		// Если есть порт, объединяем IP:PORT
		subject = parts[3] + ":" + parts[4]
		algorithm = parts[5]
		nonce = parts[6]
	} else {
		algorithm = parts[4]
		nonce = parts[5]
	}

	h := &HashcashHeader{
		Version:    version,
		Difficulty: difficulty,
		ExpiresAt:  expiresAt,
		Subject:    subject,
		Algorithm:  algorithm,
		Nonce:      nonce,
		Counter:    0, // По умолчанию 0 для challenge
	}

	// Counter есть только в solution
	if len(parts) > 7 {
		// Counter может быть в base64 формате (в solution)
		counterBytes, decodeErr := base64.StdEncoding.DecodeString(parts[7])
		if decodeErr != nil {
			return nil, fmt.Errorf("invalid counter: %w", decodeErr)
		}
		if _, parseErr := fmt.Sscanf(string(counterBytes), "%d", &h.Counter); parseErr != nil {
			return nil, fmt.Errorf("invalid counter value: %w", parseErr)
		}
	}

	return h, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(" client localhost:8080")
		os.Exit(1)
	}

	serverAddr := os.Args[1]
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		logger.Error("Failed to connect to wisdom-gate", "error", err)
		os.Exit(1)
	}
	defer func() { _ = conn.Close() }()

	if err := requestChallenge(conn); err != nil {
		logger.Error("Failed to request challenge", "error", err)
		return
	}

	challenge, err := readChallenge(conn)
	if err != nil {
		logger.Error("Failed to read challenge", "error", err)
		return
	}

	solution, err := solveChallenge(challenge, logger)
	if err != nil {
		logger.Error("Failed to solve challenge", "error", err)
		return
	}

	logger.Info("Challenge solved", "solution", solution)

	if err := sendSolution(conn, solution); err != nil {
		logger.Error("Failed to send solution", "error", err)
		return
	}

	quote, err := readQuote(conn)
	if err != nil {
		logger.Error("Failed to read quote", "error", err)
		return
	}

	fmt.Printf("\n Word of Wisdom:\n%s\n\n", quote)
}

func requestChallenge(conn net.Conn) error {
	msg := &Message{Command: REQ}
	return WriteMessage(conn, msg)
}

func readChallenge(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	msg, err := ReadMessage(reader)
	if err != nil {
		return "", err
	}

	if msg.Command != CHL {
		return "", fmt.Errorf("expected CHL, got %s", msg.Command)
	}

	return msg.Body, nil
}

func solveChallenge(challenge string, logger *slog.Logger) (string, error) {
	header, err := ParseHashcashHeader(challenge)
	if err != nil {
		return "", fmt.Errorf("failed to parse challenge: %w", err)
	}

	logger.Info("Solving challenge", "difficulty", header.Difficulty)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	solutionChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	numWorkers := 4
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			solveWorker(ctx, header, workerID, numWorkers, solutionChan, errorChan)
		}(i)
	}

	select {
	case solution := <-solutionChan:
		return solution, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("timeout solving challenge")
	}
}

func solveWorker(ctx context.Context, header *HashcashHeader, workerID, numWorkers int, solutionChan chan<- string, errorChan chan<- error) {
	startCounter := int64(workerID)

	for counter := startCounter; ; counter += int64(numWorkers) {
		select {
		case <-ctx.Done():
			return
		default:
			headerWithCounter := &HashcashHeader{
				Version:    header.Version,
				Difficulty: header.Difficulty,
				ExpiresAt:  header.ExpiresAt,
				Subject:    header.Subject,
				Algorithm:  header.Algorithm,
				Nonce:      header.Nonce,
				Counter:    counter,
			}

			if isValidSolution(headerWithCounter.String(), header.Difficulty) {
				select {
				case solutionChan <- headerWithCounter.String():
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func isValidSolution(header string, difficulty int) bool {
	hash := sha256.Sum256([]byte(header))
	hashStr := hex.EncodeToString(hash[:])

	requiredZeros := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hashStr, requiredZeros)
}

func sendSolution(conn net.Conn, solution string) error {
	msg := &Message{
		Command: RES,
		Body:    solution,
	}
	return WriteMessage(conn, msg)
}

func readQuote(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	msg, err := ReadMessage(reader)
	if err != nil {
		return "", err
	}

	if msg.Command == ERR {
		return "", fmt.Errorf("wisdom-gate error: %s", msg.Body)
	}

	if msg.Command != QOT {
		return "", fmt.Errorf("expected QOT, got %s", msg.Command)
	}

	return msg.Body, nil
}
