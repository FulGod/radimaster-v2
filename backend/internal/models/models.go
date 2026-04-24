package models

import "time"

// User represents a doctor or student.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Role      string    `gorm:"size:20;default:'student'" json:"role"` // doctor, student, admin
	AvatarURL string    `gorm:"size:500" json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MedicalCase represents a patient case with imaging data.
type MedicalCase struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:500;not null" json:"title"`
	Slug        string    `gorm:"size:500;uniqueIndex" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Modality    string    `gorm:"size:50;not null;index" json:"modality"` // CT, MRI, XRay, Ultrasound
	BodyPart    string    `gorm:"size:100" json:"body_part"`
	Diagnosis   string    `gorm:"type:text" json:"diagnosis"`
	DoctorID    uint      `gorm:"not null;index" json:"doctor_id"`
	Doctor      User      `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	FolderName  string    `gorm:"size:255" json:"folder_name"`
	Phases      []Phase   `gorm:"foreignKey:CaseID;constraint:OnDelete:CASCADE" json:"phases,omitempty"`
	IsPublic    bool      `gorm:"default:true" json:"is_public"`
	ViewCount   int       `gorm:"default:0" json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Phase represents an imaging phase (PLAIN, ARTERIAL, VENOUS, etc.).
type Phase struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	CaseID     uint    `gorm:"not null;index" json:"case_id"`
	Name       string  `gorm:"size:100;not null" json:"name"`
	FolderName string  `gorm:"size:255" json:"folder_name"`
	Position   int     `gorm:"not null;default:0" json:"position"`
	SliceCount int     `gorm:"-" json:"slice_count"`
	Slices     []Slice `gorm:"foreignKey:PhaseID;constraint:OnDelete:CASCADE" json:"slices,omitempty"`
}

// Slice represents a single image slice within a phase.
type Slice struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	PhaseID  uint   `gorm:"not null;index" json:"phase_id"`
	Position int    `gorm:"not null" json:"position"`
	FileName string `gorm:"size:255" json:"file_name"`
	Width    int    `gorm:"default:512" json:"width"`
	Height   int    `gorm:"default:512" json:"height"`
}

// Classroom represents a live teaching session.
type Classroom struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	CaseID    uint        `gorm:"not null" json:"case_id"`
	Case      MedicalCase `gorm:"foreignKey:CaseID" json:"case,omitempty"`
	DoctorID  uint        `gorm:"not null" json:"doctor_id"`
	Doctor    User        `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Title     string      `gorm:"size:200;not null" json:"title"`
	Code      string      `gorm:"size:10;uniqueIndex" json:"code"`
	IsActive  bool        `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
}
