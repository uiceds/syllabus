package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var cal = flag.Bool("cal", false, "Whether to create calendar events")

const courseInstance = "fall2020"

const calendarID = "c_fqvrphqptlccpp6pubokjsraj0@group.calendar.google.com"

var startDate, midtermExamStart, midtermExamEnd, finalExamStart, finalExamEnd time.Time

var loc *time.Location

func init() {
	flag.Parse()

	var err error
	loc, err = time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}
	startDate = time.Date(2020, time.August, 24, 12, 0, 0, 0, loc)

	midtermExamStart = time.Date(2020, time.October, 15, 12, 0, 0, 0, loc)
	midtermExamEnd = time.Date(2020, time.October, 16, 12, 0, 0, 0, loc)
	finalExamStart = time.Date(2020, time.December, 11, 13, 30, 0, 0, loc)
	finalExamEnd = time.Date(2020, time.December, 11, 16, 30, 0, 0, loc)
}

func startDates(modules []module) map[int64]time.Time {
	g := simple.NewDirectedGraph()
	mods := make(map[int64]graph.Node)
	for _, m := range modules {
		mods[m.ID()] = m
	}
	for _, m := range modules {
		g.AddNode(m)
		for _, p := range m.Parents {
			g.SetEdge(g.NewEdge(mods[p], m))
		}
	}

	startDates := make(map[int64]time.Time)
	t := traverse.BreadthFirst{
		Visit: func(n graph.Node) {
			m := n.(module)
			if len(m.Parents) == 0 { // Root starts at the start date
				startDates[m.ID()] = startDate
			} else { // Start at date that latest parent ends.
				d := startDate
				for _, p := range m.Parents {
					mp := g.Node(p).(module)
					dp := startDates[p].Add(time.Hour * 24 * time.Duration(mp.NumDays))
					if dp.After(d) {
						d = dp
					}
				}
				startDates[m.ID()] = d
			}
		},
		Traverse: func(e graph.Edge) bool {
			return e.From().ID() == e.To().ID()-1
		},
	}
	t.Walk(g, g.Node(1), nil)
	return startDates
}

type module struct {
	Number     int64
	Title      string
	Parents    []int64
	NumDays    int
	Overview   string
	Objectives []string
	Readings   []string

	DiscussionPrompts []string
	DiscussionURL     string
	// DiscussionDelay is the number of lectures to delay the
	// discussion due date by.
	DiscussionDelay int

	HomeworkURL string
	// DiscussionDelay is the number of lectures to delay the
	// discussion due date by.
	HomeworkDelay int

	LiveMeetingTopics     []string
	ProjectAssignment     string
	ProjectAssignmentDays int
}

func (m module) ID() int64 { return m.Number }

var modules = []module{
	{
		Number:   1,
		NumDays:  7,
		Title:    "Open Reproducible Science",
		Overview: "This module covers tools and methods for ensuring your work is correct, understandable, and reproducible. ",
		Objectives: []string{
			"List the Bash commands for doing different computer operations",
			"Describe how to organize data files to improve transparency and reproducibility",
			"Define the syntax for the Markdown text formatting language",
			"List the Git and Github commands and operations for performing different operations",
		},
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
		DiscussionURL:   "https://compass2g.illinois.edu/webapps/discussionboard/do/forum?action=list_threads&course_id=_52490_1&nav=discussion_board_entry&conf_id=_260881_1&forum_id=_442877_1",
		DiscussionDelay: 1,
		HomeworkURL:     "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment/2215179/",
		HomeworkDelay:   1,
		LiveMeetingTopics: []string{
			"Introduction, syllabus, and getting to know one another ([slides](https://docs.google.com/presentation/d/1U-xdr_lPprNl5HiMTZ-pG0p9lRhR9FR0zMF3pwzVPjw/edit?usp=sharing))",
			"Discussion of readings and technology check: installing Python and related tools ([notes](https://github.com/uiceds/syllabus/blob/master/modules/module1/demo/README.md))",
		},
	},
	{
		Number:   2,
		NumDays:  7,
		Parents:  []int64{1},
		Title:    "Data science topics for Civil and Environmental Engineering",
		Overview: "In this module we will learn the types of Civil and Environmental Engineering problems that data science and machine learning can help to answer, and begin to think about topics for course projects.",
		Objectives: []string{
			"interpret how data science can be used in support of CEE",
			"formulate a data science problem statement",
			"describe a strategy for solving the problem",
		},
		Readings: []string{
			"[Deep Learning State of the Art (This is an introduction to what is currently possible with data science)](https://www.youtube.com/watch?v=0VH1Lim8gL8&feature=emb_logo)",
			"[Tackling Climate Change with Machine Learning](https://arxiv.org/pdf/1906.05433.pdf)",
			"[Kaggle data science competitions](https://www.kaggle.com/)",
			"[Earth Engine Case Studies](https://earthengine.google.com/case_studies/)",
			"[OpenAQ.org](https://openaq.org/)",
			"[CACES air quality data](https://www.caces.us/data)",
			"[EIA Energy Data](https://www.eia.gov/)",
			"[UCI Machine Learning Datasets](https://archive.ics.uci.edu/ml/datasets.php)",
			"[Dataset for defect decection in Concrete](https://zenodo.org/record/2620293#.X0Gbd2jYpaY)",
			"https://medium.com/towards-artificial-intelligence/best-datasets-for-machine-learning-data-science-computer-vision-nlp-ai-c9541058cf4f",
			"https://www.visualdata.io/discovery",
		},
		DiscussionURL: "https://compass2g.illinois.edu/webapps/discussionboard/do/forum?action=list_threads&course_id=_52490_1&nav=discussion_board_entry&conf_id=_260881_1&forum_id=_442878_1",
		DiscussionPrompts: []string{
			"A big part of this course will be a semester-long project, where you will use a dataset to answer a question relevant to Civil or Environmental Engineering. What are some ideas you have for course projects, how are they related to Civil or Environmental Engineering, why do you think they would be useful, and what dataset would they be based on?",
		},
		LiveMeetingTopics: []string{
			"What is data science, and how is it relevant to CEE? A discussion.",
			"Group discussion regarding project topics",
		},
	},
	{
		Number:   3,
		NumDays:  7,
		Parents:  []int64{2},
		Title:    "Programming review",
		Overview: "This course makes extensive use of the Python programming language. By brushing up on our Python skills now, we will make the rest of the course easier.",
		Objectives: []string{
			"Express abstract concepts using Python syntax",
			"Solve mathematical problems using Python",
		},
		Readings: []string{"Complete the tutorials at https://learnpython.org, including those under 'Learn the Basics', 'Data Science Tutorials', and 'Advanced Tutorials'. " +
			"These readings may not include all the information you need to complete the homework, which will allow you to practice researching concepts on the internet."},
		HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment/2143524",
		LiveMeetingTopics: []string{
			"Python & Jupyter exercises and troubleshooting",
			"",
		},
		ProjectAssignment:     `Choose a project and complete a literature review.`,
		ProjectAssignmentDays: 17,
	},
	{
		Number:   4,
		NumDays:  7,
		Parents:  []int64{3},
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
			"",
		},
	},
	{
		Number:     5,
		NumDays:    7,
		Parents:    []int64{3, 4},
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
		ProjectAssignment:     `Exploratory data analysis for their projects, with midterm presentation.`,
		ProjectAssignmentDays: 17,
	},
	{
		Number:  6,
		Parents: []int64{3, 5},
		NumDays: 7,
		Title:   "Spatial data",
		Overview: `Spatial and Geospatial data are common in Civil and Environmental Engineering, 
but less common in other disciplines that use data science. In this module we will learn 
how to work with these types of data.`,
		Objectives: []string{"Students will learn about processing spatial data, which is common in physical data science"},
		Readings: []string{
			"Recorded lecture on raster vs. vector formats",
			"Recorded lecture on joins and boolean operations",
			"[geopandas tutorial](https://github.com/geopandas/scipy2018-geospatial-data)",
		},
		LiveMeetingTopics: []string{
			"",
			"",
		},
		HomeworkURL: "SpatialDataHomework",
	},
	{
		Number:     7,
		Parents:    []int64{6},
		NumDays:    7,
		Title:      "Spatial statistics",
		Overview:   "",
		Objectives: []string{"Students will learn how to perform statistical analysis of spatial data."},
		Readings: []string{
			"Recorded Lecture:  Spatial statistics (spatial autocorrelation, Modifiable areal unit problem, kriging)",
			"[PySAL library](https://pysal.org/) and [notebooks](http://pysal.org/notebooks/intro)",
		},
		LiveMeetingTopics: []string{
			"",
			"",
		},
	},
	{

		Number:     8,
		Parents:    []int64{2, 7},
		NumDays:    7,
		Title:      "Mid-way project presentations",
		Overview:   "",
		Objectives: []string{"Students should be able to access, characterize, and visualize the data for their projects by this point."},
		Readings:   []string{},
		LiveMeetingTopics: []string{
			"",
			"",
		},
		ProjectAssignment:     `Project Kaggle Competition.`,
		ProjectAssignmentDays: 40,
	},
	{
		Number:     9,
		Parents:    []int64{5, 8},
		NumDays:    7,
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
		LiveMeetingTopics: []string{
			"",
			"",
		},
		HomeworkURL: "SupervisedLearningHomework",
	},
	{
		Number:     10,
		Parents:    []int64{5, 9},
		NumDays:    7,
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
			"",
		},
		HomeworkURL: "UnsupervisedLearningHomework",
	},
	{
		Number:   11,
		Parents:  []int64{9, 10},
		NumDays:  7,
		Title:    "Neural Networks",
		Overview: "",
		Objectives: []string{
			"Students will learn about deep learning, the opportunities and drawbacks it presents, and applications to environmental problems.",
		},
		Readings: []string{
			"[Introduction to Neural Networks](https://developers.google.com/machine-learning/crash-course/introduction-to-neural-networks/video-lecture)",
			"[Multi-Class Neural Networks](https://developers.google.com/machine-learning/crash-course/multi-class-neural-networks/video-lecture)",
			"Recorded Lecture: hyperparameter optimization and inductive biases",
		},
		HomeworkURL: "DeepLearningHomework",
		LiveMeetingTopics: []string{
			"Remember to vote!",
			"",
		},
	},
	{
		Number:   12,
		Parents:  []int64{11},
		NumDays:  7,
		Title:    "Convolutional Neural Networks",
		Overview: "",
		Objectives: []string{
			"Students will learn about deep learning, the opportunities and drawbacks it presents, and applications to environmental problems.",
		},
		Readings: []string{
			"[Introduction to Neural Networks](https://developers.google.com/machine-learning/crash-course/introduction-to-neural-networks/video-lecture)",
			"[Multi-Class Neural Networks](https://developers.google.com/machine-learning/crash-course/multi-class-neural-networks/video-lecture)",
			"Recorded Lecture: hyperparameter optimization and inductive biases",
		},
		HomeworkURL: "DeepLearningHomework",
		LiveMeetingTopics: []string{
			"",
			"",
		},
		ProjectAssignment:     `Project Final Presentation and Report`,
		ProjectAssignmentDays: 22,
	},
	{
		Number:   13,
		Parents:  []int64{11, 12},
		NumDays:  7,
		Title:    "Neural Networks for Sequences",
		Overview: "",
		Objectives: []string{
			"Students will learn about deep learning, the opportunities and drawbacks it presents, and applications to environmental problems.",
		},
		Readings: []string{
			"[Introduction to Neural Networks](https://developers.google.com/machine-learning/crash-course/introduction-to-neural-networks/video-lecture)",
			"[Multi-Class Neural Networks](https://developers.google.com/machine-learning/crash-course/multi-class-neural-networks/video-lecture)",
			"Recorded Lecture: hyperparameter optimization and inductive biases",
		},
		HomeworkURL: "DeepLearningHomework",
		LiveMeetingTopics: []string{
			"",
			"",
		},
	},
	{
		Number:      14,
		Parents:     []int64{2, 13},
		NumDays:     7,
		Title:       "Project work",
		Overview:    "Students will work on their course projects",
		Objectives:  []string{},
		Readings:    []string{},
		HomeworkURL: "",
		LiveMeetingTopics: []string{
			"",
			"",
		},
	},
	{
		Number:      15,
		Parents:     []int64{2, 14},
		NumDays:     9,
		Title:       "Final exam; final project presentations and reports",
		Overview:    "",
		Objectives:  []string{"Students should have completed a project where they access and explore a civil or environmental dataset and use it to answer a scientific question."},
		Readings:    []string{},
		HomeworkURL: "",
		LiveMeetingTopics: []string{
			"",
			"",
			"",
		},
	},
}

func nextLecture(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Tuesday || w == time.Thursday {
			return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, d.Location())
		}
	}
}

func nextOfficeHour(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Monday || w == time.Wednesday {
			return time.Date(d.Year(), d.Month(), d.Day(), 10, 30, 0, 0, d.Location())
		}
	}
}

const dateFormat = "Mon 1/2/2006, 15:04 MST"
const dayFormat = "1/2/2006"

func moduleStart(m module, dates map[int64]time.Time) time.Time {
	return dates[m.ID()]
}
func discussionAssigned(m module, dates map[int64]time.Time) time.Time {
	d := dates[m.ID()].Add(-7 * 24 * time.Hour)
	if d.Before(startDate) {
		return startDate
	}
	return d
}
func discussionInitialDeadline(m module, dates map[int64]time.Time) time.Time {
	d := nextLecture(dates[m.ID()])
	for i := 0; i < m.DiscussionDelay; i++ {
		d = nextLecture(d)
	}
	return d
}
func discussionResponseDeadline(m module, dates map[int64]time.Time) time.Time {
	d := nextLecture(nextLecture(dates[m.ID()]))
	for i := 0; i < m.DiscussionDelay; i++ {
		d = nextLecture(d)
	}
	return d
}
func homeworkAssigned(m module, dates map[int64]time.Time) time.Time {
	d := dates[m.ID()].Add(-7 * 24 * time.Hour)
	if d.Before(startDate) {
		return startDate
	}
	return d
}
func homeworkDeadline1(m module, dates map[int64]time.Time) time.Time {
	d := nextLecture(dates[m.ID()])
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextLecture(d)
	}
	return d
}
func homeworkDeadline2(m module, dates map[int64]time.Time) time.Time {
	d := nextLecture(nextLecture(dates[m.ID()]))
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextLecture(d)
	}
	return d
}
func homeworkDeadline3(m module, dates map[int64]time.Time) time.Time {
	d := nextLecture(dates[m.ID()].Add(14 * 24 * time.Hour))
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextLecture(d)
	}
	return d
}
func assignmentDeadline(m module, dates map[int64]time.Time) time.Time {
	return nextLecture(dates[m.ID()].Add(time.Duration(m.ProjectAssignmentDays) * 24 * time.Hour))
}

func main() {
	dates := startDates(modules)

	funcMap := template.FuncMap{
		"StartDate": func(m module) string {
			return moduleStart(m, dates).Format(dayFormat)
		},
		"DiscussionAssigned": func(m module) string {
			return discussionAssigned(m, dates).Format(dayFormat)
		},
		"DiscussionInitialDeadline": func(m module) string {
			return discussionInitialDeadline(m, dates).Format(dateFormat)
		},
		"DiscussionResponseDeadline": func(m module) string {
			return discussionResponseDeadline(m, dates).Format(dateFormat)
		},
		"HomeworkAssigned": func(m module) string {
			return homeworkAssigned(m, dates).Format(dayFormat)
		},
		"HomeworkDeadline1": func(m module) string {
			return homeworkDeadline1(m, dates).Format(dateFormat)
		},
		"HomeworkDeadline2": func(m module) string {
			return homeworkDeadline2(m, dates).Format(dateFormat)
		},
		"HomeworkDeadline3": func(m module) string {
			return homeworkDeadline3(m, dates).Format(dateFormat)
		},
		"AssignmentDeadline": func(m module) string {
			return assignmentDeadline(m, dates).Format(dateFormat)
		},
		"ModuleLink": func(m module) string {
			return strings.Replace(strings.ToLower(m.Title), " ", "-", -1)
		},
	}

	tmpl := template.Must(template.New("root").Funcs(funcMap).ParseFiles("modules_template.md"))

	w, err := os.Create("04.modules.md")
	check(err)

	schedule := struct {
		MidtermExamStart, MidtermExamEnd string
		FinalExamStart, FinalExamEnd     string
		Modules                          []module
	}{
		MidtermExamStart: midtermExamStart.Format(dateFormat),
		MidtermExamEnd:   midtermExamEnd.Format(dateFormat),
		FinalExamStart:   finalExamStart.Format(dateFormat),
		FinalExamEnd:     finalExamEnd.Format(dateFormat),
		Modules:          modules,
	}

	check(tmpl.ExecuteTemplate(w, "modules_template.md", schedule))
	w.Close()

	if *cal {
		createCalendar(modules, dates, funcMap)
	}
}

func createCalendar(modules []module, startDates map[int64]time.Time, funcs template.FuncMap) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// Delete events
	events, err := srv.Events.List(calendarID).SingleEvents(true).MaxResults(2500).Do()
	check(err)
	for _, item := range events.Items {
		fmt.Println("deleting calendar event", item.Id)
		check(srv.Events.Delete(calendarID, item.Id).Do())
	}

	// Add events
	for _, m := range modules {
		d := startDates[m.Number]
		fmt.Println("Adding events to calendar for module:", m.Number)
		m.lecturesToCalendar(srv, d)
		m.officeHoursToCalendar(srv, d)
		m.discussionToCalendar(srv, startDates)
		m.homeworkToCalendar(srv, startDates)
		m.assignmentToCalendar(srv, startDates)
		m.examsToCalendar(srv)
	}
}

func (m module) lecturesToCalendar(srv *calendar.Service, startDate time.Time) {
	d := startDate
	for i := 0; i < 100; i++ {
		d = nextLecture(d)
		if d.After(startDate.Add(time.Duration(m.NumDays) * 24 * time.Hour)) {
			break
		}
		_, err := srv.Events.Insert(calendarID, &calendar.Event{
			Summary:     fmt.Sprintf("DS-CEE Zoom meeting: %s", m.Title),
			Location:    "https://compass2g.illinois.edu/webapps/blackboard/content/launchLink.jsp?course_id=_52490_1&tool_id=_2918_1&tool_type=TOOL&mode=view&mode=reset",
			Description: m.LiveMeetingTopics[i],
			Status:      "confirmed",
			Start: &calendar.EventDateTime{
				DateTime: d.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: d.Add(80 * time.Minute).Format(time.RFC3339),
			},
		}).Do()
		check(err)
	}
}

func (m module) officeHoursToCalendar(srv *calendar.Service, startDate time.Time) {
	d := startDate.Add(-24 * time.Hour)
	for i := 0; i < 2; i++ {
		d = nextOfficeHour(d)
		_, err := srv.Events.Insert(calendarID, &calendar.Event{
			Summary:  "DS-CEE Office hours",
			Location: "https://compass2g.illinois.edu/webapps/blackboard/content/launchLink.jsp?course_id=_52490_1&tool_id=_2918_1&tool_type=TOOL&mode=view&mode=reset",
			Status:   "confirmed",
			Start: &calendar.EventDateTime{
				DateTime: d.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: d.Add(90 * time.Minute).Format(time.RFC3339),
			},
		}).Do()
		check(err)
	}
}

func (m module) discussionToCalendar(srv *calendar.Service, dates map[int64]time.Time) {
	if m.DiscussionURL == "" {
		return
	}
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("Discussion Assigned: %s", m.Title),
		Location:    m.DiscussionURL,
		Status:      "confirmed",
		Description: fmt.Sprintf("https://uiceds.github.io/syllabus/#module-%d-discussion", m.Number),
		Start: &calendar.EventDateTime{
			Date: discussionAssigned(m, dates).Format("2006-01-02"),
		},
		End: &calendar.EventDateTime{
			Date: discussionAssigned(m, dates).Format("2006-01-02"),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("Initial Post Deadline: %s", m.Title),
		Location:    m.DiscussionURL,
		Status:      "confirmed",
		Description: fmt.Sprintf("https://uiceds.github.io/syllabus/#module-%d-discussion", m.Number),
		Start: &calendar.EventDateTime{
			DateTime: discussionInitialDeadline(m, dates).Add(-time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: discussionInitialDeadline(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("Response Posts Deadline: %s", m.Title),
		Location:    m.DiscussionURL,
		Status:      "confirmed",
		Description: fmt.Sprintf("https://uiceds.github.io/syllabus/#module-%d-discussion", m.Number),
		Start: &calendar.EventDateTime{
			DateTime: discussionResponseDeadline(m, dates).Add(-time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: discussionResponseDeadline(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

func (m module) homeworkToCalendar(srv *calendar.Service, dates map[int64]time.Time) {
	if m.HomeworkURL == "" {
		return
	}
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("Homework Assigned: %s", m.Title),
		Location:    m.HomeworkURL,
		Description: fmt.Sprintf("https://uiceds.github.io/syllabus/#module-%d-homework", m.Number),
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			Date: homeworkAssigned(m, dates).Format("2006-01-02"),
		},
		End: &calendar.EventDateTime{
			Date: homeworkAssigned(m, dates).Format("2006-01-02"),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("110%% credit Homework deadline: %s", m.Title),
		Location:    m.HomeworkURL,
		Description: fmt.Sprintf("https://uiceds.github.io/syllabus/#module-%d-homework", m.Number),
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: homeworkDeadline1(m, dates).Add(-time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: homeworkDeadline1(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("100%% credit Homework deadline: %s", m.Title),
		Location:    m.HomeworkURL,
		Description: fmt.Sprintf("https://uiceds.github.io/syllabus/#module-%d-homework", m.Number),
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: homeworkDeadline2(m, dates).Add(-time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: homeworkDeadline2(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("80%% credit Homework deadline: %s", m.Title),
		Location:    m.HomeworkURL,
		Description: fmt.Sprintf("https://uiceds.github.io/syllabus/#module-%d-homework", m.Number),
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: homeworkDeadline3(m, dates).Add(-time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: homeworkDeadline3(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

func (m module) assignmentToCalendar(srv *calendar.Service, dates map[int64]time.Time) {
	if m.ProjectAssignment == "" {
		return
	}
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Project Activity Assigned",
		Description: m.ProjectAssignment,
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			Date: moduleStart(m, dates).Format("2006-01-02"),
		},
		End: &calendar.EventDateTime{
			Date: moduleStart(m, dates).Format("2006-01-02"),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Project Activity Due",
		Description: m.ProjectAssignment,
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: assignmentDeadline(m, dates).Add(-time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: assignmentDeadline(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

func (m module) examsToCalendar(srv *calendar.Service) {
	if m.ProjectAssignment == "" {
		return
	}
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Midterm Exam",
		Description: "",
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: midtermExamStart.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: midtermExamEnd.Format(time.RFC3339),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Final Exam",
		Description: "",
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: finalExamStart.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: finalExamEnd.Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
