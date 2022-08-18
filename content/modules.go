package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var cal = flag.Bool("cal", false, "Whether to create calendar events")
var plPath = flag.String("pl-path", "../../pl-cee498ds/courseInstances/Fa2022", "Path to PrairieLearn course repository")

var courseInstance string

const calendarID = "c_fqvrphqptlccpp6pubokjsraj0@group.calendar.google.com"

const plWebsite = "https://www.prairielearn.org/pl/course_instance/128749/assessments"

var startDate, finalExamStart, finalExamEnd time.Time

var classDuration = 80 * time.Minute
var examDuration = 24 * time.Hour

var loc *time.Location

func init() {
	flag.Parse()
	courseInstance = *plPath

	var err error
	loc, err = time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}
	startDate = time.Date(2022, time.August, 23, 12, 0, 0, 0, loc)

	finalExamStart = time.Date(2022, time.December, 13, 8, 00, 0, 0, loc)
	finalExamEnd = time.Date(2022, time.December, 14, 8, 00, 0, 0, loc)
}

func startDates(modules []module) map[int64]time.Time {
	startDates := make(map[int64]time.Time)
	mods := make(map[int64]module)
	for _, m := range modules {
		mods[m.Number] = m
	}
	for _, m := range modules {
		// Start at date that latest parent ends.
		d := startDate
		for _, p := range m.Parents {
			mp := mods[p]
			dp := nextModuleStart(mp, startDates[p])
			if dp.After(d) {
				d = dp
			}
		}
		startDates[m.ID()] = d
	}
	return startDates
}

type project struct {
	ID       string
	Number   int
	Assigned time.Time
	Due      time.Time
}

func projects(modules []module) []project {
	var i int
	var proj []project
	for _, m := range modules {
		if m.ProjectAssignment == "" {
			continue
		}
		var assigned time.Time
		if i == 0 {
			assigned = startDate
		} else {
			assigned = proj[i-1].Due
		}
		if strings.Contains(m.ProjectAssignment, "exploratory") {
			proj = append(proj, project{
				ID:       m.ProjectAssignment,
				Number:   i + 1,
				Assigned: assigned,
				Due:      nextFridayNight(projectAssignmentDue(m, startDates(modules))), // Avoid conflict with exam. TODO: Fix
			})
		} else {
			proj = append(proj, project{
				ID:       m.ProjectAssignment,
				Number:   i + 1,
				Assigned: assigned,
				Due:      projectAssignmentDue(m, startDates(modules)),
			})
		}
		i++
	}
	return proj
}

type module struct {
	Number     int64
	Title      string
	Parents    []int64
	Overview   string
	Objectives []string
	Readings   []string

	DiscussionPrompts []string
	DiscussionURL     string
	// DiscussionDelay is the number of lectures to delay the
	// discussion due date by.
	DiscussionDelay int

	HomeworkURL string
	// HomeworkDelay is the number of lectures to delay the
	// homework due date by.
	HomeworkDelay int

	// PLName is the name of the module in PrairieLearn.
	PLName string

	// ClassNames holds the PraireLearn class names for this module.
	ClassNames []string

	ProjectAssignment     string
	ProjectAssignmentDays int
}

func (m module) ID() int64 { return m.Number }

var modules = []module{
	{
		Number:     0,
		Title:      "Introduction and motivating problems",
		Overview:   "In this module we will get to know each other and cover the format of the course, its contents, and expectations.",
		PLName:     "intro",
		ClassNames: []string{"intro"},
	},
	{
		Number:  1,
		Parents: []int64{0},
		Title:   "Linear algebra review and intro to the Julia Language",
		Overview: `In this course, we will use two key tools: linear algebra and the Julia programming language. 
		You should already be familiar with linear algebra, so we will only briefly review it here. 
		You're not expected to know anything about the Julia language before starting this class, but you are
		expected to have completed a basic computer programming class (similar to CS101) using some computing language.`,
		Objectives: []string{
			"Operate the Julia language programming interface, using both the REPL and Pluto notebooks",
			"Use variables, arrays, conditional statements, loops, and functions to process data using Julia",
			"Use Julia's built-in and library functions to operate on text and data",
			"Solve systems of equations using linear algebra in the Julia language",
			"Debug Julia programs to fix programming errors",
		},
		PLName: "intro_julia_la",
		ClassNames: []string{
			"julia_basics_1",
			"julia_basics_2",
			"linalg",
		},
	},
	{
		Number:   2,
		Parents:  []int64{1},
		Title:    "Open reproducible science",
		Overview: "This module covers tools and methods for ensuring your work is correct, understandable, and reproducible.",
		Objectives: []string{
			"Apply the theory of ['tidy data'](https://r4ds.had.co.nz/tidy-data.html) to wrangle a tabular dataset into tidy format, using for example the `groupby`, and `combine`  functions in DataFrames.jl",
			"Evaluate an unfamiliar dataset with exploratory statistical analysis, using for example the `filter` and `select` functions in DataFrames.jl as well as array indexing and basic descriptive statistics",
			"Create exploratory visualizations for tabular, array, and image data using Plots.jl and StatsPlots.jl",
			"Apply git and GitHub.com for distributed version control and collaboration on group projects",
		},
		PLName: "reproducible",
		ClassNames: []string{
			"git",
			"viz",
			"wrangle",
		},
	},
	{
		Number:   3,
		Parents:  []int64{2},
		Title:    "Singular value decomposition and principle component analysis",
		Overview: "SVD and PCA fundamental algorithms for data processing and analysis. We will learn how they work and how they can be applied to gain insight from data.",
		Objectives: []string{
			"Apply the SVD and PCA algorithms to create a low-rank approximation of a dataset",
			"Interpret the results of the algorithms in a given context, including the significance of the resulting values and how much of the variance in the original dataset is represented in the low-rank approximation",
		},
		PLName: "svd_pca",
		ClassNames: []string{
			"svd",
			"pca",
		},
	},
	{
		Number:   4,
		Parents:  []int64{3},
		Title:    "Fourier and wavelet transforms",
		Overview: `Fourier and wavelet transforms are powerful methods for coordinate transformation, data compression, and feature engineering and are used in almost every field of science and engineering.`,
		Objectives: []string{
			"Apply the FFT, Gabor transform, and Wavelet transform algorithms to determine the frequency spectra of a dataset",
			"Interpret the results of the algorithms in a given context, including the significance of the resulting values",
		},
		PLName: "fourier",
		ClassNames: []string{
			"fourier",
			"Exam 1: Computational thinking",
			"fft",
			"wavelet",
		},
		ProjectAssignment: "project/selection",
	},
	{
		Number:   5,
		Parents:  []int64{4},
		Title:    "Regression",
		Overview: `In this module, we will learn how to use regression to predict the value of a dependent variable given a set of independent variables.`,
		Objectives: []string{
			"Apply the gradient descent algorithm to mimimize error between a model prediction and observations",
			"Design and implement a linear regression model to predict a dependent variable in a dataset when given independent variables",
			"Apply regularization to the model to avoid overfitting",
			"Apply feature selection and engineering and coordinate transformation to a dataset to improve regression performance",
		},
		PLName: "regression",
		ClassNames: []string{
			"regression",
			"regularization",
			"model_selection",
			"Exam 2: Coordinate transforms",
		},
	},
	{
		Number:   6,
		Parents:  []int64{5},
		Title:    "Machine learning",
		Overview: `In this module, we learn about two popular machine learning algorithms: k-means and decision trees.`,
		Objectives: []string{
			"Implement the k-means algorithm to divide a dataset into clusters",
			"Design and implement a decision tree model to predict a dependent variable in a dataset when given independent variables",
		},
		PLName: "machine_learning",
		ClassNames: []string{
			"k-means",
			"classification_trees",
		},
		ProjectAssignment: "project/exploratory",
	},
	{
		Number:   7,
		Parents:  []int64{6},
		Title:    "Neural networks",
		Overview: `In this module, we will learn how to implement and use both fully-connected and convolutional neural networks.`,
		Objectives: []string{
			"Train a neural network to for regression and classification",
			"Identify and debug common problems with neural network training",
		},
		PLName: "neural_nets",
		ClassNames: []string{
			"neural_nets1",
			"neural_nets2",
			"conv_nets",
		},
	},
	{
		Number:   8,
		Parents:  []int64{7},
		Title:    "Data-driven dynamical systems",
		PLName:   "data_driven_dynamics",
		Overview: `In this module, we will apply the machine learning techniques we have learned so far to dynamical systems and the differential equations that describe them.`,
		Objectives: []string{
			"Apply gradient descent to fit the parameters of a system of differential equations to observed data",
			"Implement a Neural ODE to make data-driven predictions of the evolution of a dynamical system",
		},
		ClassNames: []string{
			"Voting Day! Check [here](https://champaigncountyclerk.com/elections/my-voting-information/my-polling-place) for where to vote.",
			"param_fitting",
			"neural_odes",
		},
		ProjectAssignment: "project/modeling",
	},
	{
		Number:  9,
		Parents: []int64{8},
		Title:   "Fairness in machine learning",
		Overview: `Machine learning models can contain bias, which is especially important as these models become more integrated in to human society.
		We will learn how to detect and minimize this bias.`,
		Objectives: []string{
			"Use disaggregated testing to detect bias in machine learning models",
			"Design and construct models to minimize any detected bias",
		},
		PLName: "fairness",
		ClassNames: []string{
			"fairness",
		},
		ProjectAssignment: "project/rough_draft",
	},
	{
		Number:  -1,
		Parents: []int64{9},
		Title:   "Fall break",
		ClassNames: []string{
			"Fall break",
			"Fall break",
		},
	},
	{
		Number:   10,
		Parents:  []int64{-1},
		Title:    "Final projects",
		Overview: `In this module we will present the results of our semester projects.`,
		ClassNames: []string{
			"Project workshop (or presentations?)",
			"Final project presentations",
			"Final project presentations",
		},
		ProjectAssignment: "project/final",
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

func nextSundayNight(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Sunday {
			return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, d.Location())
		}
	}
}

func nextFridayNight(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Friday {
			return time.Date(d.Year(), d.Month(), d.Day(), 17, 0, 0, 0, d.Location())
		}
	}
}

func nextTessumOfficeHour(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Thursday {
			return time.Date(d.Year(), d.Month(), d.Day(), 11, 00, 0, 0, d.Location())
		}
	}
}
func nextGuoOfficeHour(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Monday {
			return time.Date(d.Year(), d.Month(), d.Day(), 9, 00, 0, 0, d.Location())
		}
	}
}
func nextWangOfficeHour(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Wednesday {
			return time.Date(d.Year(), d.Month(), d.Day(), 12, 00, 0, 0, d.Location())
		}
	}
}

const dateFormat = "Mon 1/2/2006, 15:04 MST"
const dayFormat = "1/2/2006"

func moduleStart(m module, dates map[int64]time.Time) time.Time {
	return dates[m.ID()]
}
func nextModuleStart(m module, startDate time.Time) time.Time {
	d := startDate
	for i := 0; i < len(m.ClassNames); i++ {
		d = nextLecture(d)
	}
	return d
}
func projectAssignmentDue(m module, dates map[int64]time.Time) time.Time {
	return nextFridayNight(moduleStart(m, dates))
}
func discussionAssigned(m module, dates map[int64]time.Time) time.Time {
	d := dates[m.ID()].Add(-7 * 24 * time.Hour)
	if d.Before(startDate) {
		return startDate
	}
	return d
}
func preclassAssigned(m module, dates map[int64]time.Time, n int) time.Time {
	return homeworkAssigned(m, dates)
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
func classSession(m module, dates map[int64]time.Time, num int) time.Time {
	d := dates[m.ID()]
	for i := 0; i < num; i++ {
		d = nextLecture(d)
	}
	return d
}
func contactHours(m module) float64 {
	h := 0.0
	for _, c := range m.ClassNames {
		if !strings.Contains(strings.ToLower(c), "exam") {
			h += classDuration.Hours()
		}
	}
	return h
}

type nameDate struct {
	Name string
	Date string
}

func exams(mods []module, dates map[int64]time.Time) []nameDate {
	var o []nameDate
	for _, m := range mods {
		for i, name := range m.ClassNames {
			if strings.Contains(strings.ToLower(name), "exam") {
				o = append(o, nameDate{
					Name: name,
					Date: classSession(m, dates, i).Format(dateFormat),
				})
			}
		}
	}
	return o
}
func homeworkAssigned(m module, dates map[int64]time.Time) time.Time {
	d := dates[m.ID()].Add(-7 * 24 * time.Hour)
	if d.Before(startDate) {
		return startDate
	}
	return d
}
func homeworkDeadline1(m module, dates map[int64]time.Time) time.Time {
	d := dates[m.ID()]
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextLecture(d)
	}
	return d
}
func homeworkDeadline2(m module, dates map[int64]time.Time) time.Time {
	d := nextFridayNight(nextLecture(nextModuleStart(m, dates[m.ID()])))
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextFridayNight(d)
	}
	return d
}
func homeworkDeadline3(m module, dates map[int64]time.Time) time.Time {
	return homeworkDeadline2(m, dates).Add(14 * 24 * time.Hour)
}
func assignmentDeadline(m module, dates map[int64]time.Time) time.Time {
	return nextSundayNight(dates[m.ID()].Add(time.Duration(m.ProjectAssignmentDays) * 24 * time.Hour))
}

func stringToLink(s string) string {
	return strings.Replace(strings.Replace(strings.Replace(strings.ToLower(s), " ", "-", -1), "(", "", -1), ")", "", -1)
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
		"ContactHours": func(m module) string {
			return fmt.Sprintf("%.1f", contactHours(m))
		},
		"DiscussionInitialDeadline": func(m module) string {
			return discussionInitialDeadline(m, dates).Format(dateFormat)
		},
		"DiscussionResponseDeadline": func(m module) string {
			return discussionResponseDeadline(m, dates).Format(dateFormat)
		},
		"PreclassAssigned": func(m module, n int) string {
			return preclassAssigned(m, dates, n).Format(dayFormat)
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
		"ClassSession": func(m module, n int) string {
			return classSession(m, dates, n).Format(dateFormat)
		},
		"ModuleLink": func(m module) string {
			return stringToLink(m.Title)
		},
		"StringLink": func(s string) string {
			return stringToLink(s)
		},
		"ClassTitle": func(m module, n int) string {
			return classTitle(m, n)
		},
		"ProjectTitle": func(p project) string { return projectTitle(p) },
		"HasHomework": func(m module) bool {
			return getHomework(m) != nil
		},
		"PLWebsite": func() string {
			return plWebsite
		},
	}

	for _, mod := range modules {
		setupPreclass(mod, dates)
		setupHomework(mod, dates)
		setupInClass(mod, dates)
	}
	proj := projects(modules)
	for _, p := range proj {
		setupProject(p)
	}

	tmpl := template.Must(template.New("root").Funcs(funcMap).ParseFiles("modules_template.md"))

	w, err := os.Create("04.modules.md")
	check(err)

	schedule := struct {
		MidtermExamStart, MidtermExamEnd string
		FinalExamStart, FinalExamEnd     string
		Modules                          []module
		Projects                         []project
		Exams                            []nameDate
	}{
		FinalExamStart: finalExamStart.Format(dateFormat),
		FinalExamEnd:   finalExamEnd.Format(dateFormat),
		Modules:        modules,
		Projects:       proj,
		Exams:          exams(modules, dates),
	}

	check(tmpl.ExecuteTemplate(w, "modules_template.md", schedule))
	w.Close()

	if *cal {
		createCalendar(modules, proj, dates, funcMap)
	}
}

func createCalendar(modules []module, proj []project, startDates map[int64]time.Time, funcs template.FuncMap) {
	b, err := ioutil.ReadFile("client_secret_28501454573-amktdv82kcnrosm55muahjr2rbmr9nkr.apps.googleusercontent.com.json")
	//b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	// export GOOGLE_APPLICATION_CREDENTIALS=$PWD/class-calendar-1598027004004-259ee97255bd.json
	//ctx := context.Background()
	//srv, err := calendar.NewService(ctx)
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
	for _, p := range proj {
		p.toCalendar(srv)
	}
	for d := startDate; d.Before(finalExamStart); d = nextTessumOfficeHour(d) {
		tessumOfficeHoursToCalendar(srv, d)
	}
	for d := startDate; d.Before(finalExamStart); d = nextGuoOfficeHour(d) {
		guoOfficeHoursToCalendar(srv, d)
	}
	for d := startDate; d.Before(finalExamStart); d = nextWangOfficeHour(d) {
		wangOfficeHoursToCalendar(srv, d)
	}

	for _, m := range modules {
		d := startDates[m.Number]
		fmt.Println("Adding events to calendar for module:", m.Number)
		m.lecturesAssignmentsMidtermsToCalendar(srv, d)
		m.discussionToCalendar(srv, startDates)
		m.homeworkToCalendar(srv, startDates)
		m.preclassToCalendar(srv, startDates)
		time.Sleep(5 * time.Second)
	}
	finalExamToCalendar(srv)
}

func (m module) lecturesAssignmentsMidtermsToCalendar(srv *calendar.Service, startDate time.Time) {
	d := startDate
	for i, class := range m.ClassNames {

		if strings.Contains(strings.ToLower(class), "exam") {
			_, err := srv.Events.Insert(calendarID, &calendar.Event{
				Summary:     class,
				Description: "On Prairielearn (https://www.prairielearn.org/pl/)",
				Status:      "confirmed",
				Start: &calendar.EventDateTime{
					DateTime: d.Format(time.RFC3339),
				},
				End: &calendar.EventDateTime{
					DateTime: d.Add(examDuration).Format(time.RFC3339),
				},
			}).Do()
			check(err)
		} else {

			_, err := srv.Events.Insert(calendarID, &calendar.Event{
				Summary:     fmt.Sprintf("Class meeting: %s", classTitle(m, i)),
				Location:    "Room 1017 CEE Hydrosystems Laboratory, 301 N Mathews Ave, Urbana, IL 61801",
				Description: "Also on Zoom (see Canvas site for link).",
				Status:      "confirmed",
				Start: &calendar.EventDateTime{
					DateTime: d.Format(time.RFC3339),
				},
				End: &calendar.EventDateTime{
					DateTime: d.Add(classDuration).Format(time.RFC3339),
				},
			}).Do()
			check(err)
		}

		d = nextLecture(d)

	}
}

func tessumOfficeHoursToCalendar(srv *calendar.Service, d time.Time) {
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Tessum office hours",
		Location:    "Room 1017 CEE Hydrosystems Laboratory, 301 N Mathews Ave, Urbana, IL 61801",
		Description: "Also on Zoom (see Canvas site for link).",
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: d.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: d.Add(60 * time.Minute).Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

func guoOfficeHoursToCalendar(srv *calendar.Service, d time.Time) {
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Guo office hours",
		Location:    "Zoom",
		Description: "See Canvas site for link.",
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: d.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: d.Add(60 * time.Minute).Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

func wangOfficeHoursToCalendar(srv *calendar.Service, d time.Time) {
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Wang office hours",
		Location:    "Zoom",
		Description: "See Canvas site for link.",
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: d.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: d.Add(60 * time.Minute).Format(time.RFC3339),
		},
	}).Do()
	check(err)
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

func (m module) preclassToCalendar(srv *calendar.Service, dates map[int64]time.Time) {
	if m.PLName == "" {
		return
	}
	j := 1
	for i, className := range m.ClassNames {
		assess, err := getInfoAssessment(m, i, "preclass")
		if err != nil {
			fmt.Println("no pre-class for ", className)
			continue
		}

		var number string
		if len(m.ClassNames) > 1 {
			number = fmt.Sprintf("%d.%d", m.Number, j)
			j++
		} else {
			number = fmt.Sprintf("%d", m.Number)
		}
		_, err = srv.Events.Insert(calendarID, &calendar.Event{
			Summary:     fmt.Sprintf("Pre-class %s assigned", number),
			Location:    plWebsite,
			Description: assess.Title,
			Status:      "confirmed",
			Start: &calendar.EventDateTime{
				Date: preclassAssigned(m, dates, i).Format("2006-01-02"),
			},
			End: &calendar.EventDateTime{
				Date: preclassAssigned(m, dates, i).Format("2006-01-02"),
			},
		}).Do()
		check(err)

		_, err = srv.Events.Insert(calendarID, &calendar.Event{
			Summary:     fmt.Sprintf("Pre-class %s deadline", number),
			Location:    plWebsite,
			Description: m.Title,
			Status:      "confirmed",
			Start: &calendar.EventDateTime{
				DateTime: classSession(m, dates, i).Add(-15 * time.Minute).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: classSession(m, dates, i).Format(time.RFC3339),
			},
		}).Do()
		check(err)
	}
}

func (m module) homeworkToCalendar(srv *calendar.Service, dates map[int64]time.Time) {
	if getHomework(m) == nil {
		return
	}
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("HW%d Assigned: %s", m.Number, m.Title),
		Location:    plWebsite,
		Description: m.Title,
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
		Summary:     fmt.Sprintf("110%% credit HW%d deadline", m.Number),
		Location:    plWebsite,
		Description: m.Title,
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: homeworkDeadline1(m, dates).Add(-15 * time.Minute).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: homeworkDeadline1(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("100%% credit HW%d deadline", m.Number),
		Location:    plWebsite,
		Description: m.Title,
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: homeworkDeadline2(m, dates).Add(-15 * time.Minute).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: homeworkDeadline2(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)

	_, err = srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     fmt.Sprintf("80%% credit HW%d deadline", m.Number),
		Location:    plWebsite,
		Description: m.Title,
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: homeworkDeadline3(m, dates).Add(-15 * time.Minute).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: homeworkDeadline3(m, dates).Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

func (p project) toCalendar(srv *calendar.Service) {
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:  fmt.Sprintf("Project deliverable %d due", p.Number),
		Location: plWebsite,
		Status:   "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: p.Due.Add(-1 * time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: p.Due.Format(time.RFC3339),
		},
	}).Do()
	check(err)
}

func finalExamToCalendar(srv *calendar.Service) {
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
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

func classTitle(m module, i int) string {
	assess, err := getInfoAssessment(m, i, "preclass")
	if err != nil {
		fmt.Println("no pre-class for ", m.ClassNames[i])
		return m.ClassNames[i]
	}
	return assess.Title
}

func projectTitle(p project) string {
	return getProject(p).Title
}

func getHomework(m module) *infoAssessment {
	modPath := filepath.Join(courseInstance, "assessments", m.PLName, "homework")
	f, err := os.Open(filepath.Join(modPath, "infoAssessment.json"))
	if err != nil {
		return nil
	}
	d := json.NewDecoder(f)
	assess := new(infoAssessment)
	check(d.Decode(assess))
	f.Close()
	return assess
}

func writeHomework(assess *infoAssessment, m module) {
	modPath := filepath.Join(courseInstance, "assessments", m.PLName, "homework")
	w, err := os.Create(filepath.Join(modPath, "infoAssessment.json"))
	check(err)
	check(err)
	b, err := json.MarshalIndent(assess, "", "  ")
	check(err)
	_, err = w.Write(b)
	check(err)
	w.Close()
}

func getProject(p project) *infoAssessment {
	modPath := filepath.Join(courseInstance, "assessments", p.ID)
	f, err := os.Open(filepath.Join(modPath, "infoAssessment.json"))
	if err != nil {
		return nil
	}
	d := json.NewDecoder(f)
	assess := new(infoAssessment)
	check(d.Decode(assess))
	f.Close()
	return assess
}

func writeProject(assess *infoAssessment, p project) {
	modPath := filepath.Join(courseInstance, "assessments", p.ID)
	w, err := os.Create(filepath.Join(modPath, "infoAssessment.json"))
	check(err)
	check(err)
	b, err := json.MarshalIndent(assess, "", "  ")
	check(err)
	_, err = w.Write(b)
	check(err)
	w.Close()
}

func getInfoAssessment(m module, i int, typ string) (*infoAssessment, error) {
	modPath := filepath.Join(courseInstance, "assessments", m.PLName, typ)
	f, err := os.Open(filepath.Join(modPath, m.ClassNames[i], "infoAssessment.json"))
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(f)
	assess := new(infoAssessment)
	check(d.Decode(assess))
	f.Close()
	return assess, nil
}

func writeInfoAssessment(assess *infoAssessment, m module, i int, typ string) {
	modPath := filepath.Join(courseInstance, "assessments", m.PLName, typ)
	//os.MkdirAll(filepath.Join(modPath, m.ClassNames[i]), 0755)
	w, err := os.Create(filepath.Join(modPath, m.ClassNames[i], "infoAssessment.json"))
	check(err)
	b, err := json.MarshalIndent(assess, "", "  ")
	check(err)
	_, err = w.Write(b)
	check(err)
	w.Close()
}

func setupPreclass(m module, dates map[int64]time.Time) {
	if m.PLName == "" {
		return
	}
	j := 1
	for i, className := range m.ClassNames {
		assess, err := getInfoAssessment(m, i, "preclass")
		if err != nil {
			fmt.Println("no pre-class for ", className)
			continue
		}
		assess.Type = "Homework"
		assess.Set = "Pre-class"

		if len(m.ClassNames) > 1 {
			assess.Number = fmt.Sprintf("%d.%d", m.Number, j)
			j++
		} else {
			assess.Number = fmt.Sprintf("%d", m.Number)
		}
		assess.AllowAccess = []allowAccess{
			{
				StartDate: preclassAssigned(m, dates, i).Add(-14 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
				EndDate:   preclassAssigned(m, dates, i).Format("2006-01-02T15:04:05"),
				Credit:    0,
				Active:    false,
			},
			{
				StartDate: preclassAssigned(m, dates, i).Format("2006-01-02T15:04:05"),
				EndDate:   classSession(m, dates, i).Format("2006-01-02T15:04:05"),
				Credit:    100,
				Active:    true,
			},
			{
				StartDate: classSession(m, dates, i).Format("2006-01-02T15:04:05"),
				EndDate:   finalExamEnd.Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
				Credit:    0,
				Active:    true,
			},
		}
		if len(assess.Zones) == 1 {
			assess.Zones[0].Title = assess.Title
		}
		for i := range assess.Zones {
			assess.Zones[i].GradeRateMinutes = 5
		}
		writeInfoAssessment(assess, m, i, "preclass")
	}
}

func setupInClass(m module, dates map[int64]time.Time) {
	if m.PLName == "" {
		return
	}
	j := 1
	for i, className := range m.ClassNames {
		assess, err := getInfoAssessment(m, i, "inclass")
		if err != nil {
			fmt.Println("no in-class for ", className)
			continue
		}
		assess.Type = "Homework"
		assess.Set = "Worksheet"

		assess.GroupWork = true
		assess.StudentGroupCreate = true
		assess.StudentGroupJoin = true
		assess.StudentGroupLeave = true
		assess.GroupMaxSize = 4
		assess.GroupMinSize = 3

		if len(m.ClassNames) > 1 {
			assess.Number = fmt.Sprintf("%d.%d", m.Number, j)
			j++
		} else {
			assess.Number = fmt.Sprintf("%d", m.Number)
		}
		assess.AllowAccess = []allowAccess{
			{
				StartDate: classSession(m, dates, i).Add(-14 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
				EndDate:   classSession(m, dates, i).Format("2006-01-02T15:04:05"),
				Credit:    0,
				Active:    false,
			},
			{
				StartDate: classSession(m, dates, i).Format("2006-01-02T15:04:05"),
				EndDate:   classSession(m, dates, i).Add(classDuration).Add(time.Hour).Format("2006-01-02T15:04:05"),
				Credit:    100,
				Active:    true,
			},
			{
				StartDate: classSession(m, dates, i).Add(classDuration).Add(time.Hour).Format("2006-01-02T15:04:05"),
				EndDate:   finalExamEnd.Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
				Credit:    0,
				Active:    true,
			},
		}
		for i := range assess.Zones {
			assess.Zones[i].GradeRateMinutes = 1
		}
		if len(assess.Zones) == 1 {
			assess.Zones[0].Title = assess.Title
		}
		writeInfoAssessment(assess, m, i, "inclass")
	}
}

func setupHomework(m module, dates map[int64]time.Time) {
	hw := getHomework(m)
	if hw == nil {
		return
	}
	hw.Title = m.Title
	hw.Number = fmt.Sprint(m.Number)

	hw.AllowAccess = []allowAccess{
		{
			StartDate: homeworkAssigned(m, dates).Add(-14 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
			EndDate:   homeworkAssigned(m, dates).Format("2006-01-02T15:04:05"),
			Credit:    0,
			Active:    false,
		},
		{
			StartDate: homeworkAssigned(m, dates).Format("2006-01-02T15:04:05"),
			EndDate:   homeworkDeadline1(m, dates).Format("2006-01-02T15:04:05"),
			Credit:    110,
			Active:    true,
		},
		{
			StartDate: homeworkDeadline1(m, dates).Format("2006-01-02T15:04:05"),
			EndDate:   homeworkDeadline2(m, dates).Format("2006-01-02T15:04:05"),
			Credit:    100,
			Active:    true,
		},
		{
			StartDate: homeworkDeadline2(m, dates).Format("2006-01-02T15:04:05"),
			EndDate:   homeworkDeadline3(m, dates).Format("2006-01-02T15:04:05"),
			Credit:    80,
			Active:    true,
		},
		{
			StartDate: homeworkDeadline3(m, dates).Format("2006-01-02T15:04:05"),
			EndDate:   finalExamEnd.Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
			Credit:    0,
			Active:    true,
		},
	}
	writeHomework(hw, m)
}

func setupProject(p project) {
	assess := getProject(p)
	assess.Type = "Homework"
	assess.Set = "Project"
	assess.Number = fmt.Sprintf("%d", p.Number)

	assess.GroupWork = true
	assess.StudentGroupCreate = true
	assess.StudentGroupJoin = true
	assess.StudentGroupLeave = true
	assess.GroupMaxSize = 4
	assess.GroupMinSize = 3

	assess.AllowAccess = []allowAccess{
		{
			StartDate: p.Assigned.Add(-14 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
			EndDate:   p.Assigned.Format("2006-01-02T15:04:05"),
			Credit:    0,
			Active:    false,
		},
		{
			StartDate: p.Assigned.Format("2006-01-02T15:04:05"),
			EndDate:   p.Due.Format("2006-01-02T15:04:05"),
			Credit:    100,
			Active:    true,
		},
		{
			StartDate: p.Due.Format("2006-01-02T15:04:05"),
			EndDate:   finalExamEnd.Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
			Credit:    0,
			Active:    true,
		},
	}
	if len(assess.Zones) == 1 {
		assess.Zones[0].Title = assess.Title
	}
	writeProject(assess, p)
}

type infoAssessment struct {
	UUID               string        `json:"uuid"`
	Type               string        `json:"type"`
	Title              string        `json:"title"`
	Set                string        `json:"set"`
	Number             string        `json:"number"`
	GroupWork          bool          `json:"groupWork"`
	GroupMaxSize       int           `json:"groupMaxSize"`
	GroupMinSize       int           `json:"groupMinSize"`
	StudentGroupCreate bool          `json:"studentGroupCreate"`
	StudentGroupJoin   bool          `json:"studentGroupJoin"`
	StudentGroupLeave  bool          `json:"studentGroupLeave"`
	AllowAccess        []allowAccess `json:"allowAccess"`
	Zones              []zone        `json:"zones"`
}

type allowAccess struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Credit    int    `json:"credit"`
	Active    bool   `json:"active"`
}

type zone struct {
	Title            string     `json:"title"`
	GradeRateMinutes int        `json:"gradeRateMinutes"`
	Questions        []question `json:"questions"`
}

type question struct {
	ID     string `json:"id"`
	Points int    `json:"points"`
}
