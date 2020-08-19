## Modules

{{range .}}
### Module {{.Number}}: {{.Title}}

{{if .Overview -}}
#### Overview
{{.Overview}}
{{- end}}

{{if .Objectives -}}
#### Objectives
{{if len .Objectives | eq 1}}{{index .Objectives 0}}
{{- else}}{{range .Objectives}}  * {{.}}
{{end -}}
{{end}}
{{- end}}

{{if .Readings -}}
#### Readings and Lectures
{{if .DiscussionPrompts -}}
Develop your answers to the discussion questions below while completing the readings and lectures.
{{- end}}
{{range .Readings}}* {{.}}
{{end}}
{{- end}}

{{if .DiscussionPrompts -}}
#### Discussion

This module includes a discussion section to help you understand by articulating how the module content could be useful in your professional life.
Consider the following questions:

{{range .DiscussionPrompts}}  * {{.}}
{{end -}}

Log in to the [module discussion forum]({{.DiscussionURL}}) and make one initial post and two responses.
Refer to the [Discussion Forum Instructions and Rubric](discussion-forum-instructions-and-rubric) for instructions how to compose posts to the discussion forum, and how they will be graded.
**All posts for this module are due by {{.DiscussionDeadline}}.**


{{- end}}

{{if .QuizURL -}}
#### Quiz
The quiz for Module {{.Number}} covers the required readings and lectures and is available [here]({{.QuizURL}}).
**The quiz for this module is due by {{.QuizDeadline}}.**
{{- end}}

{{- end}}
