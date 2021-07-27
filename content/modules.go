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
	startDate = time.Date(2021, time.August, 24, 12, 0, 0, 0, loc)

	midtermExamStart = time.Date(2020, time.October, 15, 12, 0, 0, 0, loc)
	midtermExamEnd = time.Date(2020, time.October, 16, 12, 0, 0, 0, loc)
	finalExamStart = time.Date(2020, time.December, 11, 13, 30, 0, 0, loc)
	finalExamEnd = time.Date(2020, time.December, 11, 16, 30, 0, 0, loc)
}

func startDates(modules []module) map[int64]time.Time {
	startDates := make(map[int64]time.Time)
	mods := make(map[int64]module)
	for _, m := range modules {
		mods[m.Number] = m
	}
	for _, m := range modules {
		if len(m.Parents) == 0 { // Root starts at the start date
			startDates[m.ID()] = startDate
		} else { // Start at date that latest parent ends.
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
	}
	return startDates
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

	PreClassQuestions []string
	ClassQuestions    []string
	HomeworkQuestions []string

	LiveMeetingTopics     []string
	ProjectAssignment     string
	ProjectAssignmentDays int
}

func (m module) ID() int64 { return m.Number }

var modules = []module{
	{
		Number:   0,
		Title:    "Introduction and motivating problems",
		Overview: "In this module we will get to know each other and cover the format of the course, its contents, and expectations.",
		LiveMeetingTopics: []string{
			"Introduction, syllabus, and getting to know one another ([slides](https://docs.google.com/presentation/d/1U-xdr_lPprNl5HiMTZ-pG0p9lRhR9FR0zMF3pwzVPjw/edit?usp=sharing))",
		},
	},
	{
		Number:  1,
		Parents: []int64{0},
		Title:   "Linear algebra review and intro to the Julia Language",
		Overview: `In this course, we will use two key tools: linear algebra and the Julia programming language. 
		You should already be familiar with linear algebra, so we will only briefly review it here. 
		You're not expected to know anything about the Julia language before starting this class, but you are
		expected to have completed a basic computer programming class (similar to CS101) using some computing language.`,
		PreClassQuestions: []string{
			"linalg_review",
		},
		LiveMeetingTopics: []string{
			"The Julia language: variables, strings, and data structures",
			"The Julia language: loops, conditionals, and functions",
		},
	},
	{
		Number:   2,
		Parents:  []int64{1},
		Title:    "Open reproducible science",
		Overview: "This module covers tools and methods for ensuring your work is correct, understandable, and reproducible. ",
		LiveMeetingTopics: []string{
			"Git & GitHub for reproduceable science",
			"Data cleaning & visualization",
			"Exploratory data analysis",
		},
	},
	{
		Number:  3,
		Parents: []int64{2},
		Title:   "Singular value decomposition and principle component analysis",
		LiveMeetingTopics: []string{
			"Singular value decomposition: theory",
			"Singular value decomposition: uses",
		},
	},
	{
		Number:  4,
		Parents: []int64{3},
		Title:   "Fourier and wavelet transforms",
		LiveMeetingTopics: []string{
			"Fourier transforms: theory",
			"Fourier transforms: applications",
		},
	},
	{
		Number:  5,
		Parents: []int64{4},
		Title:   "Regression",
		LiveMeetingTopics: []string{
			"Classic Curve Fitting and Least-Squares Regression",
			"Nonlinear Regression and Gradient Descent",
			"Over- and Under-determined Systems (Also: Project check-in)",
		},
	},
	{
		Number:  6,
		Parents: []int64{5},
		Title:   "Regularization and model fit 1",
		LiveMeetingTopics: []string{
			"Optimization for Regressions",
			"The Pareto Front and Parsimonious Models",
		},
	},
	{
		Number:  7,
		Parents: []int64{6},
		Title:   "Regularization and model fit 2",
		LiveMeetingTopics: []string{
			"Model Selection and Cross-Validation",
			"Feature Selection and Data Mining",
		},
	},
	{
		Number:  8,
		Parents: []int64{7},
		Title:   "Machine learning",
		LiveMeetingTopics: []string{
			"Supervised versus Unsupervised Learning",
			"Unsupervised Learning - k-Means Clustering",
			"Supervised Learning - Classification Trees",
		},
	},
	{
		Number:  9,
		Parents: []int64{8},
		Title:   "Neural networks 1",
		LiveMeetingTopics: []string{
			"Basics of Neural Networks",
			"Neural networks and activation functions",
		},
	},
	{
		Number:  10,
		Parents: []int64{9},
		Title:   "Neural networks 2",
		LiveMeetingTopics: []string{
			"The backpropagation algorithm",
			"The stochastic gradient descent algorithm",
		},
	},
	{
		Number:  11,
		Parents: []int64{10},
		Title:   "Data-driven dynamical systems",
		LiveMeetingTopics: []string{
			"Parameter fitting for dynamical systems",
			"Neural network parameterization of dynamical systems (Neural ODEs)",
		},
	},
	{
		Number:  -1,
		Parents: []int64{11},
		Title:   "Fall break",
		LiveMeetingTopics: []string{
			"Fall break",
			"Fall break",
		},
	},
	{
		Number:  12,
		Parents: []int64{-1},
		Title:   "Final projects",
		LiveMeetingTopics: []string{
			"Project workshop",
			"Final project presentations",
			"Final project presentations",
		},
	},

	/*	{
				Number:   1,
				NumDays:  7,
				Title:    "Open Reproducible Science",
				Overview: "This module covers tools and methods for ensuring your work is correct, understandable, and reproducible. ",
				Objectives: []string{
					"List the Bash commands for doing different computer operations",
					"Describe how to organize data files to improve transparency and reproducibility",
					"Define the syntax for the Markdown text formatting language",
					"List the Git and GitHub commands and operations for performing different operations",
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
					"[Dataset for defect detection in Concrete](https://zenodo.org/record/2620293#.X0Gbd2jYpaY)",
					"[Best Datasets for Machine Learning Data Science](https://medium.com/towards-artificial-intelligence/best-datasets-for-machine-learning-data-science-computer-vision-nlp-ai-c9541058cf4f)",
					"[VisualData Dataset Discovery](https://www.visualdata.io/discovery)",
				},
				DiscussionURL: "https://compass2g.illinois.edu/webapps/discussionboard/do/conference?action=list_forums&course_id=_52490_1&conf_id=260881&nav=discussion_board_entry",
				DiscussionPrompts: []string{
					"A big part of this course will be a semester-long project, where you will use a dataset to answer a question relevant to Civil or Environmental Engineering. What are some ideas you have for course projects, how are they related to Civil or Environmental Engineering, why do you think they would be useful, and what dataset would they be based on?",
				},
				LiveMeetingTopics: []string{
					"Project introduction, brainstorming, and proposals (slides)[https://docs.google.com/presentation/d/1MjCbv3tA5FBN5Pu2rU670q1lk-aeQ1qmYR2vg8ai7j8/edit?usp=sharing]",
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
					"Practice excercises and Q&A. [Practice excercises](https://nbviewer.jupyter.org/github/uiceds/syllabus/blob/master/modules/module3/module3_practice_problems.ipynb)",
					"Practice excercises and literature review intro. [Slides](https://docs.google.com/presentation/d/1Hfy4NOrFLEKLOMLBpgyv_hUIUpP82ohYDSDw9HIOThY/edit?usp=sharing)",
				},
				ProjectAssignment:     `Project literature review`,
				ProjectAssignmentDays: 17,
			},
			{
				Number:  4,
				NumDays: 7,
				Parents: []int64{3},
				Title:   "Data",
				Overview: `Data relevant to Civil and Environmental Engineering comes in a number of different formats, including:

		* Tabular data (csv, excel): Rows represent observations, columns represent properties
		* Raster/image data ([geotiff](https://en.wikipedia.org/wiki/GeoTIFF), [NetCDF](https://www.unidata.ucar.edu/software/netcdf/), png, jpg)
		* graph data ([dot](https://gephi.org/users/supported-graph-formats/graphviz-dot-format/), [gtfs](https://gtfs.org/))`,
				Objectives: []string{
					"Specify information as tabular, raster, and graph data",
					"Use software tools to load and manipulate tabular, raster, and graph data",
				},
				Readings: []string{
					"[tabular data](http://vita.had.co.nz/papers/tidy-data.pdf)",
					"[Pandas for tabular data](https://www.datacamp.com/community/tutorials/pandas-tutorial-dataframe-python)",
					"[graph data structure](https://www.youtube.com/watch?v=gXgEDyodOJU)",
					"[graph properties](https://www.youtube.com/watch?v=AfYqN3fGapc)",
					"[graph lecture notes](http://www.cs.cmu.edu/afs/cs/academic/class/15210-f13/www/lectures/lecture09.pdf)",
					"[image data](https://www.youtube.com/watch?v=UhDlL-tLT2U&list=PLuh62Q4Sv7BUf60vkjePfcOQc8sHxmnDX)",
					"[pillow image processing](https://pillow.readthedocs.io/en/stable/handbook/tutorial.html)",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment_instance/1700831",
				LiveMeetingTopics: []string{
					"Group exercises [slides](https://docs.google.com/presentation/d/1SlC4bTehrP4cFuaSv36IMovJU24BMFbyEJthm8eooio/edit?usp=sharing)",
					"[Guided exercise](https://www.kaggle.com/christophertessum/module-4-class-2-exercise-1-result) and [group exercise](https://www.kaggle.com/christophertessum/module-4-class-2-exercise-2-result)",
				},
			},
			{
				Number:   5,
				NumDays:  7,
				Parents:  []int64{3, 4},
				Title:    "Exploratory data analysis (EDA)",
				Overview: "The first step in a data science project is getting a feel for the dataset you are working with. This is called Exploratory Data Analysis (EDA).",
				Objectives: []string{
					"Calculate relevant statistical properties of an unfamiliar dataset",
					"Visualize an unfamiliar dataset and describe its relevant properties",
				},
				Readings: []string{
					"[Exploratory Data Analysis](https://www.youtube.com/watch?v=zHcQPKP6NpM)",
					"[Exploratory Data Analysis in Pandas](https://youtu.be/WNoQTNOME5g?t=480) watch time 8:00 to 20:30",
					"mlcourse.ai notebooks [1](https://mlcourse.ai/articles/topic1-exploratory-data-analysis-with-pandas/), [2.1](https://mlcourse.ai/articles/topic2-visual-data-analysis-in-python/) and [2.2](https://mlcourse.ai/articles/topic2-part2-seaborn-plotly/)",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment_instance/1757363",
				LiveMeetingTopics: []string{
					"Practice project EDA [Notebook](https://www.kaggle.com/christophertessum/module-5-class-1-airplanes-result)",
					"Image processing [Notebook](https://www.kaggle.com/christophertessum/module-5-class-2-ships-result)",
				},
				ProjectAssignment:     `Project exploratory data analysis`,
				ProjectAssignmentDays: 29,
			},
			{
				Number:  6,
				Parents: []int64{3, 5},
				NumDays: 7,
				Title:   "Network analysis",
				Overview: `In the previous module we learned how to do exploratory data analysis for tabular data.
		This week we will work on exploratory data analysis for graph (network) data.`,
				Objectives: []string{
					"Load and manipulate graph data using the `networkx` python library",
					"Perform basic statistical analysis of graph data",
				},
				Readings: []string{
					"[Game of Thrones: Network Analysis](https://www.kaggle.com/mmmarchetti/game-of-thrones-network-analysis)",
					"[Networkx introduction](https://networkx.github.io/documentation/stable/reference/introduction.html)",
					"Networkx [betweenness centrality](https://networkx.github.io/documentation/stable/reference/algorithms/generated/networkx.algorithms.centrality.betweenness_centrality.html#networkx.algorithms.centrality.betweenness_centrality), [degree centrality](https://networkx.github.io/documentation/stable/reference/algorithms/generated/networkx.algorithms.centrality.degree_centrality.html#networkx.algorithms.centrality.degree_centrality), and [shortest path](https://networkx.github.io/documentation/stable/reference/algorithms/shortest_paths.html)",
				},
				LiveMeetingTopics: []string{
					"Network analysis exercise ([Notebook](https://www.kaggle.com/christophertessum/module-6-class-1-airplanes-result))",
					"Midterm review",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessments",
			},
			{
				Number:  7,
				Parents: []int64{6},
				NumDays: 7,
				Title:   "Spatial data",
				Overview: `Many Civil and Environmental Engineering analyses—for example transportation networks or environmental data—have
		a spatial component. This week we will work with spatial data.`,
				Objectives: []string{
					"Analyze raster (gridded) data using the XArray library",
					"Analyze vector (point, line, or polygon) data using the GeoPandas library",
				},
				Readings: []string{
					"[Xarray tutorial](https://www.youtube.com/playlist?list=PLTJsu1ustEMbVgE6SivbF17XvWmb3hqoR) videos 1–8. If you want to follow along, the link to the dataset they use in the video is broken, but there is another copy [here](https://esgf.nci.org.au/thredds/fileServer/master/CMIP5/output1/CSIRO-BOM/ACCESS1-3/historical/mon/atmos/Amon/r1i1p1/v20120413/tas/tas_Amon_ACCESS1-3_historical_r1i1p1_185001-200512.nc)",
					"[Intro to Geopandas](https://www.youtube.com/playlist?list=PLewNEVDy7gq3DjrPDxGFLbHE4G2QWe8Qh) videos 1, 3–14 (video 2 is installation, which you can do with anaconda)",
					"(In these and all video lectures, you can adjust the playback speed by clicking 'Settings' in the lower-right corner of the playback window.)",
				},
				LiveMeetingTopics: []string{
					"Geopandas ([Notebook](https://www.kaggle.com/christophertessum/module-7-class-1-airplanes-result))",
					"Xarray ([Notebook](https://www.kaggle.com/christophertessum/module-7-class-2-ds4g-result))",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment_instance/1837345",
			},
			{
				Number:     8,
				Parents:    []int64{2, 7},
				NumDays:    7,
				Title:      "Midterm",
				Overview:   "",
				Objectives: []string{"Access, characterize, and visualize the data for their projects"},
				Readings:   []string{},
				LiveMeetingTopics: []string{
					"Midterm course eval, project questions, visualizations",
					"Midterm Exam and project support",
				},
				ProjectAssignment:     `Project Kaggle Competition`,
				ProjectAssignmentDays: 47,
			},
			{
				Number:   9,
				Parents:  []int64{5, 8},
				NumDays:  7,
				Title:    "Intro to machine learning",
				Overview: "This week we will cover the basics of machine learning, including gradient descent, generalization, representation, and regularization.",
				Objectives: []string{
					"apply fundamental machine learning concepts",
				},
				Readings: []string{
					"[Google machine learning crash course](https://developers.google.com/machine-learning/crash-course), starting with 'Introduction to ML' and ending after 'Regularization: Sparsity'",
				},
				LiveMeetingTopics: []string{
					"Linear regression ([Notebook](https://www.kaggle.com/christophertessum/module-9-class-1-airplanes-result))",
					"Linear regression part 2 ([Notebook](https://www.kaggle.com/christophertessum/module-9-class-2-airplanes-result))",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment_instance/1924211/",
			},
			{
				Number:   10,
				Parents:  []int64{9},
				NumDays:  7,
				Title:    "Neural Networks",
				Overview: "",
				Objectives: []string{
					"Train a neural network to make predictions about CEE datasets",
					"Identify and debug common problems with neural network training",
				},
				Readings: []string{
					"[Google machine learning crash course](https://developers.google.com/machine-learning/crash-course), starting with 'Logistic Regression' and ending after 'Training Neural Nets'",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessments",
				LiveMeetingTopics: []string{
					"Neural Networks ([Notebook](https://www.kaggle.com/christophertessum/module-10-class-1-airplanes-result))",
					"Regularization ([Notebook](https://www.kaggle.com/christophertessum/module-10-class-2-regularization-result))",
				},
			},
			{
				Number:   11,
				Parents:  []int64{10},
				NumDays:  7,
				Title:    "Convolutional Neural Networks",
				Overview: "This week we will learn about convolutional neural networks for computer vision. Training these neural networks from scratch is computationally- and data-intensive, but good results can be achieved with less time and data using a technique called transfer learning.",
				Objectives: []string{
					"Apply transfer learning to solve computer vision problems",
				},
				Readings: []string{
					"[Convolutional neural network explainer](https://poloclub.github.io/cnn-explainer/)",
					"[Transfer learning video](https://youtu.be/yofjFQddwHE)",
					"[Transfer learning notebook](https://keras.io/guides/transfer_learning/)",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment_instance/1993504",
				LiveMeetingTopics: []string{
					"Remember to vote! ([class canceled](https://education.illinois.edu/about/news-events/news/article/2020/07/15/statewide-election-day-holiday-no-classes-on-tuesday,-november-3,-2020))",
					"Convolutional Neural Networks ([Notebook](https://www.kaggle.com/christophertessum/homework-11-result))",
				},
				ProjectAssignment:     `Project Group Final Presentation`,
				ProjectAssignmentDays: 24,
			},
			{
				Number:   12,
				Parents:  []int64{10, 11},
				NumDays:  7,
				Title:    "Neural Networks for Sequences",
				Overview: "Time series data are common in civil and environmental engineering, and machine learning can be used to make future predictions, for example for weather and pollution forecasts and predictions of mechanical failure.",
				Objectives: []string{
					"Create recurrent neural networks for time series predictions",
					"Apply recurrent neural networks to answer questions related to CEE",
				},
				Readings: []string{
					"[An illustrated guide to recurrent neural networks](https://youtu.be/LHXXI4-IEns)",
					"[Illustrated guide to LSTM's and GRU's: A step by step explanation](https://www.youtube.com/watch?v=8HyCNIVRbSU)",
					"[Understanding LSTM networks](https://colah.github.io/posts/2015-08-Understanding-LSTMs/)",
					"[Time series forecasting tutorial](https://www.tensorflow.org/tutorials/structured_data/time_series)",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment_instance/2040526",
				LiveMeetingTopics: []string{
					"LSTM for damage detection ([Notebook](https://www.kaggle.com/christophertessum/module-12-class-1-jet-engine-damage-result))",
					"LSTM for damage detection ([Notebook](https://www.kaggle.com/christophertessum/module-12-class-1-jet-engine-damage-result))",
				},
				ProjectAssignment:     `Project Group Final Report`,
				ProjectAssignmentDays: 25,
			},
			{
				Number:   13,
				Parents:  []int64{5, 12},
				NumDays:  7,
				Title:    "Random forests",
				Overview: "In addition to linear regression and neural networks, random forest models are another commonly used machine learning framework",
				Objectives: []string{
					"Apply random forest models to answer questions related to CEE",
				},
				Readings: []string{
					"[Visual intro to decision trees part 1](http://www.r2d3.us/visual-intro-to-machine-learning-part-1/)",
					"[Visual intro to decision trees part 2](http://www.r2d3.us/visual-intro-to-machine-learning-part-2/)",
					"[Decision trees](https://youtu.be/7VeUPuFGJHk)",
					"[Regression trees](https://youtu.be/g9c66TUylZ4)",
					"[Pruning regression trees](https://youtu.be/D0efHEJsfHo)",
					"[Random forests part 1](https://youtu.be/J4Wdy0Wc_xQ)",
					"[Random forests part 2](https://youtu.be/sQ870aTKqiM)",
				},
				HomeworkURL: "https://prairielearn.engr.illinois.edu/pl/course_instance/89830/assessment_instance/2065426",
				LiveMeetingTopics: []string{
					"Random forests ([Notebook](https://www.kaggle.com/christophertessum/module-13-class-1-result))",
					"",
				},
			},
			{
				Number:     14,
				Parents:    []int64{2, 13},
				NumDays:    7,
				Title:      "Fall Break",
				Overview:   "Happy Thanksgiving",
				Objectives: []string{},
				Readings: []string{
					"Optional: [Machine learning for fluid dynamics](https://www.youtube.com/watch?v=8e3OT2K99Kw)",
					"Optional: [Symmetry and Equivariance in neural networks](https://youtu.be/8s0Ka6Y_kIM)",
				},
				HomeworkURL: "",
				LiveMeetingTopics: []string{
					"No class",
					"No class",
				},
			},
			{
				Number:      15,
				Parents:     []int64{2, 14},
				NumDays:     9,
				Title:       "Final project presentations",
				Overview:    "",
				Objectives:  []string{"Students should have completed a project where they access and explore a civil or environmental dataset and use it to answer a scientific question."},
				Readings:    []string{},
				HomeworkURL: "",
				LiveMeetingTopics: []string{
					"Project presentations",
					"Project presentations",
					"Project presentations",
				},
			},*/
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
func nextModuleStart(m module, startDate time.Time) time.Time {
	d := startDate
	for i := 0; i < len(m.LiveMeetingTopics); i++ {
		d = nextLecture(d)
	}
	return d
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
func classSession(m module, dates map[int64]time.Time, num int) time.Time {
	d := nextLecture(dates[m.ID()])
	for i := 0; i < num; i++ {
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
	d := nextSundayNight(nextLecture(dates[m.ID()]))
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextSundayNight(d)
	}
	return d
}
func homeworkDeadline3(m module, dates map[int64]time.Time) time.Time {
	d := nextSundayNight(dates[m.ID()].Add(14 * 24 * time.Hour))
	for i := 0; i < m.HomeworkDelay; i++ {
		d = nextSundayNight(d)
	}
	return d
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
		"ClassSession": func(m module, n int) string {
			return classSession(m, dates, n).Format(dateFormat)
		},
		"ModuleLink": func(m module) string {
			return stringToLink(m.Title)
		},
		"StringLink": func(s string) string {
			return stringToLink(s)
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
	for i := 0; i < len(m.LiveMeetingTopics); i++ {
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
		d = nextLecture(d)
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
		Description: fmt.Sprintf("<a href=https://uiceds.github.io/syllabus/#%s>%s</a>", stringToLink(m.ProjectAssignment), m.ProjectAssignment),
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
		Description: fmt.Sprintf("<a href=https://uiceds.github.io/syllabus/#%s>%s</a>", stringToLink(m.ProjectAssignment), m.ProjectAssignment),
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
