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
	GetGroup(id uuid.UUID) (*models.Group, error)
	UpdateGroup(group *models.Group) error
	DeleteGroup(id uuid.UUID) error
	ListGroups(teacherID uuid.UUID) ([]*models.Group, error)
	ListGroupsForStudent(studentID uuid.UUID) ([]*models.Group, error)

	AddMember(groupID, userID uuid.UUID) error
	RemoveMember(groupID, userID uuid.UUID) error
	ListMembers(groupID uuid.UUID) ([]*models.GroupMember, error)
	IsMember(groupID, userID uuid.UUID) (bool, error)

	AssignHomeworkToGroup(groupID uuid.UUID, title, description, subject string, grade, level int, due *time.Time) error
}

type groupService struct {
	groups  repository.GroupRepository
	users   repository.UserRepository
	assigns repository.AssignmentRepository
	bot     *telegram.Bot
}

func NewGroupService(groups repository.GroupRepository, users repository.UserRepository, assigns repository.AssignmentRepository, bot *telegram.Bot) GroupService {
	return &groupService{groups: groups, users: users, assigns: assigns, bot: bot}
}

func (s *groupService) CreateGroup(teacherID uuid.UUID, name, subject string, grade, level int) (*models.Group, error) {
	g := &models.Group{
		ID:        uuid.New(),
		Name:      name,
		Subject:   subject,
		Grade:     grade,
		Level:     level,
		TeacherID: teacherID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.groups.Create(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *groupService) GetGroup(id uuid.UUID) (*models.Group, error) {
	return s.groups.GetByID(id)
}

func (s *groupService) UpdateGroup(group *models.Group) error {
	group.UpdatedAt = time.Now()
	return s.groups.Update(group)
}

func (s *groupService) DeleteGroup(id uuid.UUID) error {
	return s.groups.Delete(id)
}

func (s *groupService) ListGroups(teacherID uuid.UUID) ([]*models.Group, error) {
	return s.groups.ListByTeacher(teacherID)
}

func (s *groupService) ListGroupsForStudent(studentID uuid.UUID) ([]*models.Group, error) {
	// Получаем все группы, где студент является участником
	var groups []*models.Group
	members, err := s.groups.ListMembers(uuid.Nil) // Получаем всех участников всех групп
	if err != nil {
		return nil, err
	}

	// Фильтруем группы для конкретного студента
	groupIDs := make(map[uuid.UUID]bool)
	for _, member := range members {
		if member.UserID == studentID {
			groupIDs[member.GroupID] = true
		}
	}

	// Получаем группы
	for groupID := range groupIDs {
		group, err := s.groups.GetByID(groupID)
		if err != nil {
			continue // Пропускаем группы с ошибками
		}
		groups = append(groups, group)
	}

	return groups, nil
}

func (s *groupService) AddMember(groupID, userID uuid.UUID) error {
	m := &models.GroupMember{GroupID: groupID, UserID: userID}
	return s.groups.AddMember(m)
}

func (s *groupService) RemoveMember(groupID, userID uuid.UUID) error {
	return s.groups.RemoveMember(groupID, userID)
}

func (s *groupService) ListMembers(groupID uuid.UUID) ([]*models.GroupMember, error) {
	return s.groups.ListMembers(groupID)
}

func (s *groupService) IsMember(groupID, userID uuid.UUID) (bool, error) {
	members, err := s.groups.ListMembers(groupID)
	if err != nil {
		return false, err
	}

	for _, member := range members {
		if member.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *groupService) AssignHomeworkToGroup(groupID uuid.UUID, title, description, subject string, grade, level int, due *time.Time) error {
	// Получаем группу для проверки TeacherID
	group, err := s.groups.GetByID(groupID)
	if err != nil {
		return err
	}

	// Используем новый AssignmentService для создания группового задания
	assignmentService := NewAssignmentService(
		s.assigns,
		nil, // assignmentTargetRepo - нужно будет добавить
		s.groups,
		s.users,
		nil, // notificationRepo - нужно будет добавить
		s.bot,
	)

	dueDate := time.Now().Add(7 * 24 * time.Hour) // По умолчанию через неделю
	if due != nil {
		dueDate = *due
	}

	_, err = assignmentService.CreateGroupAssignment(
		group.TeacherID,
		groupID,
		title,
		description,
		subject,
		grade,
		level,
		dueDate,
	)

	return err
}
