package experience

import "basic-life-sim/internal/model"

var ComputerScienceUniversity = model.Experience{
	Name:           "Computer Science",
	Description:    "Bachelor's degree in Computer Science covering algorithms, data structures, software engineering, and systems programming.",
	Type:           model.UniStudent,
	Major:          "Computer Science",
	Qualifications: []model.ExperienceQualification{},
	ApplyExperience: func(pastExperiences []model.TakenExperience, stat *model.Stats) bool {
		stat.Confidence += 10
		return true
	},
}

var EducationUniversity = model.Experience{
	Name:           "Education",
	Description:    "Bachelor's degree in Education focusing on pedagogy, curriculum development, and classroom management.",
	Type:           model.UniStudent,
	Major:          "Education",
	Qualifications: []model.ExperienceQualification{},
	ApplyExperience: func(pastExperiences []model.TakenExperience, stat *model.Stats) bool {
		stat.Confidence += 10
		return true
	},
}
