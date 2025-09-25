package services

import (
	"time"

	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

type GroupService interface {
	CreateGroup(teacherID uuid.UUID, name, subject string, grade, level int) (*models.Group, error)
	ListGroups(teacherID uuid.UUID) ([]*models.Group, error)
	AddMember(groupID, userID uuid.UUID) error
	RemoveMember(groupID, userID uuid.UUID) error
	AssignHomeworkToGroup(groupID uuid.UUID, title, description, subject string, grade, level int, due *time.Time) error
}

type groupService struct {
	groups  repository.GroupRepository
	users   repository.UserRepository
	assigns *repository.AssignmentRepository
	bot     *telegram.Bot
}

func NewGroupService(groups repository.GroupRepository, users repository.UserRepository, assigns *repository.AssignmentRepository, bot *telegram.Bot) GroupService {
	return &groupService{groups: groups, users: users, assigns: assigns, bot: bot}
}

func (s *groupService) CreateGroup(teacherID uuid.UUID, name, subject string, grade, level int) (*models.Group, error) {
	g := &models.Group{ID: uuid.New(), Name: name, Subject: subject, Grade: grade, Level: level, TeacherID: teacherID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.groups.Create(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *groupService) ListGroups(teacherID uuid.UUID) ([]*models.Group, error) {
	return s.groups.ListByTeacher(teacherID)
}

func (s *groupService) AddMember(groupID, userID uuid.UUID) error {
	m := &models.GroupMember{GroupID: groupID, UserID: userID}
	return s.groups.AddMember(m)
}

func (s *groupService) RemoveMember(groupID, userID uuid.UUID) error {
	return s.groups.RemoveMember(groupID, userID)
}

func (s *groupService) AssignHomeworkToGroup(groupID uuid.UUID, title, description, subject string, grade, level int, due *time.Time) error {
	// Получаем участников группы
	members, err := s.groups.ListMembers(groupID)
	if err != nil {
		return err
	}
	// Создаем задания каждому участнику
	for _, m := range members {
		a := &models.Assignment{
			ID:          uuid.New(),
			Title:       title,
			Description: description,
			Subject:     subject,
			Grade:       grade,
			Level:       level,
			TeacherID:   uuid.Nil, // установим ниже
		}
		// Вычисляем TeacherID из группы
		g, err := s.groups.GetByID(groupID)
		if err != nil {
			return err
		}
		a.TeacherID = g.TeacherID
		// Привязываем к ученику
		a.CreatedAt = time.Now()
		a.UpdatedAt = time.Now()
		// В модели Assignment у нас есть StudentID — установим
		a.StudentID = m.UserID
		if due != nil {
			a.DueDate = *due
		}
		if err := s.assigns.Create(a); err != nil {
			return err
		}
		// Уведомление ученику
		if s.bot != nil {
			// Найдём ученика
			u, _ := s.users.GetByID(m.UserID)
			if u != nil && u.TelegramID != 0 {
				s.bot.SendAssignmentNotification(u.TelegramID, a.Title, a.Subject, a.DueDate.Format("2006-01-02 15:04"))
			}
		}
	}
	return nil
}
