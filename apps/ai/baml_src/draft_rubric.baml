class RubricBand {
  band_code string @description("Must match one of the band codes provided in the input.")
  description string @description("One sentence describing what the student must demonstrate.")
  min_score int @description("Inclusive lower bound, 0-100.")
  max_score int @description("Inclusive upper bound, 0-100. Bands must not overlap and together cover 0-100.")
}

class RubricBandList {
  bands RubricBand[]
}

function DraftRubric(
  question_title: string,
  scale_code: string,
  band_codes: string[]
) -> RubricBandList {
  client SchoolRisePrimary
  prompt #"
    You are an expert assessment designer.

    Draft a rubric for an ESSAY question on scale {{ scale_code }}:
      "{{ question_title }}"

    Use these band codes in order from lowest to highest performance:
    {% for code in band_codes %}
      - {{ code }}
    {% endfor %}

    For each band, write a 1-sentence description of what the student must
    demonstrate, plus min_score / max_score that together cover 0-100 with
    no overlaps and no gaps.

    {{ ctx.output_format }}
  "#
}

test french_5band {
  functions [DraftRubric]
  args {
    question_title "Describe your favourite story."
    scale_code "french_5level"
    band_codes ["debutant", "lettres", "mots", "paragraphe", "histoire"]
  }
}
