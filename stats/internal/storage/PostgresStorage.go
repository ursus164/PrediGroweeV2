// Package storage provides database storage implementations for the stats service.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"stats/internal/models"

	"github.com/lib/pq"
)

// Storage defines the interface for quiz statistics storage operations.
type Storage interface {
	Ping() error
	Close() error
	SaveResponse(sessionID int, response *models.QuestionResponse) error
	SaveSession(session *models.QuizSession) error
	GetUserStatsForMode(userID int, mode models.QuizMode) (correctCount int, wrongCount int, err error)
	GetQuizSessionByID(quizSessionID int) (*models.QuizSession, error)
	GetQuizQuestionsStats(quizSessionID int) ([]models.QuestionStat, error)
	GetUserQuizStats(quizSessionID int) (*models.QuizStats, error)
	FinishQuizSession(quizSessionID int) error

	// survey
	SaveSurveyResponse(response *models.SurveyResponse) error
	GetSurveyResponseForUser(userID int) (*models.SurveyResponse, error)
	GetAllSurveyResponses() ([]models.SurveyResponse, error)

	// stats
	GetAllResponses() ([]models.QuestionResponse, error)
	GetStatsForQuestion(id int) (models.QuestionAllStats, error)
	GetStatsForAllQuestions() ([]models.QuestionAllStats, error)
	GetActivityStats() ([]models.ActivityStats, error)
	CountQuizSessions() (int, error)
	CountAnswers() (int, error)
	CountCorrectAnswers() (int, error)
	GetUserQuizSessionsStats(userID int) ([]*models.QuizStats, error)
	GetStatsGroupedBySurveyField(field string) ([]models.SurveyGroupedStats, error)
	DeleteUserResponses(userID int) error
	DeleteResponse(id int) error
	GetAllUsersStats() ([]models.UserQuizStats, error)
    GetLeaderboard(minAnswers, limit int) ([]models.LeaderboardRow, error)
	GetAccuracyBatch(sessionIDs []int) ([]models.SessionAccuracy, error)
}

// ErrSessionNotFound is returned when a quiz session is not found.
var ErrSessionNotFound = fmt.Errorf("session not found")

// ErrStatsNotFound is returned when statistics are not found for a user/mode.
var ErrStatsNotFound = fmt.Errorf("stats not found")

// PostgresStorage is a PostgreSQL implementation of the Storage interface.
type PostgresStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresStorage creates a new PostgresStorage instance.
func NewPostgresStorage(db *sql.DB, logger *zap.Logger) *PostgresStorage {
	return &PostgresStorage{
		db:     db,
		logger: logger,
	}
}

// Ping checks the database connection.
func (p *PostgresStorage) Ping() error {
	return p.db.PingContext(context.Background())
}

// Close closes the database connection.
func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

// SaveResponse saves a quiz question response to the database.
func (p *PostgresStorage) SaveResponse(sessionID int, response *models.QuestionResponse) error {
	_, err := p.db.ExecContext(context.Background(), `INSERT INTO answers (session_id, question_id, answer, correct, screen_size, time_spent, case_code) values ($1, $2, $3, $4, $5, $6, $7)`, sessionID, response.QuestionID, response.Answer, response.IsCorrect, response.ScreenSize, response.TimeSpent, response.CaseCode)
	if err != nil {
		return err
	}
	return nil
}

// SaveSession saves a new quiz session to the database.
func (p *PostgresStorage) SaveSession(session *models.QuizSession) error {
	_, err := p.db.ExecContext(context.Background(), `INSERT INTO quiz_sessions (user_id, quiz_mode, session_id) values ($1, $2, $3)`, session.UserID, session.QuizMode, session.SessionID)
	if err != nil {
		return err
	}
	return nil
}

// GetUserStatsForMode retrieves statistics for a specific user and quiz mode.
func (p *PostgresStorage) GetUserStatsForMode(userID int, mode models.QuizMode) (correctCount int, wrongCount int, err error) {
	rows, err := p.db.QueryContext(context.Background(), `select correct, count(*) from answers a
    join quiz_sessions s on a.session_id = s.session_id
    where user_id=$1 and quiz_mode=$2
    group by quiz_mode, correct`, userID, mode)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, ErrStatsNotFound
		}
		return 0, 0, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()
	for rows.Next() {
		var isCorrect bool
		var count int
		err = rows.Scan(&isCorrect, &count)
		if isCorrect {
			correctCount = count
		} else {
			wrongCount = count
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}
	return
}

// GetUserQuizSessionsStats retrieves statistics for all quiz sessions of a user.
func (p *PostgresStorage) GetUserQuizSessionsStats(userID int) ([]*models.QuizStats, error) {
	query := `
        SELECT s.session_id, s.quiz_mode,
               COUNT(*) as total_questions,
               SUM(CASE WHEN correct THEN 1 ELSE 0 END) as correct_answers,
			   MIN(a.answer_time) as start_time
        FROM quiz_sessions s
        JOIN answers a ON s.session_id = a.session_id
        WHERE s.user_id = $1
        GROUP BY s.session_id, s.quiz_mode
        ORDER BY s.session_id DESC`

	rows, err := p.db.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	var stats []*models.QuizStats
	for rows.Next() {
		stat := &models.QuizStats{}
		err = rows.Scan(&stat.SessionID, &stat.Mode, &stat.TotalQuestions, &stat.CorrectAnswers, &stat.StartTime)
		if err != nil {
			return nil, err
		}
		if stat.TotalQuestions != 0 {
			stat.Accuracy = float64(stat.CorrectAnswers) / float64(stat.TotalQuestions)
		}

		stat.Questions, err = p.GetQuizQuestionsStats(stat.SessionID)
		if err != nil {
			return nil, err
		}

		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetQuizSessionByID retrieves a quiz session by its ID.
func (p *PostgresStorage) GetQuizSessionByID(quizSessionID int) (*models.QuizSession, error) {
	var session models.QuizSession
	err := p.db.QueryRowContext(context.Background(), `SELECT user_id, session_id, finish_time, quiz_mode FROM quiz_sessions WHERE session_id = $1`, quizSessionID).Scan(&session.UserID, &session.SessionID, &session.FinishTime, &session.QuizMode)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

// GetQuizQuestionsStats retrieves statistics for all questions in a quiz session.
func (p *PostgresStorage) GetQuizQuestionsStats(quizSessionID int) ([]models.QuestionStat, error) {
	query := `select a.question_id, a.answer, a.correct from answers a
where a.session_id = $1`
	rows, err := p.db.QueryContext(context.Background(), query, quizSessionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()
	var questionsStats []models.QuestionStat
	for rows.Next() {
		var qs models.QuestionStat
		err = rows.Scan(&qs.QuestionID, &qs.Answer, &qs.IsCorrect)
		if err != nil {
			return nil, err
		}
		questionsStats = append(questionsStats, qs)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return questionsStats, nil
}

// GetUserQuizStats retrieves detailed statistics for a specific quiz session.
func (p *PostgresStorage) GetUserQuizStats(quizSessionID int) (*models.QuizStats, error) {
	var quizStats models.QuizStats
	err := p.db.QueryRowContext(context.Background(), `select quiz_mode, count(*) as total_questions, sum(CASE WHEN correct THEN 1 ELSE 0 END) as correct_answers from answers a
join quiz_sessions s on a.session_id = s.session_id
where a.session_id = $1
group by quiz_mode`, quizSessionID).Scan(&quizStats.Mode, &quizStats.TotalQuestions, &quizStats.CorrectAnswers)
	if err != nil {
		return nil, err
	}
	if quizStats.TotalQuestions != 0 {
		quizStats.Accuracy = float64(quizStats.CorrectAnswers) / float64(quizStats.TotalQuestions)
	} else {
		quizStats.Accuracy = 0
	}
	quizStats.Questions, err = p.GetQuizQuestionsStats(quizSessionID)
	if err != nil {
		return nil, err
	}
	return &quizStats, nil
}

// FinishQuizSession marks a quiz session as finished.
func (p *PostgresStorage) FinishQuizSession(quizSessionID int) error {
	_, err := p.db.ExecContext(context.Background(), `UPDATE quiz_sessions SET finish_time = now() WHERE session_id = $1`, quizSessionID)
	if err != nil {
		return err
	}
	return nil
}

// SaveSurveyResponse saves a user survey response to the database.
func (p *PostgresStorage) SaveSurveyResponse(response *models.SurveyResponse) error {
	query := `INSERT INTO users_surveys
    			(user_id, gender, age, vision_defect, education, experience, country, name, surname)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := p.db.ExecContext(context.Background(), query, response.UserID, response.Gender, response.Age, response.VisionDefect, response.Education, response.Experience, response.Country, response.Name, response.Surname)
	if err != nil {
		return err
	}
	return nil
}

// GetSurveyResponseForUser retrieves survey response for a specific user.
func (p *PostgresStorage) GetSurveyResponseForUser(userID int) (*models.SurveyResponse, error) {
	query := `SELECT user_id, gender, age, vision_defect, education, experience, country, name, surname
              FROM users_surveys
              WHERE user_id = $1`
	var surveyResponse models.SurveyResponse
	err := p.db.QueryRowContext(context.Background(), query, userID).Scan(
		&surveyResponse.UserID,
		&surveyResponse.Gender,
		&surveyResponse.Age,
		&surveyResponse.VisionDefect,
		&surveyResponse.Education,
		&surveyResponse.Experience,
		&surveyResponse.Country,
		&surveyResponse.Name,
		&surveyResponse.Surname,
	)
	if err != nil {
		return nil, err
	}
	return &surveyResponse, nil
}

// GetAllSurveyResponses retrieves all survey responses from the database.
func (p *PostgresStorage) GetAllSurveyResponses() ([]models.SurveyResponse, error) {
	query := `SELECT user_id, gender, age, vision_defect, education, experience, country, name, surname
              FROM users_surveys
              ORDER BY user_id`

	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	var surveys []models.SurveyResponse
	for rows.Next() {
		var survey models.SurveyResponse
		err := rows.Scan(
			&survey.UserID,
			&survey.Gender,
			&survey.Age,
			&survey.VisionDefect,
			&survey.Education,
			&survey.Experience,
			&survey.Country,
			&survey.Name,
			&survey.Surname,
		)
		if err != nil {
			return nil, err
		}
		surveys = append(surveys, survey)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return surveys, nil
}

// GetAllResponses retrieves all quiz question responses from the database.
func (p *PostgresStorage) GetAllResponses() ([]models.QuestionResponse, error) {
	query := `SELECT id, user_id, question_id, answer, correct, answer_time, answers.screen_size, answers.time_spent, answers.case_code FROM answers
    			join quiz_sessions on answers.session_id = quiz_sessions.session_id
                order by answer_time desc;`
	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()
	var stats []models.QuestionResponse
	for rows.Next() {
		var stat models.QuestionResponse
		err = rows.Scan(&stat.ID, &stat.UserID, &stat.QuestionID, &stat.Answer, &stat.IsCorrect, &stat.Time, &stat.ScreenSize, &stat.TimeSpent, &stat.CaseCode)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetStatsForQuestion retrieves statistics for a specific question.
func (p *PostgresStorage) GetStatsForQuestion(id int) (models.QuestionAllStats, error) {
	query := `SELECT question_id, count(*), sum(CASE WHEN correct THEN 1 ELSE 0 END) FROM answers
				WHERE question_id = $1
				group by question_id`
	var stats models.QuestionAllStats
	err := p.db.QueryRowContext(context.Background(), query, id).Scan(&stats.QuestionID, &stats.Total, &stats.Correct)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.QuestionAllStats{}, nil
		}
		return models.QuestionAllStats{}, err
	}
	return stats, nil
}

// GetStatsForAllQuestions retrieves statistics for all questions.
func (p *PostgresStorage) GetStatsForAllQuestions() ([]models.QuestionAllStats, error) {
	query := `SELECT question_id, count(*), sum(CASE WHEN correct THEN 1 ELSE 0 END) FROM answers
				group by question_id`
	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.QuestionAllStats{}, nil
		}
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()
	var stats []models.QuestionAllStats
	for rows.Next() {
		var stat models.QuestionAllStats
		err = rows.Scan(&stat.QuestionID, &stat.Total, &stat.Correct)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetActivityStats retrieves daily activity statistics.
func (p *PostgresStorage) GetActivityStats() ([]models.ActivityStats, error) {
	query := `SELECT * FROM (SELECT date_trunc('day', answer_time) as date, count(*), sum(CASE WHEN correct THEN 1 ELSE 0 END) FROM answers
				group by date_trunc('day', answer_time)
				order by date_trunc('day', answer_time) desc
				limit 10) as dcs ORDER BY date ASC`
	var stats []models.ActivityStats
	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		return []models.ActivityStats{}, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()
	for rows.Next() {
		var stat models.ActivityStats
		err = rows.Scan(&stat.Date, &stat.Total, &stat.Correct)
		if err != nil {
			return []models.ActivityStats{}, err
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return []models.ActivityStats{}, err
	}
	return stats, nil
}

// CountQuizSessions returns the total number of quiz sessions.
func (p *PostgresStorage) CountQuizSessions() (int, error) {
	var count int
	err := p.db.QueryRowContext(context.Background(), `SELECT count(*) FROM quiz_sessions`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountAnswers returns the total number of answers submitted.
func (p *PostgresStorage) CountAnswers() (int, error) {
	var count int
	err := p.db.QueryRowContext(context.Background(), `SELECT count(*) FROM answers`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountCorrectAnswers returns the total number of correct answers.
func (p *PostgresStorage) CountCorrectAnswers() (int, error) {
	var count int
	err := p.db.QueryRowContext(context.Background(), `SELECT count(*) FROM answers WHERE correct = true`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetStatsGroupedBySurveyField retrieves statistics grouped by a survey field.
func (p *PostgresStorage) GetStatsGroupedBySurveyField(field string) ([]models.SurveyGroupedStats, error) {
	var query string
	switch field {
	case "gender":
		query = `
           SELECT
               us.gender as group_field,
               COUNT(a.session_id) as total,
               SUM(CASE WHEN a.correct THEN 1 ELSE 0 END) as correct
           FROM users_surveys us
           LEFT JOIN quiz_sessions qs ON us.user_id = qs.user_id
           LEFT JOIN answers a ON qs.session_id = a.session_id
           WHERE us.gender IS NOT NULL
           GROUP BY us.gender`
	case "age":
		query = `
           SELECT
               us.age as group_field,
               COUNT(a.session_id) as total,
               SUM(CASE WHEN a.correct THEN 1 ELSE 0 END) as correct
           FROM users_surveys us
           LEFT JOIN quiz_sessions qs ON us.user_id = qs.user_id
           LEFT JOIN answers a ON qs.session_id = a.session_id
           WHERE us.age IS NOT NULL
           GROUP BY us.age`
	case "vision_defect":
		query = `
           SELECT
               us.vision_defect as group_field,
               COUNT(a.session_id) as total,
               SUM(CASE WHEN a.correct THEN 1 ELSE 0 END) as correct
           FROM users_surveys us
           LEFT JOIN quiz_sessions qs ON us.user_id = qs.user_id
           LEFT JOIN answers a ON qs.session_id = a.session_id
           WHERE us.vision_defect IS NOT NULL
           GROUP BY us.vision_defect`
	case "education":
		query = `
           SELECT
               us.education as group_field,
               COUNT(a.session_id) as total,
               SUM(CASE WHEN a.correct THEN 1 ELSE 0 END) as correct
           FROM users_surveys us
           LEFT JOIN quiz_sessions qs ON us.user_id = qs.user_id
           LEFT JOIN answers a ON qs.session_id = a.session_id
           WHERE us.education IS NOT NULL
           GROUP BY us.education`
	case "experience":
		query = `
           SELECT
               us.experience as group_field,
               COUNT(a.session_id) as total,
               SUM(CASE WHEN a.correct THEN 1 ELSE 0 END) as correct
           FROM users_surveys us
           LEFT JOIN quiz_sessions qs ON us.user_id = qs.user_id
           LEFT JOIN answers a ON qs.session_id = a.session_id
           WHERE us.experience IS NOT NULL
           GROUP BY us.experience`
	case "country":
		query = `
           SELECT
               us.country as group_field,
               COUNT(a.session_id) as total,
               SUM(CASE WHEN a.correct THEN 1 ELSE 0 END) as correct
           FROM users_surveys us
           LEFT JOIN quiz_sessions qs ON us.user_id = qs.user_id
           LEFT JOIN answers a ON qs.session_id = a.session_id
           WHERE us.country IS NOT NULL
           GROUP BY us.country`
	default:
		return nil, fmt.Errorf("unsupported field: %s", field)
	}

	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	var stats []models.SurveyGroupedStats
	for rows.Next() {
		var s models.SurveyGroupedStats
		err := rows.Scan(&s.Value, &s.Total, &s.Correct)
		if err != nil {
			return nil, err
		}
		s.Group = field
		if s.Total > 0 {
			s.Accuracy = float64(s.Correct) / float64(s.Total)
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// DeleteUserResponses deletes all quiz responses for a specific user.
func (p *PostgresStorage) DeleteUserResponses(userID int) error {
	_, err := p.db.ExecContext(context.Background(), `DELETE FROM answers WHERE session_id IN (SELECT session_id FROM quiz_sessions WHERE user_id = $1)`, userID)
	if err != nil {
		return err
	}
	return nil
}

// DeleteResponse deletes a specific quiz response.
func (p *PostgresStorage) DeleteResponse(id int) error {
	_, err := p.db.ExecContext(context.Background(), `DELETE FROM answers where id = $1`, id)
	if err != nil {
		return err
	}
	return nil
}

// GetAllUsersStats retrieves aggregated statistics for all users.
func (p *PostgresStorage) GetAllUsersStats() ([]models.UserQuizStats, error) {
	query := `
        SELECT s.user_id,
               COUNT(*) as total_answers,
               SUM(CASE WHEN correct THEN 1 ELSE 0 END) as correct_answers,
               us.experience,
               us.education
        FROM answers a
        JOIN quiz_sessions s ON a.session_id = s.session_id
        LEFT JOIN users_surveys us ON s.user_id = us.user_id
        GROUP BY s.user_id, us.experience, us.education`

	rows, err := p.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	var stats []models.UserQuizStats
	for rows.Next() {
		var stat models.UserQuizStats
		var experience, education sql.NullString
		err := rows.Scan(&stat.UserID, &stat.TotalAnswers, &stat.CorrectAnswers, &experience, &education)
		if err != nil {
			return nil, err
		}
		stat.Experience = experience.String
		stat.Education = education.String
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetLeaderboard retrieves the leaderboard with minimum answer threshold and limit.
func (p *PostgresStorage) GetLeaderboard(minAnswers, limit int) ([]models.LeaderboardRow, error) {
	const q = `
WITH agg AS (
  SELECT
    qs.user_id,
    COALESCE(SUM(CASE WHEN a.correct THEN 1 ELSE 0 END),0) AS correct_answers,
    COUNT(a.id) AS total_answers
  FROM quiz_sessions qs
  JOIN answers a ON a.session_id = qs.session_id
  GROUP BY qs.user_id
)
SELECT
  agg.user_id,
  us.education,
  us.country,
  agg.total_answers,
  agg.correct_answers,
  CASE WHEN agg.total_answers = 0 THEN 0
       ELSE agg.correct_answers::decimal / agg.total_answers::decimal
  END AS accuracy
FROM agg
LEFT JOIN users_surveys us ON us.user_id = agg.user_id
WHERE agg.total_answers >= $1
ORDER BY accuracy DESC, total_answers DESC
LIMIT $2;
`
	rows, err := p.db.QueryContext(context.Background(), q, minAnswers, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	res := make([]models.LeaderboardRow, 0, limit)
	for rows.Next() {
		var r models.LeaderboardRow
		if err := rows.Scan(&r.UserID, &r.Education, &r.Country, &r.TotalAnswers, &r.CorrectAnswers, &r.Accuracy); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, rows.Err()
}

// GetAccuracyBatch retrieves accuracy statistics for multiple quiz sessions.
func (p *PostgresStorage) GetAccuracyBatch(sessionIDs []int) ([]models.SessionAccuracy, error) {
	out := make([]models.SessionAccuracy, 0, len(sessionIDs))
	if len(sessionIDs) == 0 {
		return out, nil
	}
	rows, err := p.db.QueryContext(context.Background(), `
		SELECT a.session_id,
		       COUNT(*) AS total,
		       SUM(CASE WHEN a.correct THEN 1 ELSE 0 END) AS correct
		  FROM answers a
		 WHERE a.session_id = ANY($1)
		 GROUP BY a.session_id
	`, pq.Array(sessionIDs))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("failed to close rows", zap.Error(err))
		}
	}()

	m := make(map[int]models.SessionAccuracy, len(sessionIDs))
	for rows.Next() {
		var sid, total, correct int
		if err := rows.Scan(&sid, &total, &correct); err != nil {
			return nil, err
		}
		acc := 0.0
		if total > 0 {
			acc = float64(correct) / float64(total)
		}
		m[sid] = models.SessionAccuracy{
			SessionID: sid, Correct: correct, Total: total, Accuracy: acc,
		}
	}
	for _, id := range sessionIDs {
		if v, ok := m[id]; ok {
			out = append(out, v)
		} else {
			out = append(out, models.SessionAccuracy{
				SessionID: id, Correct: 0, Total: 0, Accuracy: 0,
			})
		}
	}
	return out, rows.Err()
}
