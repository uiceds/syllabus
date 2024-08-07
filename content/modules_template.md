## Schedule

[Course calendar](https://calendar.google.com/calendar/embed?src=c_fqvrphqptlccpp6pubokjsraj0%40group.calendar.google.com)

<iframe src="https://calendar.google.com/calendar/embed?src=c_fqvrphqptlccpp6pubokjsraj0%40group.calendar.google.com" style="border: 0" width="700" height="550" frameborder="0" scrolling="no"></iframe>

### Modules

| Module | Start Date | Contact Hours |
| ---------- | --- | -- |
{{- range .Modules}}
| {{.Number}}. [{{.Title}}](#module-{{.Number}}-{{ModuleLink .}}) | {{StartDate .}} | {{ContactHours .}} |
{{- end}}

<!-- ### Discussions

| Title | Assigned | Initial Post Due | Response Posts Due |
| -- | -- | -- | -- |
{{- range .Modules}}
{{- if .DiscussionURL}}
| [{{.Title}}](#module-{{.Number}}-discussion) | {{DiscussionAssigned .}} | {{DiscussionInitialDeadline .}} | {{DiscussionResponseDeadline .}}
{{- end -}}
{{- end}}
-->

### Homeworks

Title | Assigned | Deadline for 110% Credit | Deadline for 100% Credit | Deadline for 80% Credit |
| -- | -- | -- | -- | -- |
{{- range .Modules}}
{{- if HasHomework .}}
| HW{{.Number}}: [{{.Title}}]({{PLWebsite}}) | {{HomeworkAssigned .}} | {{HomeworkDeadline1 .}} | {{HomeworkDeadline2 .}} | {{HomeworkDeadline3 .}} |
{{- end -}}
{{- end}}

### Project Deliverables

| Title | Assigned | Due |
| -- | -- | -- |
{{- range .Projects}}
| {{.Number}}. [{{ProjectTitle .}}]({{PLWebsite}}) | {{.Assigned}} | {{.Due}}
{{- end}}

### Quizzes

Each quiz covers the content in the corresponding module, e.g. Quiz 1 covers Module 1.

* Quiz 1: 2024-09-18 – 2024-09-20
* Quiz 2: 2024-10-09 – 2024-10-11
* Quiz 3: 2024-10-16 – 2024-10-18
* Quiz 4: 2024-10-23 – 2024-10-25
* Quiz 5: 2024-11-06 – 2024-11-08
* Quiz 6: 2024-11-13 – 2024-11-15
* Quiz 7: 2024-11-20 – 2024-11-22
* Quiz 8: 2024-12-04 – 2024-12-06

### Mini-Projects
{{range .Exams}}
* {{.Name}}: {{.Date}}
{{end}}
<!-- * Final Exam: {{.FinalExamStart}}—{{.FinalExamEnd}}-->


## Modules

{{range .Modules}}
### Module {{.Number}}: {{.Title}}

{{if .Overview -}}
**Module {{.Number}} Overview:** {{.Overview}}
{{- end}}

{{if .Objectives -}}
**Module {{.Number}} Learning Objectives:** By the end of this module, you should be able to:

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

{{if .ClassNames -}}
**Module {{.Number}} Class sessions:**

{{with $mod := .}}
{{range $index, $element := .ClassNames}}* {{ClassSession $mod $index}}: {{ClassTitle $mod $index}}
{{end}}
{{- end}}
{{end}}

<!--{{if .ProjectAssignment}}
#### Module {{.Number}} Project Assignment
{{.ProjectAssignment}}
{{end}}-->

{{- end}}
