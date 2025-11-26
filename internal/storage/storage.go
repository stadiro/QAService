package storage

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"QAService/internal/models"
)

var (
	ErrNotFound         = gorm.ErrRecordNotFound
	ErrQuestionNotFound = errors.New("question not found")
	ErrAnswerNotFound   = errors.New("answer not found")
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// Questions

func (s *Store) ListQuestions(ctx context.Context) ([]models.Question, error) {
	var qs []models.Question
	if err := s.db.WithContext(ctx).Find(&qs).Error; err != nil {
		return nil, err
	}
	return qs, nil
}

func (s *Store) CreateQuestion(ctx context.Context, text string) (*models.Question, error) {
	q := &models.Question{Text: text}
	if err := s.db.WithContext(ctx).Create(q).Error; err != nil {
		return nil, err
	}
	return q, nil
}

func (s *Store) GetQuestionWithAnswers(ctx context.Context, id uint) (*models.Question, error) {
	var q models.Question
	if err := s.db.WithContext(ctx).Preload("Answers").First(&q, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, err
	}
	return &q, nil
}

func (s *Store) DeleteQuestion(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.Question{}, id).Error; err != nil {
		return err
	}
	return nil
}

// Answers

func (s *Store) CreateAnswer(ctx context.Context, questionID uint, userID, text string) (*models.Answer, error) {
	// Ensure question exists
	var q models.Question
	if err := s.db.WithContext(ctx).First(&q, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, err
	}

	a := &models.Answer{
		QuestionID: questionID,
		UserID:     userID,
		Text:       text,
	}
	if err := s.db.WithContext(ctx).Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) GetAnswer(ctx context.Context, id uint) (*models.Answer, error) {
	var a models.Answer
	if err := s.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAnswerNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (s *Store) DeleteAnswer(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.Answer{}, id).Error; err != nil {
		return err
	}
	return nil
}
