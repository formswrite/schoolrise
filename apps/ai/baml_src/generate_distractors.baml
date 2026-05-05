class DistractorList {
  distractors string[]
}

function GenerateDistractors(
  question_title: string,
  correct_answer: string,
  count: int
) -> DistractorList {
  client SchoolRisePrimary
  prompt #"
    You are a test designer.

    Question: "{{ question_title }}"
    Correct answer: "{{ correct_answer }}"

    Generate exactly {{ count }} wrong-but-plausible distractors. They must be:
      - similar in length and form to the correct answer
      - genuinely incorrect (no synonyms, no near-equivalents)
      - distinct from each other

    Do NOT include the correct answer in your list.

    {{ ctx.output_format }}
  "#
}

test capital_of_guinea {
  functions [GenerateDistractors]
  args {
    question_title "What is the capital of Guinea?"
    correct_answer "Conakry"
    count 3
  }
}
