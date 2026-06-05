package experience

import "github.com/kamil5b/basic-life-sim/internal/model"

func init() {
	model.RegisterExperience(AllExperiences...)
}

var AllExperiences = []model.Experience{
	ComputerScienceUniversity,
	EducationUniversity,
	InternSoftwareEngineer,
	PrivateTeacher,
	SoftwareEngineer,
	SchoolTeacher,
	Retailer,
}
