package storage

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"quiz/internal/models"
	"strconv"
	"time"
)

type Store interface {
	Ping() error
	Close() error

	// sessions
	CreateQuizSession(session models.QuizSession) (models.QuizSession, error)
	GetQuizSessionByID(id int) (models.QuizSession, error)
	UpdateQuizSession(session models.QuizSession) error
	GetUserActiveQuizSessions(userID int) ([]models.QuizSession, error)
	GetUserLastQuizSession(userID int) (*models.QuizSession, error)
	GetTimeLimit() (int, error)
	SaveSettings(name string, value string) error

	// questions
	GetQuestionByID(id int) (models.Question, error)
	GetAllQuestions() ([]models.Question, error)
	CreateQuestion(newCase models.QuestionPayload) (models.QuestionPayload, error)
	UpdateQuestionByID(questionID int, updatedCase models.QuestionPayload) (models.QuestionPayload, error)
	UpdateQuestionCorrectOption(questionID int, option string) error
	DeleteQuestionByID(id int) error
	CountQuestions() (int, error)
	GetQuestionOptions(id int) ([]string, error)
	GetQuestionCorrectOption(id int) (string, error)

	// options
	GetAllOptions() ([]models.Option, error)
	CreateOption(option models.Option) (models.Option, error)
	UpdateOption(id int, option models.Option) error

	// cases
	CreateCase(newCase models.Case) (models.Case, error)
	UpdateCase(updatedCase models.Case) (models.Case, error)
	DeleteCaseWithParameters(id int) error
	GetAllCases() ([]models.Case, error)
	GetCaseByID(id int) (models.Case, error)
	CreateCaseParameter(caseID int, parameter models.ParameterValue) (models.ParameterValue, error)
	UpdateCaseParameters(caseID int, parameters []models.Parameter, values []models.ParameterValue) error

	// parameters
	CreateParameter(parameter models.Parameter) (models.Parameter, error)
	UpdateParameter(parameter models.Parameter) error
	DeleteParameter(id int) error
	GetAllParameters() ([]models.Parameter, error)
	GetParameterByID(id int) (models.Parameter, error)
	UpdateParametersOrder(params []models.Parameter) error
	GetCaseParametersV3(caseID int) ([]models.ParameterValue, error)
	GetCaseAge3(caseID int) (int, error)

	// groups
	GetGroupQuestionsIDsRandomOrder(groupID int) ([]int, error)
	GetNextQuestionGroupID(currentGroup int) (int, error)
	DeleteOption(id int) error
	GetSettings() ([]models.Settings, error)

	// access/security
	GetSecuritySettings() (mode string, cooldownHours int, err error)
	IsUserApproved(userID int) (bool, error)
	UpsertAndGetRegisteredAt(userID int) (time.Time, error)
	ApproveUser(userID int, adminID *int) error
	UnapproveUser(userID int, adminID *int) error
	GetApprovedUserIDs() ([]int, error)

	// bug reports
	CreateCaseReport(caseID int, userID int, description string) error
	ListCaseReports() ([]models.CaseReport, error)
	DeleteCaseReport(id int) error
	UpdateCaseReportNote(id int, note *string, adminID *int) error
    CountReportsWithoutNote() (int, error)

	// tests
	CreateTest(t models.Test, questionIDs []int) (models.Test, error)
	GetTestByCode(code string) (*models.Test, error)
	GetTestQuestionIDsOrdered(testID int) ([]int, error)
	ListTestsByOwner(userID int) ([]models.Test, error)
	ListSessionsByTestID(testID int) ([]models.QuizSession, error)
	GetTestByID(id int) (*models.Test, error)
	DeleteTest(id int) error

    // difficulty
    InsertDifficultyVote(questionID int, userID int, level models.DifficultyLevel) error
    GetMyDifficultyVote(questionID int, userID int) (*models.QuestionDifficultyVote, error)
    GetDifficultySummary(questionID int) (*models.QuestionDifficultySummary, error)
	GetDifficultySummaryBatch(ids []int) ([]models.QuestionDifficultySummary, error)


	// live sessions
	ListActiveSessions(cutoffMinutes int, limit int) ([]models.ActiveSession, error)

	// favorites
	AddFavoriteCase(userID int, caseID int) error
	RemoveFavoriteCase(userID int, caseID int) error
	ListFavoriteCases(userID int) ([]models.FavoriteCase, error)
	UpdateFavoriteNote(userID int, caseID int, note *string) error

}

type PostgresStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresStorage(db *sql.DB, logger *zap.Logger) *PostgresStorage {
	return &PostgresStorage{
		db:     db,
		logger: logger,
	}
}

func (s *PostgresStorage) Ping() error {
	return s.db.Ping()
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

//
// Quiz Sessions
//

func (s *PostgresStorage) CreateQuizSession(session models.QuizSession) (models.QuizSession, error) {
	query := `
        INSERT INTO quiz_sessions (
            user_id, status, mode, screen_size,
            current_question, current_group, group_order,
            test_id, test_code,
            created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
        RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(
		query,
		session.UserID,
		session.Status,
		session.Mode,
		session.ScreenSize,
		session.CurrentQuestionID,
		session.CurrentGroup,
		pq.Array(session.GroupOrder),
		session.TestID,
		session.TestCode,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)

	return session, err
}

func (s *PostgresStorage) GetQuizSessionByID(id int) (models.QuizSession, error) {
	var session models.QuizSession
	query := `
        SELECT id, user_id, status, mode, current_question, current_group, group_order,
               created_at, updated_at, finished_at, question_requested_time,
               test_id, test_code
        FROM quiz_sessions
        WHERE id = $1`

	var groupArr []sql.NullInt64
	var testIDNull sql.NullInt64
	var testCodeNull sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.Status,
		&session.Mode,
		&session.CurrentQuestionID,
		&session.CurrentGroup,
		pq.Array(&groupArr),
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.FinishedAt,
		&session.QuestionRequestedTime,
		&testIDNull,
		&testCodeNull,
	)
	session.GroupOrder = make([]int, 0, len(groupArr))
	for _, v := range groupArr {
		if v.Valid {
			session.GroupOrder = append(session.GroupOrder, int(v.Int64))
		}
	}
	if testIDNull.Valid {
		tid := int(testIDNull.Int64)
		session.TestID = &tid
	}
	if testCodeNull.Valid {
		code := testCodeNull.String
		session.TestCode = &code
	}

	return session, err
}

func (s *PostgresStorage) UpdateQuizSession(session models.QuizSession) error {
	query := `
        UPDATE quiz_sessions
           SET status = $1,
               mode = $2,
               current_question = $3,
               current_group = $4,
               group_order = $5,
               finished_at = $6,
               question_requested_time = $7,
               updated_at = NOW()
         WHERE id = $8`

	_, err := s.db.Exec(
		query,
		session.Status,
		session.Mode,
		session.CurrentQuestionID,
		session.CurrentGroup,
		pq.Array(session.GroupOrder),
		session.FinishedAt,
		session.QuestionRequestedTime,
		session.ID,
	)
	return err
}

// todo: handle group & group order
func (s *PostgresStorage) GetUserActiveQuizSessions(userID int) ([]models.QuizSession, error) {
	query := `
        SELECT id, user_id, status, mode, current_question, created_at, updated_at, finished_at
        FROM quiz_sessions
        WHERE user_id = $1 and status != 'finished'
        ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.QuizSession
	for rows.Next() {
		var session models.QuizSession
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Status,
			&session.Mode,
			&session.CurrentQuestionID,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (s *PostgresStorage) GetUserLastQuizSession(userID int) (*models.QuizSession, error) {
	query := `
        SELECT id, user_id, status, mode, current_question, current_group, group_order,
               created_at, updated_at, finished_at, test_id, test_code
        FROM quiz_sessions
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT 1`

	var session models.QuizSession
	var groupArr []int64
	var testIDNull sql.NullInt64
	var testCodeNull sql.NullString

	err := s.db.QueryRow(query, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.Status,
		&session.Mode,
		&session.CurrentQuestionID,
		&session.CurrentGroup,
		pq.Array(&groupArr),
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.FinishedAt,
		&testIDNull,
		&testCodeNull,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	session.GroupOrder = make([]int, len(groupArr))
	for i, v := range groupArr {
		session.GroupOrder[i] = int(v)
	}
	if testIDNull.Valid {
		tid := int(testIDNull.Int64)
		session.TestID = &tid
	}
	if testCodeNull.Valid {
		code := testCodeNull.String
		session.TestCode = &code
	}

	return &session, nil
}

//
// Questions
//

func (s *PostgresStorage) GetQuestionByID(id int) (models.Question, error) {
	query := `
        SELECT q.id, q.question, q.prediction_age,
               c.id, c.code, c.patient_gender, c.age1, c.age2, c.age3, q.group_number
        FROM questions q
        JOIN cases c ON q.case_id = c.id
        WHERE q.id = $1`

	var question models.Question
	err := s.db.QueryRow(query, id).Scan(
		&question.ID,
		&question.Question,
		&question.PredictionAge,
		&question.Case.ID,
		&question.Case.Code,
		&question.Case.Gender,
		&question.Case.Age1,
		&question.Case.Age2,
		&question.Case.Age3,
		&question.Group,
	)
	if err != nil {
		return question, err
	}

	question.Options, err = s.GetQuestionOptions(id)
	if err != nil {
		return question, err
	}

	question.Case.Parameters, question.Case.ParameterValues, err = s.getCaseParameters(question.Case.ID)
	if err != nil {
		return question, err
	}

	return question, nil
}

func (s *PostgresStorage) GetQuestionOptions(id int) ([]string, error) {
	query := `
		SELECT o.option from options o
			JOIN question_options qo on o.id = qo.option_id
			WHERE qo.question_id = $1 ORDER BY o.id`

	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	var options []string
	for rows.Next() {
		var option string
		err := rows.Scan(&option)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	return options, rows.Err()
}

func (s *PostgresStorage) GetQuestionCorrectOption(id int) (string, error) {
	query := `
		SELECT o.option from options o
			JOIN question_options qo on o.id = qo.option_id
			WHERE qo.question_id = $1 and qo.is_correct = true`

	var option string
	err := s.db.QueryRow(query, id).Scan(&option)
	return option, err
}

func (s *PostgresStorage) GetAllQuestions() ([]models.Question, error) {
	query := `
		SELECT q.id, q.question, q.prediction_age,
		       c.id, c.code, c.patient_gender, c.age1, c.age2, c.age3, group_number
		  FROM questions q
		  JOIN cases c ON q.case_id = c.id
		 ORDER BY q.id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	var questions []models.Question
	for rows.Next() {
		var question models.Question
		err = rows.Scan(
			&question.ID,
			&question.Question,
			&question.PredictionAge,
			&question.Case.ID,
			&question.Case.Code,
			&question.Case.Gender,
			&question.Case.Age1,
			&question.Case.Age2,
			&question.Case.Age3,
			&question.Group,
		)
		if err != nil {
			return nil, err
		}
		question.Options, err = s.GetQuestionOptions(question.ID)
		if err != nil {
			s.logger.Error("Failed to get question options", zap.Error(err))
		}
		correct, err := s.GetQuestionCorrectOption(question.ID)
		if err != nil {
			s.logger.Error("Failed to get question correct option", zap.Error(err))
		}
		question.Correct = &correct
		questions = append(questions, question)
	}
	defer rows.Close()
	return questions, nil
}

//
// Options
//

func (s *PostgresStorage) GetAllOptions() ([]models.Option, error) {
	query := `
		SELECT o.id, o.option, count(qo.id) from options o
		      left join public.question_options qo on o.id = qo.option_id
				group by o.id, o.option ORDER BY o.id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	var options []models.Option
	for rows.Next() {
		var option models.Option
		err := rows.Scan(&option.ID, &option.Option, &option.Questions)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	return options, nil
}

//
// Groups
//

func (s *PostgresStorage) GetGroupQuestionsIDsRandomOrder(groupNumber int) ([]int, error) {
	query := `
		SELECT id from questions
		WHERE group_number = $1
		order by random()`

	rows, err := s.db.Query(query, groupNumber)
	if err != nil {
		return nil, err
	}
	var questions []int
	for rows.Next() {
		var questionID int
		err = rows.Scan(&questionID)
		if err != nil {
			return nil, err
		}
		questions = append(questions, questionID)
	}
	defer rows.Close()
	return questions, nil
}

func (s *PostgresStorage) GetNextQuestionGroupID(currentGroup int) (int, error) {
	if currentGroup == 0 {
		var g int
		err := s.db.QueryRow(`
			SELECT group_number
			  FROM questions
			 GROUP BY group_number
			 ORDER BY random()
			 LIMIT 1
		`).Scan(&g)
		return g, err
	}

	var g int
	err := s.db.QueryRow(`
		SELECT group_number
		  FROM questions
		 WHERE group_number <> $1
		 GROUP BY group_number
		 ORDER BY random()
		 LIMIT 1
	`, currentGroup).Scan(&g)
	return g, err
}

//
// Cases
//

func (s *PostgresStorage) CreateQuestion(payload models.QuestionPayload) (models.QuestionPayload, error) {
	query := `
        INSERT INTO questions (question, prediction_age, case_id)
        VALUES ($1, $2, $3)
        RETURNING id`

	err := s.db.QueryRow(
		query,
		payload.Question,
		payload.PredictionAge,
		payload.CaseID,
	).Scan(&payload.ID)

	return payload, err
}

func (s *PostgresStorage) UpdateQuestionByID(questionID int, payload models.QuestionPayload) (models.QuestionPayload, error) {
	query := `
        UPDATE questions
           SET question = $1, prediction_age = $2, case_id = $3, group_number = $5
         WHERE id = $4`

	_, err := s.db.Exec(
		query,
		payload.Question,
		payload.PredictionAge,
		payload.CaseID,
		questionID,
		payload.Group,
	)

	payload.ID = questionID
	return payload, err
}

func (s *PostgresStorage) UpdateQuestionCorrectOption(questionID int, option string) error {
	var newCorrectID int
	err := s.db.QueryRow("SELECT id FROM options WHERE option = $1", option).Scan(&newCorrectID)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Reset all options to false for this question
	if _, err := tx.Exec(`
        UPDATE question_options 
           SET is_correct = false 
         WHERE question_id = $1`, questionID); err != nil {
		return fmt.Errorf("reset options: %w", err)
	}

	// Set new correct option
	if _, err := tx.Exec(`
        UPDATE question_options 
           SET is_correct = true 
         WHERE question_id = $1 AND option_id = $2`,
		questionID, newCorrectID); err != nil {
		return fmt.Errorf("update correct option: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresStorage) DeleteQuestionByID(id int) error {
	query := "DELETE FROM questions WHERE id = $1"
	_, err := s.db.Exec(query, id)
	return err
}

func (s *PostgresStorage) CountQuestions() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM questions").Scan(&count)
	return count, err
}

func (s *PostgresStorage) CreateCase(newCase models.Case) (models.Case, error) {
	query := `
        INSERT INTO cases (code, patient_gender, age1, age2)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	err := s.db.QueryRow(
		query,
		newCase.Code,
		newCase.Gender,
		newCase.Age1,
		newCase.Age2,
	).Scan(&newCase.ID)

	return newCase, err
}

func (s *PostgresStorage) UpdateCase(updatedCase models.Case) (models.Case, error) {
	query := `
        UPDATE cases
           SET code = $1, patient_gender = $2, age1 = $3, age2 = $4
         WHERE id = $5`

	_, err := s.db.Exec(
		query,
		updatedCase.Code,
		updatedCase.Gender,
		updatedCase.Age1,
		updatedCase.Age2,
		updatedCase.ID,
	)

	return updatedCase, err
}

func (s *PostgresStorage) DeleteCaseWithParameters(id int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM case_parameters WHERE case_id = $1", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM cases WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *PostgresStorage) GetAllCases() ([]models.Case, error) {
	query := `
        SELECT id, code, patient_gender, age1, age2
          FROM cases
         ORDER BY id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []models.Case
	for rows.Next() {
		var c models.Case
		err = rows.Scan(
			&c.ID,
			&c.Code,
			&c.Gender,
			&c.Age1,
			&c.Age2,
		)
		if err != nil {
			return nil, err
		}

		c.Parameters, c.ParameterValues, err = s.getCaseParameters(c.ID)
		if err != nil {
			return nil, err
		}

		cases = append(cases, c)
	}

	return cases, rows.Err()
}

func (s *PostgresStorage) GetCaseByID(id int) (models.Case, error) {
	query := `
        SELECT id, code, patient_gender, age1, age2
          FROM cases
         WHERE id=$1`

	var c models.Case
	err := s.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.Code,
		&c.Gender,
		&c.Age1,
		&c.Age2)
	if err != nil {
		return c, err
	}
	c.Parameters, c.ParameterValues, err = s.getCaseParameters(c.ID)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (s *PostgresStorage) CreateCaseParameter(caseID int, parameter models.ParameterValue) (models.ParameterValue, error) {
	query := `
		INSERT INTO case_parameters (case_id, parameter_id, value_1, value_2)
		VALUES ($1, $2, $3, $4)
		RETURNING parameter_id, value_1, value_2`

	err := s.db.QueryRow(
		query,
		caseID,
		parameter.ParameterID,
		parameter.Value1,
		parameter.Value2,
	).Scan(&parameter.ParameterID, &parameter.Value1, &parameter.Value2)

	return parameter, err
}

//
// Parameters
//

func (s *PostgresStorage) CreateParameter(parameter models.Parameter) (models.Parameter, error) {
	query := `
        INSERT INTO parameters (name, description, reference_value)
        VALUES ($1, $2, $3)
        RETURNING id`

	err := s.db.QueryRow(
		query,
		parameter.Name,
		parameter.Description,
		parameter.ReferenceValues,
	).Scan(&parameter.ID)

	return parameter, err
}

func (s *PostgresStorage) UpdateParameter(parameter models.Parameter) error {
	query := `
        UPDATE parameters
           SET name = $1, description = $2, reference_value = $3
         WHERE id = $4`

	_, err := s.db.Exec(
		query,
		parameter.Name,
		parameter.Description,
		parameter.ReferenceValues,
		parameter.ID,
	)

	return err
}

func (s *PostgresStorage) DeleteParameter(id int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM case_parameters WHERE parameter_id = $1`, id); err != nil {
		return err
	}

	res, err := tx.Exec(`DELETE FROM parameters WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetParameterByID(id int) (models.Parameter, error) {
	query := `
		SELECT id, name, description, reference_value
		  FROM parameters
		 WHERE id = $1`

	var p models.Parameter
	err := s.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.ReferenceValues,
	)
	return p, err
}

func (s *PostgresStorage) GetAllParameters() ([]models.Parameter, error) {
	query := `
        SELECT id, name, description, reference_value, display_order
          FROM parameters
         ORDER BY display_order, id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parameters []models.Parameter
	for rows.Next() {
		var p models.Parameter
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.ReferenceValues,
			&p.Order,
		)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, p)
	}

	return parameters, rows.Err()
}

//
// Helpers
//

func (s *PostgresStorage) getCaseParameters(caseID int) ([]models.Parameter, []models.ParameterValue, error) {
	query := `select cp.parameter_id, cp.value_1, cp.value_2, cp.value_3, p.description, p.name, p.reference_value from cases c
		join case_parameters cp on c.id = cp.case_id
		join parameters p on cp.parameter_id = p.id
		where c.id=$1 ORDER BY p.display_order, p.id`
	rows, err := s.db.Query(query, caseID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	parameters := make([]models.Parameter, 0)
	parameterValues := make([]models.ParameterValue, 0)
	for rows.Next() {
		var p models.Parameter
		var pv models.ParameterValue
		err := rows.Scan(&p.ID, &pv.Value1, &pv.Value2, &pv.Value3, &p.Description, &p.Name, &p.ReferenceValues)
		if err != nil {
			return nil, nil, err
		}
		pv.ParameterID = p.ID
		parameters = append(parameters, p)
		parameterValues = append(parameterValues, pv)
	}
	return parameters, parameterValues, nil
}

func (s *PostgresStorage) UpdateCaseParameters(caseID int, parameters []models.Parameter, values []models.ParameterValue) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing parameters for this case
	_, err = tx.Exec("DELETE FROM case_parameters WHERE case_id = $1", caseID)
	if err != nil {
		return err
	}

	// Insert new parameters
	stmt, err := tx.Prepare(`
        INSERT INTO case_parameters (case_id, parameter_id, value_1, value_2, value_3) 
        VALUES ($1, $2, $3, $4, $5)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range parameters {
		_, err = stmt.Exec(
			caseID,
			parameters[i].ID,
			values[i].Value1,
			values[i].Value2,
			values[i].Value3,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

//
// Options mutate
//

func (s *PostgresStorage) CreateOption(option models.Option) (models.Option, error) {
	query := `
		INSERT INTO options (option)
		VALUES ($1)
		RETURNING id`

	err := s.db.QueryRow(query, option.Option).Scan(&option.ID)
	return option, err
}

func (s *PostgresStorage) UpdateOption(id int, option models.Option) error {
	query := `
		UPDATE options
		   SET option = $1
		 WHERE id = $2`

	_, err := s.db.Exec(query, option.Option, id)
	return err
}

func (s *PostgresStorage) DeleteOption(id int) error {
	query := "DELETE FROM options WHERE id = $1"
	_, err := s.db.Exec(query, id)
	return err
}

func (s *PostgresStorage) UpdateParametersOrder(params []models.Parameter) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE parameters SET display_order = $1 WHERE id = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, param := range params {
		_, err = stmt.Exec(param.Order, param.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

//
// Settings / limits
//

func (s *PostgresStorage) GetTimeLimit() (int, error) {
	var timeLimitStr string
	err := s.db.QueryRow("SELECT value FROM settings WHERE name = 'time_limit'").Scan(&timeLimitStr)
	if err != nil {
		return 0, err
	}
	timeLimit, err := strconv.Atoi(timeLimitStr)
	return timeLimit, err
}

func (s *PostgresStorage) SaveSettings(name string, value string) error {
	query := `
		INSERT INTO settings (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = $2`

	_, err := s.db.Exec(query, name, value)
	return err
}

func (s *PostgresStorage) GetSettings() ([]models.Settings, error) {
	query := `
		SELECT name, value
		  FROM settings`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.Settings
	for rows.Next() {
		var setting models.Settings
		err = rows.Scan(&setting.Name, &setting.Value)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

func (s *PostgresStorage) GetSecuritySettings() (string, int, error) {
	var mode, hrsStr string
	if err := s.db.QueryRow("SELECT value FROM settings WHERE name='quiz_security_mode'").Scan(&mode); err != nil {
		mode = "cooldown"
	}
	if err := s.db.QueryRow("SELECT value FROM settings WHERE name='quiz_cooldown_hours'").Scan(&hrsStr); err != nil {
		hrsStr = "24"
	}
	h, err := strconv.Atoi(hrsStr)
	if err != nil {
		h = 24
	}
	return mode, h, nil
}

func (s *PostgresStorage) IsUserApproved(userID int) (bool, error) {
	var ok bool
	err := s.db.QueryRow("SELECT approved FROM quiz_user_access WHERE user_id=$1", userID).Scan(&ok)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return ok, err
}

func (s *PostgresStorage) UpsertAndGetRegisteredAt(userID int) (time.Time, error) {
	var t time.Time
	err := s.db.QueryRow("SELECT registered_at FROM quiz_user_registry WHERE user_id=$1", userID).Scan(&t)
	if err == nil {
		return t, nil
	}
	err = s.db.QueryRow(`
    INSERT INTO quiz_user_registry(user_id, registered_at, source)
    VALUES($1, NOW(), 'first_seen')
    ON CONFLICT (user_id) DO NOTHING
    RETURNING registered_at
  `, userID).Scan(&t)
	if err == sql.ErrNoRows {
		_ = s.db.QueryRow("SELECT registered_at FROM quiz_user_registry WHERE user_id=$1").Scan(&t)
		return t, nil
	}
	return t, err
}

func (s *PostgresStorage) ApproveUser(userID int, adminID *int) error {
	_, err := s.db.Exec(`
    INSERT INTO quiz_user_access(user_id, approved, approved_by, approved_at)
    VALUES ($1, TRUE, $2, NOW())
    ON CONFLICT (user_id)
    DO UPDATE SET approved=TRUE, approved_by=EXCLUDED.approved_by, approved_at=EXCLUDED.approved_at
  `, userID, adminID)
	return err
}

func (s *PostgresStorage) UnapproveUser(userID int, adminID *int) error {
  _, err := s.db.Exec(`
    INSERT INTO quiz_user_access(user_id, approved, approved_by, approved_at, created_at)
    VALUES ($1, FALSE, $2, NOW(), NOW())
    ON CONFLICT (user_id)
    DO UPDATE SET approved=FALSE, approved_by=EXCLUDED.approved_by, approved_at=EXCLUDED.approved_at
  `, userID, adminID)
  return err
}

func (s *PostgresStorage) GetApprovedUserIDs() ([]int, error) {
	rows, err := s.db.Query(`SELECT user_id FROM quiz_user_access WHERE approved = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

//
// Bug reports
//

func (s *PostgresStorage) CreateCaseReport(caseID int, userID int, description string) error {
	_, err := s.db.Exec(`
		INSERT INTO case_reports(case_id, user_id, description)
		VALUES ($1, $2, $3)
	`, caseID, userID, description)
	return err
}

func (s *PostgresStorage) ListCaseReports() ([]models.CaseReport, error) {
	rows, err := s.db.Query(`
		SELECT r.id, r.case_id, c.code, r.user_id, r.description, r.created_at,
         r.admin_note, r.admin_note_updated_at, r.admin_note_updated_by
         FROM case_reports r
        JOIN cases c ON c.id = r.case_id
        ORDER BY r.created_at DESC, r.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.CaseReport, 0)

	for rows.Next() {
		var r models.CaseReport
		if err := rows.Scan(&r.ID, &r.CaseID, &r.CaseCode, &r.UserID, &r.Description, &r.CreatedAt, &r.AdminNote, &r.AdminNoteUpdatedAt, &r.AdminNoteUpdatedBy,); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *PostgresStorage) DeleteCaseReport(id int) error {
	_, err := s.db.Exec(`DELETE FROM case_reports WHERE id = $1`, id)
	return err
}

func (s *PostgresStorage) UpdateCaseReportNote(id int, note *string, adminID *int) error {
	_, err := s.db.Exec(`
    UPDATE case_reports
       SET admin_note = $2,
           admin_note_updated_at = NOW(),
           admin_note_updated_by = $3
     WHERE id = $1
  `, id, note, adminID)
	return err
}

func (s *PostgresStorage) CountReportsWithoutNote() (int, error) {
	var n int
	err := s.db.QueryRow(`
    SELECT COUNT(*)
      FROM case_reports
     WHERE admin_note IS NULL OR btrim(admin_note) = ''
  `).Scan(&n)
	return n, err
}


//
// V3 helpers
//

func (s *PostgresStorage) GetCaseParametersV3(caseID int) ([]models.ParameterValue, error) {
	query := `
        SELECT cp.parameter_id, cp.value_3
        FROM case_parameters cp
        JOIN parameters p ON p.id = cp.parameter_id
        WHERE cp.case_id = $1
        ORDER BY p.display_order, p.id
    `
	rows, err := s.db.Query(query, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]models.ParameterValue, 0)
	for rows.Next() {
		var pv models.ParameterValue
		if err := rows.Scan(&pv.ParameterID, &pv.Value3); err != nil {
			return nil, err
		}
		vals = append(vals, pv)
	}
	return vals, rows.Err()
}

func (s *PostgresStorage) GetCaseAge3(caseID int) (int, error) {
	var age3 sql.NullInt64
	err := s.db.QueryRow(`SELECT age3 FROM cases WHERE id = $1`, caseID).Scan(&age3)
	if err != nil {
		return 0, err
	}
	if !age3.Valid {
		return 0, nil
	}
	return int(age3.Int64), nil
}

//
// Tests
//

func (s *PostgresStorage) CreateTest(t models.Test, questionIDs []int) (models.Test, error) {
	err := s.db.QueryRow(
		`INSERT INTO tests(code, name, created_by) VALUES ($1,$2,$3)
         RETURNING id, created_at`,
		t.Code, t.Name, t.CreatedBy,
	).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return t, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return t, err
	}
	stmt, err := tx.Prepare(`INSERT INTO test_questions(test_id, question_id, sort_order) VALUES ($1,$2,$3)`)
	if err != nil {
		tx.Rollback()
		return t, err
	}
	defer stmt.Close()

	for i, qid := range questionIDs {
		if _, err := stmt.Exec(t.ID, qid, i); err != nil {
			tx.Rollback()
			return t, err
		}
	}
	if err := tx.Commit(); err != nil {
		return t, err
	}
	return t, nil
}

func (s *PostgresStorage) GetTestByCode(code string) (*models.Test, error) {
	var t models.Test
	err := s.db.QueryRow(
		`SELECT id, code, name, created_by, created_at FROM tests WHERE code=$1`,
		code,
	).Scan(&t.ID, &t.Code, &t.Name, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (s *PostgresStorage) GetTestQuestionIDsOrdered(testID int) ([]int, error) {
	rows, err := s.db.Query(
		`SELECT question_id FROM test_questions WHERE test_id=$1 ORDER BY sort_order`,
		testID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]int, 0)
	for rows.Next() {
		var qid int
		if err := rows.Scan(&qid); err != nil {
			return nil, err
		}
		out = append(out, qid)
	}
	return out, rows.Err()
}

func (s *PostgresStorage) ListTestsByOwner(userID int) ([]models.Test, error) {
	rows, err := s.db.Query(
		`SELECT id, code, name, created_by, created_at
		   FROM tests
		  WHERE created_by = $1
		  ORDER BY created_at DESC, id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Test
	for rows.Next() {
		var t models.Test
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *PostgresStorage) ListSessionsByTestID(testID int) ([]models.QuizSession, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, status, mode, current_question, current_group, group_order,
		       created_at, updated_at, finished_at, question_requested_time, test_id, test_code
		  FROM quiz_sessions
		 WHERE test_id = $1
		 ORDER BY created_at DESC, id DESC
	`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.QuizSession, 0)
	for rows.Next() {
		var sss models.QuizSession
		var goArr []sql.NullInt64
		var testIDNull sql.NullInt64
		var testCodeNull sql.NullString

		if err := rows.Scan(
			&sss.ID, &sss.UserID, &sss.Status, &sss.Mode,
			&sss.CurrentQuestionID, &sss.CurrentGroup,
			pq.Array(&goArr),
			&sss.CreatedAt, &sss.UpdatedAt, &sss.FinishedAt, &sss.QuestionRequestedTime,
			&testIDNull, &testCodeNull,
		); err != nil {
			return nil, err
		}
		sss.GroupOrder = make([]int, 0, len(goArr))
		for _, v := range goArr {
			if v.Valid {
				sss.GroupOrder = append(sss.GroupOrder, int(v.Int64))
			}
		}
		if testIDNull.Valid {
			id := int(testIDNull.Int64)
			sss.TestID = &id
		}
		if testCodeNull.Valid {
			code := testCodeNull.String
			sss.TestCode = &code
		}
		out = append(out, sss)
	}
	return out, rows.Err()
}

func (s *PostgresStorage) GetTestByID(id int) (*models.Test, error) {
	var t models.Test
	err := s.db.QueryRow(
		`SELECT id, code, name, created_by, created_at FROM tests WHERE id=$1`,
		id,
	).Scan(&t.ID, &t.Code, &t.Name, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (s *PostgresStorage) DeleteTest(id int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM test_questions WHERE test_id=$1`, id); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM tests WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func (s *PostgresStorage) InsertDifficultyVote(questionID int, userID int, level models.DifficultyLevel) error {
    _, err := s.db.Exec(`
        INSERT INTO question_difficulty_votes (question_id, user_id, difficulty)
        VALUES ($1, $2, $3::difficulty_level)
    `, questionID, userID, string(level))
    return err
}

func (s *PostgresStorage) GetMyDifficultyVote(questionID int, userID int) (*models.QuestionDifficultyVote, error) {
    var v models.QuestionDifficultyVote
    err := s.db.QueryRow(`
        SELECT question_id, user_id, difficulty, created_at
          FROM question_difficulty_votes
         WHERE question_id = $1 AND user_id = $2
    `, questionID, userID).Scan(&v.QuestionID, &v.UserID, &v.Difficulty, &v.CreatedAt)
    if err == sql.ErrNoRows { return nil, nil }
    if err != nil { return nil, err }
    return &v, nil
}

func (s *PostgresStorage) GetDifficultySummary(questionID int) (*models.QuestionDifficultySummary, error) {
    var out models.QuestionDifficultySummary
    err := s.db.QueryRow(`
        SELECT question_id, total_votes, hard_votes, easy_votes, hard_pct
          FROM question_difficulty_summary
         WHERE question_id = $1
    `, questionID).Scan(&out.QuestionID, &out.TotalVotes, &out.HardVotes, &out.EasyVotes, &out.HardPct)
    if err == sql.ErrNoRows { return &models.QuestionDifficultySummary{QuestionID: questionID}, nil }
    return &out, err
}

func (s *PostgresStorage) GetDifficultySummaryBatch(ids []int) ([]models.QuestionDifficultySummary, error) {
  out := make([]models.QuestionDifficultySummary, 0)
  if len(ids) == 0 {
    return out, nil
  }
  rows, err := s.db.Query(`
    SELECT question_id, total_votes, hard_votes, easy_votes, hard_pct
      FROM question_difficulty_summary
     WHERE question_id = ANY($1)
     ORDER BY question_id
  `, pq.Array(ids))
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  for rows.Next() {
    var r models.QuestionDifficultySummary
    if err := rows.Scan(&r.QuestionID, &r.TotalVotes, &r.HardVotes, &r.EasyVotes, &r.HardPct); err != nil {
      return nil, err
    }
    out = append(out, r)
  }

  have := make(map[int]struct{}, len(out))
  for _, r := range out { have[r.QuestionID] = struct{}{} }
  for _, id := range ids {
    if _, ok := have[id]; !ok {
      out = append(out, models.QuestionDifficultySummary{
        QuestionID: id, TotalVotes: 0, HardVotes: 0, EasyVotes: 0, HardPct: 0,
      })
    }
  }
  return out, rows.Err()
}

func (s *PostgresStorage) ListActiveSessions(cutoffMinutes int, limit int) ([]models.ActiveSession, error) {
	if cutoffMinutes <= 0 { cutoffMinutes = 5 }
	if limit <= 0 || limit > 500 { limit = 200 }

	rows, err := s.db.Query(`
		SELECT id, user_id, status, mode, current_question, current_group, group_order,
		       created_at, updated_at, finished_at, question_requested_time, test_id, test_code,
		       COALESCE(question_requested_time, updated_at) AS last_seen
		  FROM quiz_sessions
		 WHERE finished_at IS NULL
		   AND COALESCE(question_requested_time, updated_at) > NOW() - ($1 || ' minutes')::interval
		 ORDER BY last_seen DESC
		 LIMIT $2
	`, cutoffMinutes, limit)
	if err != nil { return nil, err }
	defer rows.Close()

	out := make([]models.ActiveSession, 0, limit)
	for rows.Next() {
		var (
			a models.ActiveSession
			goArr []sql.NullInt64
			testIDNull sql.NullInt64
			testCodeNull sql.NullString
			qreqNull sql.NullTime
		)
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Status, &a.Mode,
			&a.CurrentQuestionID, &a.CurrentGroup, pq.Array(&goArr),
			&a.CreatedAt, &a.UpdatedAt, &a.FinishedAt, &qreqNull, &testIDNull, &testCodeNull,
			&a.LastSeen,
		); err != nil { return nil, err }
		a.GroupOrder = make([]int, 0, len(goArr))
		for _, v := range goArr { if v.Valid { a.GroupOrder = append(a.GroupOrder, int(v.Int64)) } }
		if testIDNull.Valid { v := int(testIDNull.Int64); a.TestID = &v }
		if testCodeNull.Valid { v := testCodeNull.String; a.TestCode = &v }
		out = append(out, a)
	}
	return out, rows.Err()
}

//favorites

func (s *PostgresStorage) AddFavoriteCase(userID int, caseID int) error {
	_, err := s.db.Exec(`
		INSERT INTO user_favorites(user_id, case_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, case_id) DO NOTHING
	`, userID, caseID)
	return err
}

func (s *PostgresStorage) RemoveFavoriteCase(userID int, caseID int) error {
	_, err := s.db.Exec(`
		DELETE FROM user_favorites
		 WHERE user_id = $1 AND case_id = $2
	`, userID, caseID)
	return err
}

func (s *PostgresStorage) ListFavoriteCases(userID int) ([]models.FavoriteCase, error) {
    rows, err := s.db.Query(`
        SELECT uf.created_at,
               c.id, c.code, c.patient_gender, c.age1, c.age2, c.age3,
               q.id AS question_id,
               o.option AS correct_option,
               uf.note, uf.note_updated_at
          FROM user_favorites uf
          JOIN cases c ON c.id = uf.case_id
          LEFT JOIN questions q ON q.case_id = c.id
          LEFT JOIN question_options qo ON qo.question_id = q.id AND qo.is_correct = TRUE
          LEFT JOIN options o ON o.id = qo.option_id
         WHERE uf.user_id = $1
         ORDER BY uf.created_at DESC, c.id DESC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    out := make([]models.FavoriteCase, 0)
    for rows.Next() {
        var (
            f          models.FavoriteCase
            qidNull    sql.NullInt64
            corrNull   sql.NullString
            age3Null   sql.NullInt64
            noteNull   sql.NullString
            noteAtNull sql.NullTime
        )
        if err := rows.Scan(
            &f.CreatedAt,
            &f.CaseID, &f.CaseCode, &f.Gender, &f.Age1, &f.Age2, &age3Null,
            &qidNull,
            &corrNull,
            &noteNull, &noteAtNull,
        ); err != nil {
            return nil, err
        }

        if qidNull.Valid {
            v := int(qidNull.Int64)
            f.QuestionID = &v
        }
        if corrNull.Valid {
            v := corrNull.String
            f.Correct = &v
        }
        if age3Null.Valid {
            v := int(age3Null.Int64)
            f.Age3 = &v
        }
        if noteNull.Valid {
            v := noteNull.String
            f.Note = &v
        }
        if noteAtNull.Valid {
            v := noteAtNull.Time
            f.NoteUpdatedAt = &v
        }

        params, vals, err := s.getCaseParameters(f.CaseID)
        if err != nil {
            return nil, err
        }
        f.Parameters = params
        f.ParameterValues = vals

        out = append(out, f)
    }
    return out, rows.Err()
}


func (s *PostgresStorage) UpdateFavoriteNote(userID int, caseID int, note *string) error {
	_, err := s.db.Exec(`
		UPDATE user_favorites
		   SET note = $3::text,
		       note_updated_at = CASE
		           WHEN $3::text IS NULL OR btrim($3::text) = '' THEN NULL
		           ELSE NOW()
		       END
		 WHERE user_id = $1 AND case_id = $2
	`, userID, caseID, note)
	return err
}
