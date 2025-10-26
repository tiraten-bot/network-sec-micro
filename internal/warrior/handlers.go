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

// GetWarriors returns all warriors (admin only)
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

// Example endpoints for different resources

// GetWeapons - accessible by Knight and Archer
func (h *Handler) GetWeapons(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "weapons retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}

// GetArmor - accessible by Knight
func (h *Handler) GetArmor(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "armor retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}

// GetBattles - accessible by Knight
func (h *Handler) GetBattles(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "battles retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}

// GetArrows - accessible by Archer
func (h *Handler) GetArrows(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "arrows retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}

// GetScouting - accessible by Archer
func (h *Handler) GetScouting(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "scouting information retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}

// GetSpells - accessible by Mage
func (h *Handler) GetSpells(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "spells retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}

// GetPotions - accessible by Mage
func (h *Handler) GetPotions(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "potions retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}

// GetLibrary - accessible by Mage
func (h *Handler) GetLibrary(c *gin.Context) {
	warrior, _ := GetCurrentWarrior(c)
	c.JSON(200, gin.H{
		"message": "library information retrieved",
		"warrior": warrior.Username,
		"role":    warrior.Role,
	})
}
