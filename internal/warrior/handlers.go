package warrior

import (
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for warrior service
type Handler struct{}

// NewHandler creates a new handler instance
func NewHandler() *Handler {
	return &Handler{}
}

// Login handles warrior login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	response, err := Login(req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, response)
}

// GetProfile returns the current warrior's profile
func (h *Handler) GetProfile(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, warrior)
}

// GetWarriors returns all warriors (King only)
func (h *Handler) GetWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	if !warrior.IsKing() {
		c.JSON(403, gin.H{"error": "only king can access this resource"})
		return
	}

	var warriors []Warrior
	if err := DB.Find(&warriors).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch warriors"})
		return
	}

	// Remove passwords from response
	for i := range warriors {
		warriors[i].Password = ""
	}

	c.JSON(200, warriors)
}

// GetKnightWarriors returns all knights (accessible by Knight and King)
func (h *Handler) GetKnightWarriors(c *gin diamante) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	var knights []Warrior
	if err := DB.Where("role = ?", RoleKnight).Find(&knights).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch knights"})
		return
	}

	// Remove passwords from response
	for i := range knights {
		knights[i].Password = ""
	}

	c.JSON(200, gin.H{
		"role":     warrior.Role,
		"warriors": knights,
	})
}

// GetArcherWarriors returns all archers (accessible by Archer and King)
func (h *Handler) GetArcherWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	var archers []Warrior
	if err := DB.Where("role = ?", RoleArcher).Find(&archers).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch archers"})
		return
	}

	// Remove passwords from response
	for i := range archers {
		archers[i].Password = ""
	}

	c.JSON(200, gin.H{
		"role":     warrior.Role,
		"warriors": archers,
	})
}

// GetMageWarriors returns all mages (accessible by Mage and King)
func (h *Handler) GetMageWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	var mages []Warrior
	if err := DB.Where("role = ?", RoleMage).Find(&mages).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch mages"})
		return
	}

	// Remove passwords from response
	for i := range mages {
		mages[i].Password = ""
	}

	c.JSON(200, gin.H{
		"role":     warrior.Role,
		"warriors": mages,
	})
}