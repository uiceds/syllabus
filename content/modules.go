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
	HomeworkURL        string
	HomeworkDeadline   time.Time
	LiveMeetingTopics  []string
	Assignment         string
	AssignmentDays     int
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
		HomeworkURL:   "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment/2215179/",
		LiveMeetingTopics: []string{
			"Introduction, syllabus, and getting to know one another",
			"Discussion of readings and technology check: installing Python and related tools",
		},
	},
	{
		Number:     2,
		Title:      "Data science topics for Civil and Environmental Engineering",
		Overview:   "In this module we will learn the types of Civil and Environmental Engineering problems that data science and machine learning can help to answer, and begin to think about topics for course projects.",
		Objectives: []string{},
		Readings: []string{
			"[Tackling Climate Change with Machine Learning](https://arxiv.org/pdf/1906.05433.pdf)",
			"[PANGEO Geoscience Use Cases](https://pangeo.io/use_cases/index.html)",
			"[Kaggle data science competitions](https://www.kaggle.com/)",
			"[Earth Engine Case Studies](https://earthengine.google.com/case_studies/)",
			"[OpenAQ.org](https://openaq.org/)",
			"[Array of Things](https://arrayofthings.github.io/)",
			"[CACES air quality data](https://www.caces.us/data)",
		},
		DiscussionPrompts: []string{
			"What are some ideas you have for course projects, and why do you think they would be useful?",
		},
		LiveMeetingTopics: []string{
			"Group discussion regarding project topics",
			"Select project groups",
		},
		Assignment: `Write an engineering memo describing your project team, the problem 
you plan to solve, and the methods you plan to use to solve it, including the data and algorithm 
you will use.`,
		AssignmentDays: 17,
	},
	{
		Number:      3,
		Title:       "Programming review",
		Overview:    "This course makes extensive use of the Python programming language. By brushing up on our Python skills now, we will make the rest of the course easier.",
		Objectives:  []string{"Students will refresh their skills in basic Python programming."},
		Readings:    []string{"Complete the tutorials at https://learnpython.org, including those under 'Learn the Basics', 'Data Science Tutorials', and 'Advanced Tutorials'."},
		HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment/2143524",
		LiveMeetingTopics: []string{
			"Python & Jupyter exercises and troubleshooting",
		},
	},
	{
		Number:   4,
		Title:    "Data",
		Overview: "Data comes first in data science.",
		Objectives: []string{
			"Students will learn about types of data that are relevant to Civil and Environmental Engineering problems",
			"Students will become acquainted with tools for processing data.",
			"Students will learn strategies for staging and processing large amounts of data.",
		},
		Readings: []string{
			"Recorded lecture on Cloud / High-performance computing, Pangeo, and Earth engine",
		},
		HomeworkURL: "numerical python assignment in prarielearn",
		LiveMeetingTopics: []string{
			"Practice and Discussion",
		},
	},
	{
		Number:     5,
		Title:      "Exploratory data analysis (EDA)",
		Overview:   "The first step in a data science project is getting a feel for the dataset you are working with. This is called Exploratory Data Analysis (EDA).",
		Objectives: []string{"Students will learn how to explore and process an unfamiliar dataset."},
		Readings: []string{
			"Watch mlcourse.ai video lectures on [exploratory data analysis](https://youtu.be/fwWCw_cE5aI) and [visualization](https://www.youtube.com/watch?v=WNoQTNOME5g)",
			"Work through accompanying notebooks [1](https://mlcourse.ai/articles/topic1-exploratory-data-analysis-with-pandas/), [2.1](https://mlcourse.ai/articles/topic2-visual-data-analysis-in-python/) and [2.2](https://mlcourse.ai/articles/topic2-part2-seaborn-plotly/)",
		},
		LiveMeetingTopics: []string{
			"Lecture: Statistics review",
			"EDA group exercises",
		},
		Assignment:     `Students should begin working on EDA for their projects, which will be due in Week 9.`,
		AssignmentDays: 17,
	},
	{
		Number: 6,
		Title:  "Spatial data",
		Overview: `Spatial and Geospatial data are common in Civil and Environmental Engineering, 
but less common in other disciplines that use data science. In this module we will learn 
how to work with these types of data.`,
		Objectives: []string{"Students will learn about processing spatial data, which is common in physical data science"},
		Readings: []string{
			"Recorded lecture on raster vs. vector formats",
			"Recorded lecture on joins and boolean operations",
			"[geopandas tutorial](https://github.com/geopandas/scipy2018-geospatial-data)",
		},
		LiveMeetingTopics: []string{},
		HomeworkURL:       "SpatialDataHomework",
	},
	{
		Number:     7,
		Title:      "Spatial statistics",
		Overview:   "",
		Objectives: []string{"Students will learn how to perform statistical analysis of spatial data."},
		Readings: []string{
			"Recorded Lecture:  Spatial statistics (spatial autocorrelation, Modifiable areal unit problem, kriging)",
			"[PySAL library](https://pysal.org/) and [notebooks](http://pysal.org/notebooks/intro)",
		},
		LiveMeetingTopics: []string{},
	},
	{

		Number:            8,
		Title:             "Mid-way project presentations",
		Overview:          "",
		Objectives:        []string{"Students should be able to access, characterize, and visualize the data for their projects by this point."},
		Readings:          []string{},
		LiveMeetingTopics: []string{},
	},
	{
		Number:     9,
		Title:      "Supervised learning",
		Overview:   "",
		Objectives: []string{"Students will learn what supervised machine learning is and how it can help solve Civil and Environmental Engineering problems."},
		Readings: []string{
			"[framing machine learning](https://developers.google.com/machine-learning/crash-course/framing/video-lecture)",
			"[gradient descent](https://developers.google.com/machine-learning/crash-course/descending-into-ml/video-lecture)",
			"[optimization](https://developers.google.com/machine-learning/crash-course/reducing-loss/video-lecture)",
			"[tensorflow](https://developers.google.com/machine-learning/crash-course/first-steps-with-tensorflow/toolkit)",
			"[generalization](https://developers.google.com/machine-learning/crash-course/generalization/video-lecture)",
			"[training and testing](https://developers.google.com/machine-learning/crash-course/training-and-test-sets/video-lecture)",
			"[validation](https://developers.google.com/machine-learning/crash-course/validation/check-your-intuition)",
		},
		LiveMeetingTopics: []string{},
		HomeworkURL:       "SupervisedLearningHomework",
	},
	{
		Number:     10,
		Title:      "Unsupervised learning",
		Overview:   "",
		Objectives: []string{"Students will learn about basic unsupervised learning algorithms and how they can be used on Civil and Environmental Engineering applications."},
		Readings: []string{
			"[unsupervised learning](https://www.youtube.com/watch?v=jAA2g9ItoAc)",
			"[clustering](https://www.youtube.com/watch?v=Ev8YbxPu_bQ)",
			"[mlcourse.ai workbook](https://mlcourse.ai/articles/topic7-unsupervised/)",
		},
		LiveMeetingTopics: []string{
			"In class, we will work through some applications to environmental data and discuss how supervised learning can be applied to student projects.",
		},
		HomeworkURL: "UnsupervisedLearningHomework",
	},
	{
		Number:   11,
		Title:    "Deep learning",
		Overview: "",
		Objectives: []string{
			"Students will learn about deep learning, the opportunities and drawbacks it presents, and applications to environmental problems.",
		},
		Readings: []string{
			"[Introduction to Neural Networks](https://developers.google.com/machine-learning/crash-course/introduction-to-neural-networks/video-lecture)",
			"[Multi-Class Neural Networks](https://developers.google.com/machine-learning/crash-course/multi-class-neural-networks/video-lecture)",
			"Recorded Lecture: hyperparameter optimization and inductive biases",
		},
		HomeworkURL:       "DeepLearningHomework",
		LiveMeetingTopics: []string{},
	},
	{
		Number:      12,
		Title:       "Project work",
		Overview:    "Students will work on their course projects",
		Objectives:  []string{},
		Readings:    []string{},
		HomeworkURL: "DeepLearningHomework",
		LiveMeetingTopics: []string{
			"During class time we will work together to troubleshoot student course projects.",
			"Students can sign up for time slots where they can present a problem they have encountered and the class will discuss possible solutions.",
		},
	},
	{
		Number:      13,
		Title:       "Final exam; final project presentations and reports",
		Overview:    "",
		Objectives:  []string{"Students should have completed a project where they access and explore a civil or environmental dataset and use it to answer a scientific question."},
		Readings:    []string{},
		HomeworkURL: "DeepLearningHomework",
		LiveMeetingTopics: []string{
			"Written report due",
			"Oral presentations to class",
			"Comprehensive final exam",
		},
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
