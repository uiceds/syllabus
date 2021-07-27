## Schedule

[Course calendar](https://calendar.google.com/calendar/embed?src=c_fqvrphqptlccpp6pubokjsraj0%40group.calendar.google.com)

<iframe src="https://calendar.google.com/calendar/embed?src=c_fqvrphqptlccpp6pubokjsraj0%40group.calendar.google.com" style="border: 0" width="700" height="550" frameborder="0" scrolling="no"></iframe>

### Modules

| Module | Title | Start Date |
| -- | -- | -- |
{{- range .Modules}}
| {{.Number}} | [{{.Title}}](#module-{{.Number}}-{{ModuleLink .}}) | {{StartDate .}}
{{- end}}

### Discussions

| Title | Assigned | Initial Post Due | Response Posts Due |
| -- | -- | -- | -- |
{{- range .Modules}}
{{- if .DiscussionURL}}
| [{{.Title}}](#module-{{.Number}}-discussion) | {{DiscussionAssigned .}} | {{DiscussionInitialDeadline .}} | {{DiscussionResponseDeadline .}}
{{- end -}}
{{- end}}

### Homeworks

Title | Assigned | Deadline for 110% Credit | Deadline for 100% Credit | Deadline for 80% Credit |
| -- | -- | -- | -- | -- |
{{- range .Modules}}
{{- if .HomeworkURL}}
| [{{.Title}}](#module-{{.Number}}-homework) | {{HomeworkAssigned .}} | {{HomeworkDeadline1 .}} | {{HomeworkDeadline2 .}} | {{HomeworkDeadline3 .}} |
{{- end -}}
{{- end}}

### Project Assignments

| Title | Assigned | Due |
| -- | -- | -- |
{{- range .Modules}}
{{- if .ProjectAssignment}}
| [{{.ProjectAssignment}}](#{{StringLink .ProjectAssignment}}) | {{StartDate .}} | {{AssignmentDeadline .}}
{{- end -}}
{{- end}}

### Exams

* Midterm Exam: {{.MidtermExamStart}}—{{.MidtermExamEnd}}
* Final Exam: {{.FinalExamStart}}—{{.FinalExamEnd}}


## Modules

{{range .Modules}}
### Module {{.Number}}: {{.Title}}

{{if .Overview -}}
#### Module {{.Number}} Overview
{{.Overview}}
{{- end}}

{{if .Objectives -}}
#### Module {{.Number}} Learning Objectives

By the end of this module, you should be able to:

{{if len .Objectives | eq 1}}{{index .Objectives 0}}
{{- else}}
{{range .Objectives}}* {{.}}
{{end -}}
{{end}}
{{- end}}

{{if .Readings -}}
#### Module {{.Number}} Readings and Lectures
{{if .DiscussionPrompts -}}
Develop your answers to the discussion questions below while completing the readings and lectures.
{{end}}
{{range .Readings}}* {{.}}
{{end}}
{{- end}}

{{if .DiscussionPrompts -}}
#### Module {{.Number}} Discussion

This module includes a discussion section to help you understand by articulating how the module content could be useful in your professional life.
Consider the following questions:

{{range .DiscussionPrompts}}  * {{.}}
{{end}}

Log in to the [module discussion forum]({{.DiscussionURL}}) and make one initial post and two responses.
Refer to the [Discussion Forum Instructions and Rubric](#discussion-forum-instructions-and-rubric) for instructions how to compose posts to the discussion forum, and how they will be graded.

**The initial post for this module are due by {{DiscussionInitialDeadline .}} and all response posts are due by {{DiscussionResponseDeadline .}}.**

{{- end}}

{{if .HomeworkURL -}}
#### Module {{.Number}} Homework
The homework for Module {{.Number}} covers the required readings and lectures and is available [here]({{.HomeworkURL}}).
General information about homework assignments is [here](#homeworks-and-exams).

**The homework for this module is due by {{HomeworkDeadline1 .}} for 110% credit, by {{HomeworkDeadline2 .}} for 100% credit, and by {{HomeworkDeadline3 .}} for 80% credit.**
{{- end}}

{{if .LiveMeetingTopics -}}
#### Module {{.Number}} Topics for Zoom Meetings

{{with $mod := .}}
{{range $index, $element := .LiveMeetingTopics}}* {{ClassSession $mod $index}}: {{$element}}
{{end}}
{{- end}}
{{end}}

{{if .ProjectAssignment}}
#### Module {{.Number}} Project Assignment
{{.ProjectAssignment}}
{{end}}

{{- end}}
