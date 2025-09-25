package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/services"
)

type GroupHandler struct{ svc services.GroupService }

func NewGroupHandler(svc services.GroupService) *GroupHandler { return &GroupHandler{svc: svc} }

type createGroupReq struct {
	Name, Subject string
	Grade, Level  int
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	roleVal, _ := c.Get("user_role")
	if roleVal != models.RoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	tval, _ := c.Get("user_id")
	teacherID, ok := tval.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	var req createGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	g, err := h.svc.CreateGroup(teacherID, req.Name, req.Subject, req.Grade, req.Level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"group": g})
}

func (h *GroupHandler) ListGroups(c *gin.Context) {
	roleVal, _ := c.Get("user_role")
	if roleVal != models.RoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	tval, _ := c.Get("user_id")
	teacherID, ok := tval.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}
	gs, err := h.svc.ListGroups(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": gs})
}

type memberReq struct {
	UserID uuid.UUID `json:"user_id"`
}

func (h *GroupHandler) AddMember(c *gin.Context) {
	gid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req memberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AddMember(gid, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	gid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	uid, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.svc.RemoveMember(gid, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type assignGroupReq struct {
	Title, Description, Subject string
	Grade, Level                int
	DueAt                       *time.Time
}

func (h *GroupHandler) AssignHomework(c *gin.Context) {
	gid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req assignGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AssignHomeworkToGroup(gid, req.Title, req.Description, req.Subject, req.Grade, req.Level, req.DueAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "assigned"})
}
