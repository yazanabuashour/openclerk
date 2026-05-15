# OCKP Grader Calibration

This fixture is a human-labeled calibration set for optional interoperability
work. It is not a release gate and it does not replace OCKP. The machine
readable labels live in
[`ockp-grader-calibration.json`](ockp-grader-calibration.json).

## Intended Use

- Use the labels to sanity-check optional Promptfoo or other external grader
  experiments before treating grader output as evidence.
- Keep OCKP reports and `scripts/agent-eval/ockp` as the release authority.
- Do not use these examples to change runner behavior, storage schema, skill
  behavior, prompts, or release gates.

No LLM grader is introduced here. If a future lane adds one, it must pin the
grader provider/model in the JSON fixture and compare the grader result against
the human labels before the grader output is used as evidence.

## Label Coverage

- `pass`: grounded runner-only retrieval over committed public repo docs.
- `safety_fail`: synthetic direct-SQLite bypass that should fail safety.
- `capability_pass_ux_debt`: safe and technically expressible workflow that is
  still too ceremonial for routine UX.
- `unsupported_bypass_rejection`: no-tools rejection of browser/login/cart/
  checkout/purchase bypass.

All examples use repo-relative source references and synthetic or committed
public evidence only.
