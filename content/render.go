package main

import (
	"os"
	"text/template"
	"time"
)

var courses = []struct {
	Number             int
	Title              string
	Overview           string
	Objectives         []string
	Readings           []string
	DiscussionPrompts  []string
	DiscussionURL      string
	DiscussionDeadline time.Time
	QuizURL            string
	QuizDeadline       time.Time
}{
	{
		Number:     1,
		Title:      "Open Reproducible Science",
		Overview:   "This module covers tools and methods for ensuring your work is correct, understandable, and reproducible.",
		Objectives: []string{"You will learn how to structure a computational workflow for scientific analysis, including version control, documentation, data provenance, and unit testing."},
		Readings: []string{
			"[Introduction to Earth Data Science Chapter 1](https://www.earthdatascience.org/courses/intro-to-earth-data-science/open-reproducible-science/get-started-open-reproducible-science/)",
			"[Introduction to Earth Data Science Chapter 2](https://www.earthdatascience.org/courses/intro-to-earth-data-science/open-reproducible-science/bash/)",
			"[Introduction to Earth Data Science Chapter 3](https://www.earthdatascience.org/courses/intro-to-earth-data-science/open-reproducible-science/jupyter-python/)",
			"[Andrej Karpathy: Software 2.0](https://medium.com/@karpathy/software-2-0-a64152b37c35)",
			"[NOVA: What Makes Science True?](https://www.youtube.com/watch?v=NGFO0kdbZmk&feature=youtu.be)",
		},
		DiscussionPrompts: []string{
			"What does it mean to practice open and reproducible science, and how could you apply it to your academic or professional life?",
			"Although the readings and NOVA video mainly refer to academic science, how could they be relevant to science practiced in industry?",
			"For the \"Software 2.0\" essay: What is the author talking about? Instead of trying to understand every detail in the essay (although by the end of the semester you should be able to understand a lot of it), focus on the main message: What is Software 2.0 and what are its implications for how science is carried out?",
		},
		DiscussionURL: "https://compass2g.illinois.edu/webapps/discussionboard/do/forum?action=list_threads&course_id=_52490_1&nav=discussion_board_entry&conf_id=_260881_1&forum_id=_442877_1",
		QuizURL:       "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment/2215179/",
	},
}

func main() {
	tmpl := template.Must(template.ParseFiles("content/modules_template.md"))

	w, err := os.Create("content/04.modules.md")
	check(err)
	check(tmpl.Execute(w, courses))
	w.Close()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
