package experience

import "basic-life-sim/internal/model"

var InternSoftwareEngineer = model.Experience{
	Name:        "Intern Software Engineer",
	Description: "Part-time internship as a software engineer, assisting with development tasks and learning professional workflows.",
	Type:        model.PartTimer,
	Major:       "Computer Science",
	Qualifications: []model.ExperienceQualification{
		{Type: model.UniStudent, Major: []string{"Computer Science"}, Duration: 0},
	},
	ApplyExperience: func(pastExperiences []model.TakenExperience, stat *model.Stats, char *model.Character) bool {
		for _, exp := range pastExperiences {
			if exp.Type == model.UniStudent && exp.Major == "Computer Science" {
				char.Confidence.Current += 5
				return true
			}
		}
		return false
	},
}

var PrivateTeacher = model.Experience{
	Name:        "Private Teacher",
	Description: "Part-time tutoring students individually, sharing knowledge and helping them improve their academic performance.",
	Type:        model.PartTimer,
	Major:       "Education",
	Qualifications: []model.ExperienceQualification{
		{Type: model.UniStudent, Major: []string{"Education", "Computer Science"}, Duration: 0},
	},
	ApplyExperience: func(pastExperiences []model.TakenExperience, stat *model.Stats, char *model.Character) bool {
		for _, exp := range pastExperiences {
			if exp.Type == model.UniStudent && (exp.Major == "Education" || exp.Major == "Computer Science") {
				char.Confidence.Current += 5
				return true
			}
		}
		return false
	},
}
