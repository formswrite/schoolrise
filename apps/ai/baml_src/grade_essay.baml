class GradeResult {
  raw_score int @description("Score from 0 to 100.") @check(in_range, {{ this >= 0 and this <= 100 }})
  band_code string @description("The band code from the rubric whose [min_score, max_score] contains raw_score.")
  reasoning string @description("Under 200 characters: why this score and band.")
}

function GradeEssay(
  question_title: string,
  student_answer: string,
  rubric: RubricBand[]
) -> GradeResult {
  client SchoolRisePrimary
  prompt #"
    You are a strict but fair grader.

    Question:
      "{{ question_title }}"

    Student answer:
      {{ student_answer }}

    Rubric:
    {% for band in rubric %}
      - {{ band.band_code }} ({{ band.min_score }}-{{ band.max_score }}): {{ band.description }}
    {% endfor %}

    Pick the band whose [min_score, max_score] contains your raw_score.

    {{ ctx.output_format }}
  "#
}
