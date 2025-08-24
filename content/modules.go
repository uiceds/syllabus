package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/context"
	calendar "google.golang.org/api/calendar/v3"
)

var cal = flag.Bool("cal", false, "Whether to create calendar events")
var plPath = flag.String("pl-path", "../../pl-cee498ds/courseInstances/Fa2025", "Path to PrairieLearn course repository")

var courseInstance string

const calendarID = "c_fqvrphqptlccpp6pubokjsraj0@group.calendar.google.com"

const plWebsite = "https://us.prairielearn.com/pl/course_instance/191683"

var startDate, finalExamStart, finalExamEnd time.Time

var classDuration = 80 * time.Minute
var examDuration = 24 * 5 * time.Hour

var loc *time.Location

func init() {
	flag.Parse()
	courseInstance = *plPath

	var err error
	loc, err = time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}
	startDate = time.Date(2025, time.August, 26, 12, 0, 0, 0, loc)

	finalExamStart = time.Date(2025, time.December, 13, 8, 00, 0, 0, loc)
	finalExamEnd = time.Date(2025, time.December, 14, 8, 00, 0, 0, loc)
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
		// if strings.Contains(m.ProjectAssignment, "exploratory") {
		// 	proj = append(proj, project{
		// 		ID:       m.ProjectAssignment,
		// 		Number:   i + 1,
		// 		Assigned: assigned,
		// 		Due:      nextFridayNight(projectAssignmentDue(m, startDates(modules))), // Avoid conflict with exam. TODO: Fix
		// 	})
		// } else {
		proj = append(proj, project{
			ID:       m.ProjectAssignment,
			Number:   i + 1,
			Assigned: assigned,
			Due:      projectAssignmentDue(m, startDates(modules)),
		})
		// }
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

	// URLs to videos for post-worksheet reviews.
	ClassVideos []string

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
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_4zfiknz9/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_hhfatagk/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_9k17i0n4/uiConfId/26883701/st/0",
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
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_ii4qwjk8/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_iwknq0p5/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_ummczmvd/uiConfId/26883701/st/0",
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
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_wc3t3uvo/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_8pwyed7l/uiConfId/26883701/st/0",
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
			"fft",
			"Culmination Project 1: Computational thinking",
			"wavelet",
		},
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_0humutkz/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_e94r5ina/uiConfId/26883701/st/0",
			"",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_fo7d1mac/uiConfId/26883701/st/0",
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
		},
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_0kpib7fh/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_vm8s1u90/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_1j1pruhy/uiConfId/26883701/st/0",
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
			"Culmination Project 2: Coordinate transforms",
			"classification_trees",
		},
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_eklcjbu9/uiConfId/26883701/st/0",
			"",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_wr4mjts9/uiConfId/26883701/st/0",
		},
		ProjectAssignment: "project/exploratory",
	},
	{
		Number:   7,
		Parents:  []int64{6},
		Title:    "Neural networks",
		Overview: `In this module, we will learn how to implement and use fully-connected neural networks.`,
		Objectives: []string{
			"Train a neural network to for regression and classification",
			"Identify and debug common problems with neural network training",
		},
		PLName: "neural_nets",
		ClassNames: []string{
			"neural_nets1",
			"neural_nets2",
		},
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_d4c85x9s/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_fjoia5fi/uiConfId/26883701/st/0",
		},
	},
	{
		Number:   8,
		Parents:  []int64{7},
		Title:    "Convolutional neural networks",
		Overview: `In this module, we will learn how to implement and use convolutional neural networks.`,
		Objectives: []string{
			"Train a convolutional neural network to for regression and classification",
			"Identify and debug common problems with convolutional neural network training",
		},
		PLName: "conv_nets",
		ClassNames: []string{
			"conv_nets",
			"conv_nets2",
		},
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_w6xzu3ef/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_9auzitjw/uiConfId/26883701/st/0",
		},
	},
	{
		Number:   9,
		Parents:  []int64{8},
		Title:    "Data-driven dynamical systems",
		PLName:   "data_driven_dynamics",
		Overview: `In this module, we will apply the machine learning techniques we have learned so far to dynamical systems and the differential equations that describe them.`,
		Objectives: []string{
			"Apply gradient descent to fit the parameters of a system of differential equations to observed data",
			"Implement a Neural ODE to make data-driven predictions of the evolution of a dynamical system",
		},
		ClassNames: []string{
			"param_fitting",
			"neural_odes",
		},
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_va17wzzh/uiConfId/26883701/st/0",
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_xtn8r0a7/uiConfId/26883701/st/0",
		},
		ProjectAssignment: "project/modeling",
	},
	{
		Number:  10,
		Parents: []int64{9},
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
			"Fall break",
			"Fall break",
			"Culmination Project 3: Machine Learning",
		},
		ClassVideos: []string{
			"https://mediaspace.illinois.edu/embed/secure/iframe/entryId/1_cxewm7yq/uiConfId/26883701/st/0",
		},
		ProjectAssignment: "project/rough_draft",
	},
	{
		Number:   11,
		Parents:  []int64{10},
		Title:    "Final projects",
		Overview: `This week we will not be meeting as a class to give time to finish your remaining coursework.`,
		ClassNames: []string{
			"No class",
			"No class",
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
			return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, d.Location()).Add(5 * time.Hour)
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
func nextTAOfficeHour(t time.Time) time.Time {
	d := t
	for {
		d = d.Add(24 * time.Hour)
		if w := d.Weekday(); w == time.Friday {
			return time.Date(d.Year(), d.Month(), d.Day(), 15, 0, 0, 0, d.Location())
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
	// if d.Before(startDate) {
	// 	return startDate
	// }
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
			if strings.Contains(strings.ToLower(name), "culmination project") {
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
	// if d.Before(startDate) {
	// 	return startDate
	// }
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
	d := nextSundayNight(nextLecture(nextModuleStart(m, dates[m.ID()])))
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextSundayNight(d)
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
		setupPostClass(mod, dates)
		setupExtraCredit(mod, dates)
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
	// export GOOGLE_APPLICATION_CREDENTIALS=class-calendar-1598027004004-ee3067aa11ec.json
	// Make sure service account has edit access to the calendar, which is owned by ctessum@illinois.edu.
	srv, err := calendar.NewService(context.Background())
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
	for d := startDate; d.Before(finalExamStart); d = nextTAOfficeHour(d) {
		taOfficeHoursToCalendar(srv, d)
	}

	for _, m := range modules {
		fmt.Println("Adding events to calendar for module:", m.Number)
		m.lecturesAssignmentsMidtermsToCalendar(srv, startDates)
		m.discussionToCalendar(srv, startDates)
		m.homeworkToCalendar(srv, startDates)
		m.preclassToCalendar(srv, startDates)
		time.Sleep(5 * time.Second)
	}
	//finalExamToCalendar(srv)
}

func (m module) lecturesAssignmentsMidtermsToCalendar(srv *calendar.Service, dates map[int64]time.Time) {
	for i, class := range m.ClassNames {
		d := classSession(m, dates, i)
		if strings.Contains(strings.ToLower(class), "culmination project") {
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
				Location:    "Room 3019 CEE Hydrosystems Laboratory, 301 N Mathews Ave, Urbana, IL 61801",
				Description: "https://www.prairielearn.org/pl/",
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
	}
}

func tessumOfficeHoursToCalendar(srv *calendar.Service, d time.Time) {
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "Tessum office hours",
		Location:    "Common area in the Smart Bridge, CEE Hydrosystems Laboratory, 301 N Mathews Ave, Urbana, IL 61801",
		Description: "",
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

func taOfficeHoursToCalendar(srv *calendar.Service, d time.Time) {
	_, err := srv.Events.Insert(calendarID, &calendar.Event{
		Summary:     "TA office hours",
		Location:    "Common area in the Smart Bridge, 301 N Mathews Ave, Urbana, IL 61801",
		Description: "",
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
	b, err := json.MarshalIndent(assess, "", "  ")
	check(err)
	_, err = w.Write(b)
	check(err)
	w.Close()
}

func getExtraCredit(m module) *infoAssessment {
	modPath := filepath.Join(courseInstance, "assessments", "extra_credit", m.PLName)
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

func writeExtraCredit(assess *infoAssessment, m module) {
	modPath := filepath.Join(courseInstance, "assessments", "extra_credit", m.PLName)
	w, err := os.Create(filepath.Join(modPath, "infoAssessment.json"))
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
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(assess)
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
		assess.Module = m.PLName

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
		assess.GroupRoles = []grouprole{}
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
		assess.Module = m.PLName

		assess.GroupWork = true
		assess.StudentGroupCreate = true
		assess.StudentGroupJoin = true
		assess.StudentGroupLeave = true
		assess.GroupMaxSize = 4
		assess.GroupMinSize = 3

		assess.GroupRoles = []grouprole{
			{
				Name:           "Manager",
				Minimum:        1,
				Maximum:        1,
				CanAssignRoles: true,
			},
			{
				Name:    "Recorder",
				Minimum: 1,
				Maximum: 1,
			},
			{
				Name:    "Reflector",
				Minimum: 1,
				Maximum: 1,
			},
			{
				Name:    "Spokesperson",
				Minimum: 1,
				Maximum: 1,
			},
		}

		if len(m.ClassNames) > 1 {
			assess.Number = fmt.Sprintf("%d.%d", m.Number, j)
			j++
		} else {
			assess.Number = fmt.Sprintf("%d", m.Number)
		}
		assess.AllowAccess = []allowAccess{
			{
				StartDate: classSession(m, dates, i).Add(-14 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
				EndDate:   classSession(m, dates, i).Add(-5 * time.Minute).Format("2006-01-02T15:04:05"),
				Credit:    0,
				Active:    false,
			},
			{
				StartDate: classSession(m, dates, i).Add(-5 * time.Minute).Format("2006-01-02T15:04:05"),
				EndDate:   classSession(m, dates, i).Add(3 * time.Hour).Format("2006-01-02T15:04:05"),
				Credit:    100,
				Active:    true,
			},
			{
				StartDate: classSession(m, dates, i).Add(3 * time.Hour).Format("2006-01-02T15:04:05"),
				EndDate:   finalExamEnd.Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
				Credit:    0,
				Active:    true,
			},
		}
		if len(assess.Zones) == 1 {
			assess.Zones[0].Title = assess.Title
		}
		writeInfoAssessment(assess, m, i, "inclass")
	}
}

func setupPostClass(m module, dates map[int64]time.Time) {
	if m.PLName == "" {
		return
	}
	j := 1
	for i, className := range m.ClassNames {
		assess, err := getInfoAssessment(m, i, "postclass")
		if err != nil {
			fmt.Println("no in-class for ", className)
			continue
		}
		assess.Type = "Homework"
		assess.Set = "Post-class"
		assess.Module = m.PLName
		if i < len(m.ClassVideos) {
			assess.Text = `<p>Try the worksheet again, this time following along with this lecture video. Points you get on this worksheet can ` +
				`partially count toward your grade on the in-class version.</p><div class="embed-responsive embed-responsive-16by9">` +
				`<iframe id="kmsembed-1_4zfiknz9" width="720" height="439" src="` +
				m.ClassVideos[i] +
				`" class="kmsembed" allowfullscreen webkitallowfullscreen mozAllowFullScreen allow="autoplay *; fullscreen *; encrypted-media *" ` +
				`referrerPolicy="no-referrer-when-downgrade" sandbox="allow-forms allow-same-origin allow-scripts allow-top-navigation ` +
				`allow-pointer-lock allow-popups allow-modals allow-orientation-lock allow-popups-to-escape-sandbox allow-presentation ` +
				`allow-top-navigation-by-user-activation" frameborder="0" title="CEE 492: Worksheet ` +
				fmt.Sprintf("%d.%d", m.Number, j) +
				`"></iframe></div>"`
		} else {
			fmt.Println("no video for ", className)
			continue
		}

		assess.GroupWork = false
		assess.GroupRoles = []grouprole{}

		if len(m.ClassNames) > 1 {
			assess.Number = fmt.Sprintf("%d.%d", m.Number, j)
			j++
		} else {
			assess.Number = fmt.Sprintf("%d", m.Number)
		}
		closeDate := classSession(m, dates, i).Add(classDuration).Add(time.Hour * 24 * 21)
		if finalExamEnd.Before(closeDate) {
			closeDate = finalExamEnd
		}
		assess.AllowAccess = []allowAccess{
			{
				StartDate: classSession(m, dates, i).Add(-14 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
				EndDate:   classSession(m, dates, i).Add(3 * time.Hour).Format("2006-01-02T15:04:05"),
				Credit:    0,
				Active:    false,
			},
			{
				StartDate: classSession(m, dates, i).Add(3 * time.Hour).Format("2006-01-02T15:04:05"),
				EndDate:   closeDate.Format("2006-01-02T15:04:05"),
				Credit:    100,
				Active:    true,
			},
			{
				StartDate: closeDate.Format("2006-01-02T15:04:05"),
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
		writeInfoAssessment(assess, m, i, "postclass")
	}
}

func setupHomework(m module, dates map[int64]time.Time) {
	hw := getHomework(m)
	if hw == nil {
		return
	}
	hw.Title = m.Title
	hw.Number = fmt.Sprint(m.Number)
	hw.Module = m.PLName

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
	hw.GroupRoles = []grouprole{}
	writeHomework(hw, m)
}

func setupExtraCredit(m module, dates map[int64]time.Time) {
	xc := getExtraCredit(m)
	if xc == nil {
		return
	}
	xc.Title = m.Title
	xc.Number = fmt.Sprint(m.Number)
	xc.Module = m.PLName

	start := moduleStart(m, dates)
	end := nextModuleStart(m, start).Add(7 * 24 * time.Hour)
	if end.After(finalExamEnd) {
		end = finalExamEnd
	}

	xc.AllowAccess = []allowAccess{
		{
			StartDate: start.Add(-14 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
			EndDate:   start.Format("2006-01-02T15:04:05"),
			Credit:    0,
			Active:    false,
		},
		{
			StartDate: start.Format("2006-01-02T15:04:05"),
			EndDate:   end.Format("2006-01-02T15:04:05"),
			Credit:    100,
			Active:    true,
		},
		{
			StartDate: end.Format("2006-01-02T15:04:05"),
			EndDate:   finalExamEnd.Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05"),
			Credit:    0,
			Active:    true,
		},
	}
	xc.GroupRoles = []grouprole{}
	writeExtraCredit(xc, m)
}

func setupProject(p project) {
	assess := getProject(p)
	assess.Type = "Homework"
	assess.Set = "Project"
	assess.Module = "project"
	assess.Number = fmt.Sprintf("%d", p.Number)

	assess.GroupWork = true
	assess.StudentGroupCreate = true
	assess.StudentGroupJoin = true
	assess.StudentGroupLeave = true
	assess.GroupMaxSize = 4
	assess.GroupMinSize = 3
	assess.GroupRoles = []grouprole{}

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
	Module             string        `json:"module"`
	Set                string        `json:"set"`
	Number             string        `json:"number"`
	GroupWork          bool          `json:"groupWork"`
	GroupRoles         []grouprole   `json:"groupRoles,omitempty"`
	GroupMaxSize       int           `json:"groupMaxSize"`
	GroupMinSize       int           `json:"groupMinSize"`
	StudentGroupCreate bool          `json:"studentGroupCreate"`
	StudentGroupJoin   bool          `json:"studentGroupJoin"`
	StudentGroupLeave  bool          `json:"studentGroupLeave"`
	Text               string        `json:"text,omitempty"`
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
	GradeRateMinutes int        `json:"gradeRateMinutes,omitempty"`
	Questions        []question `json:"questions"`
	CanView          []string   `json:"canView,omitempty"`
	CanSubmit        []string   `json:"canSubmit,omitempty"`
}

type question struct {
	ID           string  `json:"id,omitempty"`
	Alternatives []alt   `json:"alternatives,omitempty"`
	Points       float64 `json:"points,omitempty"`
	ManualPoints float64 `json:"manualPoints,omitempty"`
	NumberChoose int     `json:"numberChoose,omitempty"`
}

type alt struct {
	ID string `json:"id"`
}

type grouprole struct {
	Name           string `json:"name"`
	Minimum        int    `json:"minimum"`
	Maximum        int    `json:"maximum"`
	CanAssignRoles bool   `json:"canAssignRoles"`
}
