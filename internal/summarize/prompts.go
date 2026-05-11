package summarize

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type Kind string

const (
	Refinement   Kind = "refinement"
	Planning     Kind = "planning"
	Retro        Kind = "retro"
	OneOnOne      Kind = "1on1"
	Architecture  Kind = "architecture"
	TechInterview Kind = "tech-interview"
	Generic       Kind = "generic"
)

type Lang string

const (
	PtPT Lang = "pt-PT"
	PtBR Lang = "pt-BR"
	EnUS Lang = "en-US"
)

type BuildInput struct {
	Kind         Kind
	Lang         Lang
	Participants []string
	Transcript   string
}

var langInstruction = map[Lang]string{
	PtPT: "Responde em Português Europeu (Portugal). Usa léxico PT-PT (ex.: 'utilizador', 'ficheiro', 'a fazer'), nunca brasileiro.",
	PtBR: "Responda em Português Brasileiro. Use léxico PT-BR (ex.: 'usuário', 'arquivo', 'fazendo'), nunca europeu.",
	EnUS: "Respond in US English.",
}

type templateData struct {
	LangInstruction string
	Participants    string
	Transcript      string
}

func Build(in BuildInput) (string, error) {
	tmplStr, ok := templates[in.Kind]
	if !ok {
		return "", fmt.Errorf("unknown meeting kind: %q", in.Kind)
	}
	langInstr, ok := langInstruction[in.Lang]
	if !ok {
		return "", fmt.Errorf("unsupported language: %q", in.Lang)
	}
	tmpl, err := template.New(string(in.Kind)).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", in.Kind, err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData{
		LangInstruction: langInstr,
		Participants:    strings.Join(in.Participants, ", "),
		Transcript:      in.Transcript,
	})
	if err != nil {
		return "", fmt.Errorf("execute template %s: %w", in.Kind, err)
	}
	return buf.String(), nil
}

var templates = map[Kind]string{
	Refinement:    refinementTmpl,
	Planning:      planningTmpl,
	Retro:         retroTmpl,
	OneOnOne:      oneOnOneTmpl,
	Architecture:  architectureTmpl,
	TechInterview: techInterviewTmpl,
	Generic:       genericTmpl,
}

const refinementTmpl = `You are summarizing a BACKLOG REFINEMENT meeting transcript for a software engineering team.

{{.LangInstruction}}

Extract and structure the summary as:

## Stories discussed
For each: title, ticket ID if mentioned, scope changes, any reshaping.

## Acceptance criteria
Additions, removals, modifications. Quote exact wording when it was debated.

## Estimation outcomes
Story points or sizing agreed. If there was disagreement, note who held which position and why.

## Open questions
Anything blocking estimation or commitment. Include who is supposed to resolve it.

## Action items
- [ ] <action> — owner: <name>, due: <date if stated, else "TBD">

## Technical concerns
Architecture, dependencies, debt, risk, security. Be specific.

## Rules
- Specificity over brevity. "Concerns about auth" is wrong; "concerns about token refresh leaking across tenants in the multi-tenant migration" is right.
- Attribute decisions to named speakers when possible.
- Skip small talk, meta-commentary, and off-topic threads.
- Preserve ticket IDs, code names, and technical terms verbatim — do not translate them.
- If a story was discussed but not resolved, list it under Open questions, not Stories discussed.

Participants: {{.Participants}}

Transcript:
---
{{.Transcript}}
---
`

const planningTmpl = `You are summarizing a SPRINT PLANNING meeting transcript for a software engineering team.

{{.LangInstruction}}

Extract and structure the summary as:

## Sprint goal
The single overarching goal agreed for this sprint, verbatim if stated.

## Committed stories
For each: title, ticket ID, story points, owner if assigned.

## Capacity vs commitment
Available capacity, committed points, delta. Note any explicit over- or under-commit decisions and their rationale.

## Cross-team dependencies
External teams, blockers needing coordination, who owns the conversation.

## Risks
Specific risks raised during planning and any mitigation discussed.

## Stretch items
Stories pulled in only if capacity allows. Mark clearly as stretch, not committed.

## Action items
- [ ] <action> — owner: <name>, due: <date if stated, else "TBD">

## Rules
- Preserve ticket IDs, code names, and technical terms verbatim — do not translate them.
- Attribute commitments to named owners when possible.
- Skip small talk, meta-commentary, and off-topic threads.
- A story is "committed" only if the team explicitly agreed to it; otherwise list under Stretch or Open questions.

Participants: {{.Participants}}

Transcript:
---
{{.Transcript}}
---
`

const retroTmpl = `You are summarizing a SPRINT RETROSPECTIVE meeting transcript for a software engineering team.

{{.LangInstruction}}

Extract and structure the summary as:

## What went well
Concrete wins. Cite who raised each point when possible.

## What didn't
Concrete problems. Cite who raised each point when possible.

## Root causes
For the top issues, go past symptoms. Why did this happen? What underlying process, system, or communication issue caused it?

## Action items
- [ ] <action> — owner: <name>, due: <date if stated, else "TBD">

## Recurring themes
Patterns the team noted from prior retros, if mentioned.

## Rules
- Root causes are not symptoms. "Deployments were slow" is a symptom; "CI step X retries 5x on flake before failing" is a root cause.
- Attribute observations to named speakers when possible.
- Skip small talk, meta-commentary, and off-topic threads.
- Preserve ticket IDs, code names, and technical terms verbatim — do not translate them.

Participants: {{.Participants}}

Transcript:
---
{{.Transcript}}
---
`

const oneOnOneTmpl = `You are summarizing a 1-ON-1 meeting transcript between a manager and a direct report.

{{.LangInstruction}}

Extract and structure the summary as:

## Topics discussed
Brief list of the main topics that came up.

## Decisions
Anything explicitly agreed or decided during the conversation.

## Feedback exchanged
Both directions. Manager → report, report → manager. Quote specific wording when significant.

## Career and growth
Threads about development, promotion, scope expansion, skill gaps.

## Follow-ups for next 1:1
What to revisit next time.

## Signals worth attention
Disengagement, frustration, burnout, or other indicators that need follow-up. Be honest but careful with framing.

## Rules
- Treat this content as sensitive. Preserve nuance; do not flatten emotional content into corporate language.
- Attribute statements to the right speaker (manager vs report).
- Preserve specific examples and names raised in the conversation.
- Skip small talk and off-topic threads.

Participants: {{.Participants}}

Transcript:
---
{{.Transcript}}
---
`

const architectureTmpl = `You are summarizing an ARCHITECTURE / TECHNICAL DESIGN meeting transcript for a software engineering team.

{{.LangInstruction}}

Extract and structure the summary as:

## Problem statement
What is being solved, in technical terms. Cite stated constraints.

## Options considered
For each option: brief description, trade-offs raised (pros/cons), who advocated for it.

## Decision
What was chosen. If no decision, say so explicitly.

## Rationale
Why this option won over the others.

## Dissenting views
Positions held by people who disagreed with the chosen direction, and the reasons.

## ADR-worthy items
Items that should become Architecture Decision Records — call them out distinctly.

## Follow-up spikes / POCs
Investigations needed before commitment, with owners if assigned.

## Action items
- [ ] <action> — owner: <name>, due: <date if stated, else "TBD">

## Rules
- Preserve technical terms, library/framework names, and protocols verbatim — do not translate them.
- Attribute opinions to named speakers when possible.
- Distinguish "decided" from "leaning toward". Do not overstate consensus.
- Skip small talk, meta-commentary, and off-topic threads.

Participants: {{.Participants}}

Transcript:
---
{{.Transcript}}
---
`

const techInterviewTmpl = `You are summarizing a TECHNICAL INTERVIEW transcript for a software engineering hiring loop.

{{.LangInstruction}}

Extract and structure the summary as:

## Candidate and role
Candidate name (if stated), role being interviewed for, interview stage (screen, on-site, system design, etc.).

## Topics covered
Questions asked or problems posed. For each: brief description and the area it probed (DSA, system design, language depth, debugging, architecture, behavioral, etc.).

## Solutions and approaches
For each problem: the candidate's approach, key insights, dead ends, the final solution. Quote pseudocode or specific algorithmic choices when significant. Note time/space complexity if discussed.

## Technical strengths demonstrated
Specific, evidence-backed. "Strong on hash-map intuition — picked O(n) two-pass on problem 2 without prompting." Not "good at algorithms".

## Technical gaps observed
Specific, evidence-backed. "Did not consider race condition on shared counter even after a leading question about concurrency." Not "weak on concurrency".

## Communication and collaboration
How the candidate framed thinking out loud, responded to hints, handled disagreement, asked clarifying questions, negotiated trade-offs.

## Problem-solving process
Did they restate the problem, enumerate cases, test their solution, iterate? Walking style: top-down, bottom-up, test-first?

## Signals worth attention
Red flags, surprising positives, ambiguities. Be specific and avoid editorializing.

## Hire recommendation
One of: Strong hire / Hire / Lean hire / Lean no-hire / No hire / Strong no-hire — only if explicitly stated by the interviewer in the transcript. Otherwise: "Not stated".

## Rationale
Why this level. Cite the specific moments in the interview that drove the call.

## Follow-ups for next round
What to probe further if the candidate advances. Areas not covered, doubts to resolve.

## Rules
- Quote the candidate verbatim for technically meaningful statements; do not paraphrase code or complexity claims.
- Distinguish "candidate said" from "interviewer led them to". Hint-driven answers are not the same as independent ones.
- Preserve algorithm names, data structure names, language/framework names, and Big-O notation verbatim — do not translate them.
- Skip small talk, meta-commentary, and off-topic threads unless they reveal a signal.
- Be honest. A summary that softens gaps to be polite is a worse summary.

Participants: {{.Participants}}

Transcript:
---
{{.Transcript}}
---
`

const genericTmpl = `You are summarizing a meeting transcript for a software engineering team.

{{.LangInstruction}}

Extract and structure the summary as:

## Agenda items
The main topics that came up.

## Decisions
Anything explicitly agreed or decided.

## Action items
- [ ] <action> — owner: <name>, due: <date if stated, else "TBD">

## Open questions
Anything unresolved, and who is expected to resolve it.

## Rules
- Specificity over brevity.
- Attribute decisions to named speakers when possible.
- Preserve ticket IDs, code names, and technical terms verbatim — do not translate them.
- Skip small talk, meta-commentary, and off-topic threads.

Participants: {{.Participants}}

Transcript:
---
{{.Transcript}}
---
`
