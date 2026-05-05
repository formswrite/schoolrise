client<llm> SchoolRiseGPT5Mini {
  provider openai-responses
  retry_policy Exponential
  options {
    model "gpt-5-mini"
    api_key env.OPENAI_API_KEY
  }
}

client<llm> SchoolRiseHaiku {
  provider anthropic
  retry_policy Exponential
  options {
    model "claude-3-5-haiku-20241022"
    api_key env.ANTHROPIC_API_KEY
  }
}

client<llm> SchoolRisePrimary {
  provider fallback
  options {
    strategy [SchoolRiseGPT5Mini, SchoolRiseHaiku]
  }
}

retry_policy Exponential {
  max_retries 2
  strategy {
    type exponential_backoff
    delay_ms 300
    multiplier 1.5
    max_delay_ms 10000
  }
}
