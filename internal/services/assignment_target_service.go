package services

import (
	"time"

	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

type AssignmentService interface {
	// Assignment CRUD
	CreateAssignment(assignment *models.Assignment) error
	GetAssignment(id uuid.UUID) (*models.Assignment, error)
	UpdateAssignment(assignment *models.Assignment) error
	DeleteAssignment(id uuid.UUID) error

	// Group assignments
	CreateGroupAssignment(teacherID, groupID uuid.UUID, title, description, subject string, grade, level int, dueDate time.Time) (*models.Assignment, error)
	ListAssignmentsByTeacher(teacherID uuid.UUID) ([]*models.Assignment, error)
	ListAssignmentsByGroup(groupID uuid.UUID) ([]*models.Assignment, error)

	// AssignmentTarget operations
	GetAssignmentTargetsByStudent(studentID uuid.UUID) ([]*models.AssignmentTarget, error)
	GetAssignmentTargetsByAssignment(assignmentID uuid.UUID) ([]*models.AssignmentTarget, error)
	GetAssignmentTarget(id uuid.UUID) (*models.AssignmentTarget, error)

	// Status management
	MarkAsOverdue() error
	GetOverdueAssignments() ([]*models.AssignmentTarget, error)
}

type assignmentService struct {
	assignmentRepo       repository.AssignmentRepository
	assignmentTargetRepo repository.AssignmentTargetRepository
	groupRepo            repository.GroupRepository
	userRepo             repository.UserRepository
	notificationRepo     repository.NotificationRepository
	bot                  *telegram.Bot
}

func NewAssignmentService(
	assignmentRepo repository.AssignmentRepository,
	assignmentTargetRepo repository.AssignmentTargetRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
	bot *telegram.Bot,
) AssignmentService {
	return &assignmentService{
		assignmentRepo:       assignmentRepo,
		assignmentTargetRepo: assignmentTargetRepo,
		groupRepo:            groupRepo,
		userRepo:             userRepo,
		notificationRepo:     notificationRepo,
		bot:                  bot,
	}
}

func (s *assignmentService) CreateAssignment(assignment *models.Assignment) error {
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	assignment.CreatedAt = time.Now()
	assignment.UpdatedAt = time.Now()

	return s.assignmentRepo.Create(assignment)
}

func (s *assignmentService) GetAssignment(id uuid.UUID) (*models.Assignment, error) {
	return s.assignmentRepo.GetByID(id)
}

func (s *assignmentService) UpdateAssignment(assignment *models.Assignment) error {
	assignment.UpdatedAt = time.Now()
	return s.assignmentRepo.Update(assignment)
}

func (s *assignmentService) DeleteAssignment(id uuid.UUID) error {
	return s.assignmentRepo.Delete(id)
}

func (s *assignmentService) CreateGroupAssignment(teacherID, groupID uuid.UUID, title, description, subject string, grade, level int, dueDate time.Time) (*models.Assignment, error) {
	// Создаем основное задание
	assignment := &models.Assignment{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Subject:     subject,
		Grade:       grade,
		Level:       level,
		TeacherID:   teacherID,
		GroupID:     &groupID,
		DueDate:     dueDate,
		Status:      "active",
		CreatedBy:   teacherID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.CreateAssignment(assignment); err != nil {
		return nil, err
	}

	// Получаем участников группы
	members, err := s.groupRepo.ListMembers(groupID)
	if err != nil {
		return nil, err
	}

	// Создаем AssignmentTarget для каждого участника
	var studentIDs []uuid.UUID
	for _, member := range members {
		studentIDs = append(studentIDs, member.UserID)
	}

	if err := s.assignmentTargetRepo.CreateForGroup(assignment.ID, studentIDs); err != nil {
		return nil, err
	}

	// Отправляем уведомления участникам группы
	if s.bot != nil {
		for _, member := range members {
			student, err := s.userRepo.GetByID(member.UserID)
			if err == nil && student.TelegramID != 0 {
				s.bot.SendAssignmentNotification(
					student.TelegramID,
					assignment.Title,
					assignment.Subject,
					assignment.DueDate.Format("2006-01-02 15:04"),
				)
			}
		}
	}

	// Создаем уведомления в системе
	if len(studentIDs) > 0 {
		s.notificationRepo.CreateForGroup(
			studentIDs,
			models.NotificationTypeNewAssignment,
			"Новое задание",
			assignment.Title,
			`{"assignment_id":"`+assignment.ID.String()+`","group_id":"`+groupID.String()+`"}`,
		)
	}

	return assignment, nil
}

func (s *assignmentService) ListAssignmentsByTeacher(teacherID uuid.UUID) ([]*models.Assignment, error) {
	assignments, err := s.assignmentRepo.GetByTeacherID(teacherID)
	if err != nil {
		return nil, err
	}

	// Convert to pointers
	result := make([]*models.Assignment, len(assignments))
	for i := range assignments {
		result[i] = &assignments[i]
	}

	return result, nil
}

func (s *assignmentService) ListAssignmentsByGroup(groupID uuid.UUID) ([]*models.Assignment, error) {
	return s.assignmentRepo.GetByGroupID(groupID)
}

func (s *assignmentService) GetAssignmentTargetsByStudent(studentID uuid.UUID) ([]*models.AssignmentTarget, error) {
	return s.assignmentTargetRepo.ListByStudent(studentID)
}

func (s *assignmentService) GetAssignmentTargetsByAssignment(assignmentID uuid.UUID) ([]*models.AssignmentTarget, error) {
	return s.assignmentTargetRepo.ListByAssignment(assignmentID)
}

func (s *assignmentService) GetAssignmentTarget(id uuid.UUID) (*models.AssignmentTarget, error) {
	return s.assignmentTargetRepo.GetByID(id)
}

func (s *assignmentService) MarkAsOverdue() error {
	return s.assignmentTargetRepo.MarkAsOverdue()
}

func (s *assignmentService) GetOverdueAssignments() ([]*models.AssignmentTarget, error) {
	return s.assignmentTargetRepo.ListByStatus(models.AssignmentTargetStatusOverdue)
}
