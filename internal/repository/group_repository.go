package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

type GroupRepository interface {
	Create(group *models.Group) error
	Update(group *models.Group) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*models.Group, error)
	ListByTeacher(teacherID uuid.UUID) ([]*models.Group, error)

	AddMember(member *models.GroupMember) error
	RemoveMember(groupID, userID uuid.UUID) error
	ListMembers(groupID uuid.UUID) ([]*models.GroupMember, error)
	IsMember(groupID, userID uuid.UUID) (bool, error)
}

type groupRepository struct{ db *gorm.DB }

func NewGroupRepository(db *gorm.DB) GroupRepository { return &groupRepository{db: db} }

func (r *groupRepository) Create(group *models.Group) error {
	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}
	return r.db.Create(group).Error
}

func (r *groupRepository) Update(group *models.Group) error { return r.db.Save(group).Error }

func (r *groupRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Group{}, "id = ?", id).Error
}

func (r *groupRepository) GetByID(id uuid.UUID) (*models.Group, error) {
	var g models.Group
	err := r.db.Preload("Members").Preload("Teacher").First(&g, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupRepository) ListByTeacher(teacherID uuid.UUID) ([]*models.Group, error) {
	var gs []*models.Group
	err := r.db.Where("teacher_id = ?", teacherID).Order("created_at DESC").Find(&gs).Error
	return gs, err
}

func (r *groupRepository) AddMember(member *models.GroupMember) error {
	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}
	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}
	return r.db.Create(member).Error
}

func (r *groupRepository) RemoveMember(groupID, userID uuid.UUID) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

func (r *groupRepository) ListMembers(groupID uuid.UUID) ([]*models.GroupMember, error) {
	var ms []*models.GroupMember
	err := r.db.Preload("User").Where("group_id = ?", groupID).Find(&ms).Error
	return ms, err
}

func (r *groupRepository) IsMember(groupID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error
	return count > 0, err
}
