enum QuestionType {
  SHORT_ANSWER
  PARAGRAPH
  MULTIPLE_CHOICE
  YES_NO
  LINEAR_SCALE
}

class SuggestedItem {
  type QuestionType
  title string @description("Concise question prompt, under 120 characters.")
  required bool
  options string[]? @description("4 options for MULTIPLE_CHOICE; null/empty otherwise.")
}

class SuggestedItemList {
  items SuggestedItem[]
}

function SuggestItems(
  topic: string,
  scale_code: string,
  niveau_label: string,
  count: int
) -> SuggestedItemList {
  client SchoolRisePrimary
  prompt #"
    You are an expert curriculum designer for a national assessment program.

    Generate {{ count }} assessment questions for:
      - Topic: "{{ topic }}"
      - Scale: {{ scale_code }}
      - Niveau (grade level): {{ niveau_label }}

    Mix question types. For MULTIPLE_CHOICE, include exactly 4 plausible options
    with one obviously correct. Keep titles short (under 120 chars).

    {{ ctx.output_format }}
  "#
}

test ce1_arithmetic {
  functions [SuggestItems]
  args {
    topic "Basic arithmetic"
    scale_code "maths_5level"
    niveau_label "CE1"
    count 3
  }
}
