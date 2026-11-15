package judge

import (
	"fmt"
	"strings"
)

const defaultRubric = `# Standard Pairwise Evaluation Rubric

Evaluate which response is superior for enterprise deployment:
1. Accuracy and Policy Adherence: Does the answer follow company policies and factual requirements?
2. Correctness and Completeness: Does the response address all aspects of the user's prompt without omissions?
3. Clarity and Professional Tone: Is the answer concise, well-structured, and free of unnecessary filler?
4. Constraint Satisfaction: Were all explicit formatting or negative constraints observed?
`

// BuildJudgePrompt assembles the prompt presenting Option A and Option B to the evaluator.
func BuildJudgePrompt(userPrompt, responseA, responseB string) string {
	return fmt.Sprintf(`[USER PROMPT]
%s

[RESPONSE A]
%s

[RESPONSE B]
%s

[EVALUATION TASK]
Carefully evaluate Response A and Response B against the rubric.
Determine whether Response A is better, Response B is better, or both are of equivalent quality (tie).

You MUST respond strictly with a valid JSON object matching this schema:
{
  "winner": "A" | "B" | "tie",
  "reasoning": "A concise explanation comparing both responses."
}
`, userPrompt, responseA, responseB)
}

// BuildJudgeSystemPrompt creates the system instruction incorporating the active rubric.
func BuildJudgeSystemPrompt(customRubric string) string {
	rubric := customRubric
	if strings.TrimSpace(rubric) == "" {
		rubric = defaultRubric
	}

	return fmt.Sprintf(`You are an expert, impartial AI evaluation judge assessing foundation model completions for enterprise deployment.
Your evaluation must be objective, analytical, and strictly grounded in the criteria below. Do not favor length or stylistic flourish over substance.

%s

Always output strictly valid JSON with keys "winner" and "reasoning". Do not include extra conversational text.`, rubric)
}
