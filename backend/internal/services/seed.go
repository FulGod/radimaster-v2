package services

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"radimaster/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type caseMeta struct {
	Title     string
	Modality  string
	BodyPart  string
	Diagnosis string
	Desc      string
}

var caseData = []caseMeta{
	{"CT Abdomen — Hepatocellular Carcinoma", "CT", "Abdomen", "Ung thư biểu mô tế bào gan (HCC) với tăng quang điển hình thì động mạch", "Bệnh nhân nam 58 tuổi, tiền sử viêm gan B mạn. CT bụng có tiêm thuốc cản quang cho thấy khối u gan phải."},
	{"CT Brain — Ischemic Stroke", "CT", "Brain", "Nhồi máu não cấp vùng cấp máu động mạch não giữa (MCA)", "Bệnh nhân nữ 72 tuổi, đột ngột yếu liệt nửa người phải. CT não không tiêm thuốc cho thấy vùng giảm tỷ trọng."},
	{"MRI Knee — Anterior Cruciate Ligament Tear", "MRI", "Knee", "Đứt dây chằng chéo trước (ACL) hoàn toàn", "Bệnh nhân nam 25 tuổi, chấn thương thể thao. MRI gối cho thấy mất liên tục dây chằng chéo trước."},
	{"CT Chest — Pulmonary Embolism", "CT", "Chest", "Thuyên tắc phổi cấp hai bên", "Bệnh nhân nữ 45 tuổi, khó thở đột ngột sau phẫu thuật. CT ngực có thuốc cản quang cho thấy khuyết thuốc trong động mạch phổi."},
	{"MRI Brain — Meningioma", "MRI", "Brain", "U màng não vùng cạnh đường giữa (Falx meningioma)", "Bệnh nhân nữ 55 tuổi, nhức đầu kéo dài. MRI sọ não cho thấy khối u ngoài trục bắt thuốc đồng nhất."},
	{"CT Abdomen — Acute Appendicitis", "CT", "Abdomen", "Viêm ruột thừa cấp có biến chứng", "Bệnh nhân nam 30 tuổi, đau bụng hố chậu phải. CT bụng cho thấy ruột thừa sưng to với thâm nhiễm mỡ xung quanh."},
	{"X-Ray Chest — Pneumothorax", "XRay", "Chest", "Tràn khí màng phổi tự phát bên phải", "Bệnh nhân nam 22 tuổi, đau ngực phải đột ngột. X-quang ngực thẳng cho thấy đường tràn khí màng phổi."},
	{"CT Spine — Herniated Disc", "CT", "Spine", "Thoát vị đĩa đệm L4-L5 chèn ép rễ thần kinh", "Bệnh nhân nam 45 tuổi, đau lưng lan xuống chân trái. CT cột sống thắt lưng cho thấy thoát vị đĩa đệm."},
	{"MRI Shoulder — Rotator Cuff Tear", "MRI", "Shoulder", "Rách gân cơ trên gai (Supraspinatus) bán phần", "Bệnh nhân nam 50 tuổi, đau vai phải khi nâng tay. MRI vai cho thấy tín hiệu bất thường tại gân cơ trên gai."},
	{"CT Abdomen — Kidney Stone", "CT", "Abdomen", "Sỏi thận phải 12mm gây ứ nước thận", "Bệnh nhân nữ 40 tuổi, đau hông lưng phải dữ dội. CT bụng không thuốc cho thấy sỏi cản quang tại bể thận phải."},
	{"MRI Brain — Multiple Sclerosis", "MRI", "Brain", "Xơ cứng rải rác (MS) — nhiều ổ tổn thương chất trắng", "Bệnh nhân nữ 32 tuổi, tê bì tay chân từng đợt. MRI não cho thấy nhiều ổ tăng tín hiệu T2/FLAIR quanh não thất."},
	{"CT Chest — Lung Cancer", "CT", "Chest", "Ung thư phổi không tế bào nhỏ, khối u thùy trên phải", "Bệnh nhân nam 60 tuổi, ho kéo dài, gầy sút cân. CT ngực có thuốc cho thấy khối u đặc ở thùy trên phổi phải."},
	{"X-Ray Hand — Fracture", "XRay", "Hand", "Gãy xương bàn tay số 5 (Boxer's fracture)", "Bệnh nhân nam 28 tuổi, chấn thương tay phải. X-quang bàn tay cho thấy gãy cổ xương bàn tay thứ 5 với di lệch."},
	{"CT Brain — Subdural Hematoma", "CT", "Brain", "Máu tụ dưới màng cứng cấp tính bán cầu trái", "Bệnh nhân nam 70 tuổi, té ngã, lơ mơ. CT não không thuốc cho thấy khối máu tụ hình lưỡi liềm dưới màng cứng."},
	{"MRI Liver — Hemangioma", "MRI", "Liver", "U mạch máu gan (Hemangioma) — lành tính", "Bệnh nhân nữ 35 tuổi, phát hiện tình cờ. MRI gan cho thấy khối u tăng tín hiệu T2 với bắt thuốc ly tâm điển hình."},
	{"CT Abdomen — Pancreatitis", "CT", "Abdomen", "Viêm tụy cấp — phù nề lan tỏa tuyến tụy", "Bệnh nhân nam 48 tuổi, đau bụng thượng vị dữ dội. CT bụng cho thấy tuyến tụy sưng phù với thâm nhiễm mỡ quanh tụy."},
	{"X-Ray Pelvis — Hip Fracture", "XRay", "Pelvis", "Gãy cổ xương đùi bên trái", "Bệnh nhân nữ 75 tuổi, té ngã. X-quang khung chậu cho thấy gãy cổ xương đùi trái với di lệch."},
	{"MRI Bone — Rheumatoid Arthritis", "MRI", "Bone", "Viêm khớp dạng thấp — tổn thương xương khớp bàn tay", "Bệnh nhân nữ 42 tuổi, đau sưng khớp bàn tay. MRI cho thấy viêm màng hoạt dịch và bào mòn xương."},
	{"CT Chest — Aortic Dissection", "CT", "Chest", "Bóc tách động mạch chủ type B (Stanford)", "Bệnh nhân nam 65 tuổi, đau ngực xé sau lưng. CT ngực có thuốc cho thấy vạt nội mạc bóc tách động mạch chủ xuống."},
}

// extractNumber extracts the leading number from a filename for natural sorting.
// e.g. "IM000002.jpg" -> 2, "10.jpg" -> 10, "thumbnails.jpg" -> 999999
var numRegex = regexp.MustCompile(`(\d+)`)

func extractNumber(name string) int {
	matches := numRegex.FindStringSubmatch(name)
	if len(matches) > 1 {
		n, err := strconv.Atoi(matches[1])
		if err == nil {
			return n
		}
	}
	return 999999 // non-numeric files go last
}

// sortFilesNaturally sorts DirEntry slice by numeric value in filename.
func sortFilesNaturally(files []os.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		return extractNumber(files[i].Name()) < extractNumber(files[j].Name())
	})
}

// Seed populates the database from V1 medical data directory if empty.
func Seed(db *gorm.DB, mediaDir string) {
	var count int64
	db.Model(&models.MedicalCase{}).Count(&count)
	if count > 0 {
		log.Println("📦 Database already seeded, skipping")
		return
	}

	log.Println("🌱 Seeding database from medical data...")

	// Create default users
	doctorPw, _ := bcrypt.GenerateFromPassword([]byte("doctor123"), bcrypt.DefaultCost)
	doctor := models.User{Name: "BS. Nguyễn Văn An", Email: "doctor@radimaster.com", Password: string(doctorPw), Role: "doctor"}
	db.Create(&doctor)

	studentPw, _ := bcrypt.GenerateFromPassword([]byte("student123"), bcrypt.DefaultCost)
	student := models.User{Name: "SV. Trần Minh Khoa", Email: "student@radimaster.com", Password: string(studentPw), Role: "student"}
	db.Create(&student)
	log.Printf("👤 Created users: doctor@radimaster.com / doctor123, student@radimaster.com / student123")

	// Scan media directory
	entries, err := os.ReadDir(mediaDir)
	if err != nil {
		log.Printf("⚠️ Media dir not found: %s — skipping seed", mediaDir)
		return
	}

	caseIdx := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if caseIdx >= len(caseData) {
			break
		}

		meta := caseData[caseIdx]
		slug := slugify(meta.Title)

		medCase := models.MedicalCase{
			Title:       meta.Title,
			Slug:        slug,
			Modality:    meta.Modality,
			BodyPart:    meta.BodyPart,
			Diagnosis:   meta.Diagnosis,
			Description: meta.Desc,
			DoctorID:    doctor.ID,
			FolderName:  entry.Name(),
			IsPublic:    true,
		}
		db.Create(&medCase)

		// Scan phases (sorted by folder name e.g. 1.PLAIN, 2.ARTERIAL)
		phaseEntries, _ := os.ReadDir(filepath.Join(mediaDir, entry.Name()))
		phaseCount := 0
		for j, phaseEntry := range phaseEntries {
			if !phaseEntry.IsDir() {
				continue
			}
			phaseName := phaseEntry.Name()
			displayName := phaseName
			if parts := strings.SplitN(phaseName, ".", 2); len(parts) == 2 {
				displayName = parts[1]
			}

			phase := models.Phase{
				CaseID:     medCase.ID,
				Name:       displayName,
				FolderName: phaseName,
				Position:   j,
			}
			db.Create(&phase)

			// Scan slices — NATURAL NUMERIC SORT
			sliceFiles, _ := os.ReadDir(filepath.Join(mediaDir, entry.Name(), phaseName))

			// Filter to .jpg only, exclude thumbnails
			var jpgFiles []os.DirEntry
			for _, sf := range sliceFiles {
				if sf.IsDir() {
					continue
				}
				name := strings.ToLower(sf.Name())
				if !strings.HasSuffix(name, ".jpg") {
					continue
				}
				if strings.Contains(name, "thumbnail") {
					continue
				}
				jpgFiles = append(jpgFiles, sf)
			}

			// Sort by number: 1.jpg, 2.jpg, 3.jpg... or IM000001.jpg, IM000002.jpg...
			sortFilesNaturally(jpgFiles)

			for k, sf := range jpgFiles {
				slice := models.Slice{PhaseID: phase.ID, Position: k + 1, FileName: sf.Name(), Width: 512, Height: 512}
				db.Create(&slice)
			}
			phaseCount++
		}
		log.Printf("  ✅ Case %d: %s (%d phases)", caseIdx+1, meta.Title, phaseCount)
		caseIdx++
	}

	log.Printf("🌱 Seeding complete: %d cases imported", caseIdx)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "—", "-")
	var result []byte
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, byte(c))
		}
	}
	return string(result)
}
