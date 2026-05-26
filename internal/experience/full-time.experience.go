package experience

import "basic-life-sim/internal/model"

var SoftwareEngineer = model.Experience{
	Name:        "Software Engineer",
	Description: "Full-time software engineer designing, developing, and maintaining software systems at a company.",
	Type:        model.FullTimer,
	Major:       "Computer Science",
	Qualifications: []model.ExperienceQualification{
		{Type: "UniStudent", Major: []string{"Computer Science"}, Duration: 0},
		{Type: "PartTimer", Major: []string{"Computer Science"}, Duration: 6},
	},
	ApplyExperience: func(pastExperiences []model.TakenExperience, stat *model.Stats) bool {
		for _, exp := range pastExperiences {
			if exp.Type == model.UniStudent && exp.Major == "Computer Science" {
				stat.Confidence += 15
				return true
			}
			if exp.Type == model.PartTimer && exp.Major == "Computer Science" {
				stat.Confidence += 15
				return true
			}
		}
		return false
	},
}

var SchoolTeacher = model.Experience{
	Name:        "School Teacher",
	Description: "Full-time teacher at a school, responsible for educating students and managing a classroom.",
	Type:        model.FullTimer,
	Major:       "Education",
	Qualifications: []model.ExperienceQualification{
		{Type: "UniStudent", Major: []string{"Education"}, Duration: 0},
		{Type: "PartTimer", Major: []string{"Education"}, Duration: 6},
	},
	ApplyExperience: func(pastExperiences []model.TakenExperience, stat *model.Stats) bool {
		for _, exp := range pastExperiences {
			if exp.Type == model.UniStudent && exp.Major == "Education" {
				stat.Confidence += 15
				return true
			}
			if exp.Type == model.PartTimer && exp.Major == "Education" {
				stat.Confidence += 15
				return true
			}
		}
		return false
	},
}

var Retailer = model.Experience{
	Name:           "Retailer",
	Description:    "Full-time retail worker managing store operations, customer service, and inventory.",
	Type:           model.FullTimer,
	Major:          "Working Class",
	Qualifications: []model.ExperienceQualification{},
	ApplyExperience: func(pastExperiences []model.TakenExperience, stat *model.Stats) bool {
		stat.Confidence += 5
		return true
	},
}
